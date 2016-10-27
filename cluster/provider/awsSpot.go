package provider

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/NetSys/quilt/constants"
	"github.com/NetSys/quilt/db"
	"github.com/NetSys/quilt/join"
	"github.com/NetSys/quilt/stitch"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"

	log "github.com/Sirupsen/logrus"
)

// EC2Client defines an interface that can be mocked out for interacting with EC2.
type EC2Client interface {
	AuthorizeSecurityGroupIngress(*ec2.AuthorizeSecurityGroupIngressInput) (
		*ec2.AuthorizeSecurityGroupIngressOutput, error)

	CancelSpotInstanceRequests(*ec2.CancelSpotInstanceRequestsInput) (
		*ec2.CancelSpotInstanceRequestsOutput, error)

	CreateSecurityGroup(*ec2.CreateSecurityGroupInput) (
		*ec2.CreateSecurityGroupOutput, error)

	CreateTags(*ec2.CreateTagsInput) (*ec2.CreateTagsOutput, error)

	DescribeSecurityGroups(*ec2.DescribeSecurityGroupsInput) (
		*ec2.DescribeSecurityGroupsOutput, error)

	DescribeInstances(*ec2.DescribeInstancesInput) (
		*ec2.DescribeInstancesOutput, error)

	DescribeSpotInstanceRequests(*ec2.DescribeSpotInstanceRequestsInput) (
		*ec2.DescribeSpotInstanceRequestsOutput, error)

	DescribeVolumes(*ec2.DescribeVolumesInput) (
		*ec2.DescribeVolumesOutput, error)

	RevokeSecurityGroupIngress(*ec2.RevokeSecurityGroupIngressInput) (
		*ec2.RevokeSecurityGroupIngressOutput, error)

	TerminateInstances(*ec2.TerminateInstancesInput) (
		*ec2.TerminateInstancesOutput, error)

	RequestSpotInstances(*ec2.RequestSpotInstancesInput) (
		*ec2.RequestSpotInstancesOutput, error)
}

const spotPrice = "0.5"

// Ubuntu 16.04, 64-bit hvm-ssd
var amis = map[string]string{
	"ap-southeast-2": "ami-550c3c36",
	"us-west-1":      "ami-26074946",
	"us-west-2":      "ami-e1fe2281",
}

func newAmazonCluster(sessionGetter func(string) EC2Client) *amazonCluster {
	return &amazonCluster{
		sessions:      make(map[string]EC2Client),
		sessionGetter: sessionGetter,
	}
}

func newEC2Session(region string) EC2Client {
	session := session.New()
	session.Config.Region = aws.String(region)
	return ec2.New(session)
}

// blockDevice returns the block device we use for our AWS machines.
func blockDevice(diskSize int) *ec2.BlockDeviceMapping {
	return &ec2.BlockDeviceMapping{
		DeviceName: aws.String("/dev/sda1"),
		Ebs: &ec2.EbsBlockDevice{
			DeleteOnTermination: aws.Bool(true),
			VolumeSize:          aws.Int64(int64(diskSize)),
			VolumeType:          aws.String("gp2"),
		},
	}
}

type amazonCluster struct {
	sessionGetter func(string) EC2Client
	sessions      map[string]EC2Client

	namespace string
}

type awsID struct {
	spotID string
	region string
}

func getSpotIDs(ids []awsID) []string {
	var spotIDs []string
	for _, id := range ids {
		spotIDs = append(spotIDs, id.spotID)
	}

	return spotIDs
}

func groupByRegion(ids []awsID) map[string][]awsID {
	grouped := make(map[string][]awsID)
	for _, id := range ids {
		region := id.region
		if _, ok := grouped[region]; !ok {
			grouped[region] = []awsID{}
		}
		grouped[region] = append(grouped[region], id)
	}

	return grouped
}

func (clst *amazonCluster) Connect(namespace string) error {
	clst.namespace = strings.ToLower(namespace)

	if _, err := clst.List(); err != nil {
		return errors.New("AWS failed to connect")
	}
	return nil
}

func (clst amazonCluster) getSession(region string) EC2Client {
	if _, ok := clst.sessions[region]; !ok {
		clst.sessions[region] = clst.sessionGetter(region)
	}

	return clst.sessions[region]
}

func (clst amazonCluster) Boot(bootSet []Machine) error {
	if len(bootSet) <= 0 {
		return nil
	}

	type bootReq struct {
		cfg      string
		size     string
		region   string
		diskSize int
	}

	bootReqMap := make(map[bootReq]int64) // From boot request to an instance count.
	for _, m := range bootSet {
		br := bootReq{
			cfg:      cloudConfigUbuntu(m.SSHKeys, "xenial"),
			size:     m.Size,
			region:   m.Region,
			diskSize: m.DiskSize,
		}
		bootReqMap[br] = bootReqMap[br] + 1
	}

	var awsIDs []awsID
	for br, count := range bootReqMap {
		session := clst.getSession(br.region)
		groupID, _, err := clst.GetCreateSecurityGroup(session)
		if err != nil {
			return err
		}

		cloudConfig64 := base64.StdEncoding.EncodeToString([]byte(br.cfg))
		resp, err := session.RequestSpotInstances(&ec2.RequestSpotInstancesInput{
			SpotPrice: aws.String(spotPrice),
			LaunchSpecification: &ec2.RequestSpotLaunchSpecification{
				ImageId:          aws.String(amis[br.region]),
				InstanceType:     aws.String(br.size),
				UserData:         &cloudConfig64,
				SecurityGroupIds: []*string{aws.String(groupID)},
				BlockDeviceMappings: []*ec2.BlockDeviceMapping{
					blockDevice(br.diskSize),
				},
			},
			InstanceCount: &count,
		})

		if err != nil {
			return err
		}

		for _, request := range resp.SpotInstanceRequests {
			awsIDs = append(awsIDs, awsID{
				spotID: *request.SpotInstanceRequestId,
				region: br.region})
		}
	}

	if err := clst.tagSpotRequests(awsIDs); err != nil {
		return err
	}

	return clst.wait(awsIDs, true)
}

func (clst amazonCluster) Stop(machines []Machine) error {
	var awsIDs []awsID
	for _, m := range machines {
		awsIDs = append(awsIDs, awsID{
			region: m.Region,
			spotID: m.ID,
		})
	}
	for region, ids := range groupByRegion(awsIDs) {
		session := clst.getSession(region)
		spotIDs := getSpotIDs(ids)

		spots, err := session.DescribeSpotInstanceRequests(
			&ec2.DescribeSpotInstanceRequestsInput{
				SpotInstanceRequestIds: aws.StringSlice(spotIDs),
			})
		if err != nil {
			return err
		}

		instIds := []string{}
		for _, spot := range spots.SpotInstanceRequests {
			if spot.InstanceId != nil {
				instIds = append(instIds, *spot.InstanceId)
			}
		}

		if len(instIds) > 0 {
			_, err = session.TerminateInstances(&ec2.TerminateInstancesInput{
				InstanceIds: aws.StringSlice(instIds),
			})
			if err != nil {
				return err
			}
		}

		_, err = session.CancelSpotInstanceRequests(
			&ec2.CancelSpotInstanceRequestsInput{
				SpotInstanceRequestIds: aws.StringSlice(spotIDs),
			})
		if err != nil {
			return err
		}

		if err := clst.wait(ids, false); err != nil {
			return err
		}
	}

	return nil
}

func (clst amazonCluster) List() ([]Machine, error) {
	machines := []Machine{}
	for region := range amis {
		session := clst.getSession(region)

		spots, err := session.DescribeSpotInstanceRequests(nil)
		if err != nil {
			return nil, err
		}

		insts, err := session.DescribeInstances(&ec2.DescribeInstancesInput{
			Filters: []*ec2.Filter{
				{
					Name:   aws.String("instance.group-name"),
					Values: []*string{aws.String(clst.namespace)},
				},
			},
		})
		if err != nil {
			return nil, err
		}

		instMap := make(map[string]*ec2.Instance)
		for _, res := range insts.Reservations {
			for _, inst := range res.Instances {
				instMap[*inst.InstanceId] = inst
			}
		}

		for _, spot := range spots.SpotInstanceRequests {
			if *spot.State != ec2.SpotInstanceStateActive &&
				*spot.State != ec2.SpotInstanceStateOpen {
				continue
			}

			var inst *ec2.Instance
			if spot.InstanceId != nil {
				inst = instMap[*spot.InstanceId]
			}

			// Due to a race condition in the AWS API, it's possible that
			// spot requests might lose their Tags. If handled naively,
			// those spot requests would technically be without a namespace,
			// meaning the instances they create would be live forever as
			// zombies.
			//
			// To mitigate this issue, we rely not only on the spot request
			// tags, but additionally on the instance security group. If a
			// spot request has a running instance in the appropriate
			// security group, it is by definition in our namespace.
			// Thus, we only check the tags for spot requests without
			// running instances.
			if inst == nil {
				var isOurs bool
				for _, tag := range spot.Tags {
					ns := clst.namespace
					if tag != nil && tag.Key != nil &&
						*tag.Key == ns {
						isOurs = true
						break
					}
				}

				if !isOurs {
					continue
				}
			}

			machine := Machine{
				ID:       *spot.SpotInstanceRequestId,
				Region:   region,
				Provider: db.Amazon,
			}

			if inst != nil {
				if *inst.State.Name != ec2.InstanceStateNamePending &&
					*inst.State.Name != ec2.InstanceStateNameRunning {
					continue
				}

				if inst.PublicIpAddress != nil {
					machine.PublicIP = *inst.PublicIpAddress
				}

				if inst.PrivateIpAddress != nil {
					machine.PrivateIP = *inst.PrivateIpAddress
				}

				if inst.InstanceType != nil {
					machine.Size = *inst.InstanceType
				}

				if len(inst.BlockDeviceMappings) != 0 {
					volumeID := inst.BlockDeviceMappings[0].
						Ebs.VolumeId
					filters := []*ec2.Filter{
						{
							Name: aws.String("volume-id"),
							Values: []*string{
								aws.String(*volumeID),
							},
						},
					}

					volumeInfo, err := session.DescribeVolumes(
						&ec2.DescribeVolumesInput{
							Filters: filters,
						})
					if err != nil {
						return nil, err
					}
					if len(volumeInfo.Volumes) == 1 {
						machine.DiskSize = int(
							*volumeInfo.Volumes[0].Size)
					}
				}
			}

			machines = append(machines, machine)
		}
	}

	return machines, nil
}

func (clst *amazonCluster) ChooseSize(ram stitch.Range, cpu stitch.Range,
	maxPrice float64) string {
	return pickBestSize(constants.AwsDescriptions, ram, cpu, maxPrice)
}

func (clst *amazonCluster) tagSpotRequests(awsIDs []awsID) error {
OuterLoop:
	for region, ids := range groupByRegion(awsIDs) {
		session := clst.getSession(region)
		spotIDs := getSpotIDs(ids)

		var err error
		for i := 0; i < 30; i++ {
			_, err = session.CreateTags(&ec2.CreateTagsInput{
				Tags: []*ec2.Tag{
					{
						Key:   aws.String(clst.namespace),
						Value: aws.String(""),
					},
				},
				Resources: aws.StringSlice(spotIDs),
			})
			if err == nil {
				continue OuterLoop
			}
			time.Sleep(5 * time.Second)
		}

		log.Warn("Failed to tag spot requests: ", err)
		session.CancelSpotInstanceRequests(
			&ec2.CancelSpotInstanceRequestsInput{
				SpotInstanceRequestIds: aws.StringSlice(spotIDs),
			})

		return err
	}

	return nil
}

/* Wait for the spot request 'ids' to have booted or terminated depending on the value
 * of 'boot' */
func (clst *amazonCluster) wait(awsIDs []awsID, boot bool) error {
OuterLoop:
	for i := 0; i < 100; i++ {
		machines, err := clst.List()
		if err != nil {
			log.WithError(err).Warn("Failed to get machines.")
			time.Sleep(10 * time.Second)
			continue
		}

		exists := make(map[awsID]struct{})
		for _, inst := range machines {
			id := awsID{
				spotID: inst.ID,
				region: inst.Region,
			}

			exists[id] = struct{}{}
		}

		for _, id := range awsIDs {
			if _, ok := exists[id]; ok != boot {
				time.Sleep(10 * time.Second)
				continue OuterLoop
			}
		}

		return nil
	}

	return errors.New("timed out")
}

func (clst *amazonCluster) GetCreateSecurityGroup(session EC2Client) (
	string, []*ec2.IpPermission, error) {

	resp, err := session.DescribeSecurityGroups(
		&ec2.DescribeSecurityGroupsInput{
			Filters: []*ec2.Filter{
				{
					Name: aws.String("group-name"),
					Values: []*string{
						aws.String(clst.namespace),
					},
				},
			},
		})

	if err != nil {
		return "", nil, err
	}

	groups := resp.SecurityGroups
	if len(groups) > 1 {
		err := errors.New("Multiple Security Groups with the same name: " +
			clst.namespace)
		return "", nil, err
	}

	if len(groups) == 1 {
		return *groups[0].GroupId, groups[0].IpPermissions, nil
	}

	csgResp, err := session.CreateSecurityGroup(
		&ec2.CreateSecurityGroupInput{
			Description: aws.String("Quilt Group"),
			GroupName:   aws.String(clst.namespace),
		})
	if err != nil {
		return "", nil, err
	}

	return *csgResp.GroupId, nil, nil
}

func (clst *amazonCluster) SetACLs(acls []ACL) error {
	for region := range amis {
		session := clst.getSession(region)

		groupID, ingress, err := clst.GetCreateSecurityGroup(session)
		if err != nil {
			return err
		}

		rangesToAdd, foundGroup, rulesToRemove := syncACLs(acls, groupID, ingress)

		if len(rangesToAdd) != 0 {
			logACLs(true, rangesToAdd)
			_, err = session.AuthorizeSecurityGroupIngress(
				&ec2.AuthorizeSecurityGroupIngressInput{
					GroupName:     aws.String(clst.namespace),
					IpPermissions: rangesToAdd,
				},
			)
			if err != nil {
				return err
			}
		}

		if !foundGroup {
			log.WithField("Group", clst.namespace).Debug("Amazon: Add group")
			_, err = session.AuthorizeSecurityGroupIngress(
				&ec2.AuthorizeSecurityGroupIngressInput{
					GroupName: aws.String(
						clst.namespace),
					SourceSecurityGroupName: aws.String(
						clst.namespace),
				},
			)
			if err != nil {
				return err
			}
		}

		if len(rulesToRemove) != 0 {
			logACLs(false, rulesToRemove)
			_, err = session.RevokeSecurityGroupIngress(
				&ec2.RevokeSecurityGroupIngressInput{
					GroupName:     aws.String(clst.namespace),
					IpPermissions: rulesToRemove,
				},
			)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// syncACLs returns the permissions that need to be removed and added in order
// for the cloud ACLs to match the policy.
// rangesToAdd is guaranteed to always have exactly one item in the IpRanges slice.
func syncACLs(desiredACLs []ACL, desiredGroupID string,
	current []*ec2.IpPermission) (rangesToAdd []*ec2.IpPermission, foundGroup bool,
	toRemove []*ec2.IpPermission) {

	var currRangeRules []*ec2.IpPermission
	for _, perm := range current {
		for _, ipRange := range perm.IpRanges {
			currRangeRules = append(currRangeRules, &ec2.IpPermission{
				IpProtocol: perm.IpProtocol,
				FromPort:   perm.FromPort,
				ToPort:     perm.ToPort,
				IpRanges: []*ec2.IpRange{
					ipRange,
				},
			})
		}
		for _, pair := range perm.UserIdGroupPairs {
			if *pair.GroupId != desiredGroupID {
				toRemove = append(toRemove, &ec2.IpPermission{
					UserIdGroupPairs: []*ec2.UserIdGroupPair{
						pair,
					},
				})
			} else {
				foundGroup = true
			}
		}
	}

	var desiredRangeRules []*ec2.IpPermission
	for _, acl := range desiredACLs {
		desiredRangeRules = append(desiredRangeRules, &ec2.IpPermission{
			FromPort: aws.Int64(int64(acl.MinPort)),
			ToPort:   aws.Int64(int64(acl.MaxPort)),
			IpRanges: []*ec2.IpRange{
				{
					CidrIp: aws.String(acl.CidrIP),
				},
			},
			IpProtocol: aws.String("tcp"),
		}, &ec2.IpPermission{
			FromPort: aws.Int64(int64(acl.MinPort)),
			ToPort:   aws.Int64(int64(acl.MaxPort)),
			IpRanges: []*ec2.IpRange{
				{
					CidrIp: aws.String(acl.CidrIP),
				},
			},
			IpProtocol: aws.String("udp"),
		}, &ec2.IpPermission{
			FromPort: aws.Int64(-1),
			ToPort:   aws.Int64(-1),
			IpRanges: []*ec2.IpRange{
				{
					CidrIp: aws.String(acl.CidrIP),
				},
			},
			IpProtocol: aws.String("icmp"),
		})
	}

	_, toAdd, rangesToRemove := join.HashJoin(ipPermSlice(desiredRangeRules),
		ipPermSlice(currRangeRules), permToACLKey, permToACLKey)
	for _, intf := range toAdd {
		rangesToAdd = append(rangesToAdd, intf.(*ec2.IpPermission))
	}
	for _, intf := range rangesToRemove {
		toRemove = append(toRemove, intf.(*ec2.IpPermission))
	}

	return rangesToAdd, foundGroup, toRemove
}

func logACLs(add bool, perms []*ec2.IpPermission) {
	action := "Remove"
	if add {
		action = "Add"
	}

	for _, perm := range perms {
		if len(perm.IpRanges) != 0 {
			// Each rule has three variants (TCP, UDP, and ICMP), but
			// we only want to log once.
			protocol := *perm.IpProtocol
			if protocol != "tcp" {
				continue
			}

			cidrIP := *perm.IpRanges[0].CidrIp
			ports := fmt.Sprintf("%d", *perm.FromPort)
			if *perm.FromPort != *perm.ToPort {
				ports += fmt.Sprintf("-%d", *perm.ToPort)
			}
			log.WithField("ACL",
				fmt.Sprintf("%s:%s", cidrIP, ports)).
				Debugf("Amazon: %s ACL", action)
		} else {
			log.WithField("Group",
				*perm.UserIdGroupPairs[0].GroupName).
				Debugf("Amazon: %s group", action)
		}
	}
}

type ipPermissionKey struct {
	protocol string
	ipRange  string
	minPort  int
	maxPort  int
}

func permToACLKey(permIntf interface{}) interface{} {
	perm := permIntf.(*ec2.IpPermission)

	return ipPermissionKey{
		protocol: resolveString(perm.IpProtocol),
		ipRange:  resolveString(perm.IpRanges[0].CidrIp),
		minPort:  int(resolveInt64(perm.FromPort)),
		maxPort:  int(resolveInt64(perm.ToPort)),
	}
}

type ipPermSlice []*ec2.IpPermission

func (slc ipPermSlice) Get(ii int) interface{} {
	return slc[ii]
}

func (slc ipPermSlice) Len() int {
	return len(slc)
}

func (slc ipPermSlice) Less(i, j int) bool {
	return strings.Compare(slc[i].String(), slc[j].String()) < 0
}

func (slc ipPermSlice) Swap(i, j int) {
	slc[i], slc[j] = slc[j], slc[i]
}
