//go:generate mockery -name=EC2Client
package provider

import (
	"encoding/base64"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/NetSys/quilt/cluster/provider/mocks"
	"github.com/NetSys/quilt/db"
)

const testNamespace = "namespace"

func TestList(t *testing.T) {
	t.Parallel()

	mockClient := new(mocks.EC2Client)
	instances := []*ec2.Instance{
		// A booted spot instance (with a matching spot tag).
		{
			InstanceId:            aws.String("inst1"),
			SpotInstanceRequestId: aws.String("spot1"),
			PublicIpAddress:       aws.String("publicIP"),
			PrivateIpAddress:      aws.String("privateIP"),
			InstanceType:          aws.String("size"),
			State: &ec2.InstanceState{
				Name: aws.String(ec2.InstanceStateNameRunning),
			},
		},
		// A booted spot instance (with a lost spot tag).
		{
			InstanceId:            aws.String("inst2"),
			SpotInstanceRequestId: aws.String("spot2"),
			InstanceType:          aws.String("size2"),
			State: &ec2.InstanceState{
				Name: aws.String(ec2.InstanceStateNameRunning),
			},
		},
	}
	mockClient.On("DescribeInstances", mock.Anything).Return(
		&ec2.DescribeInstancesOutput{
			Reservations: []*ec2.Reservation{
				{
					Instances: instances,
				},
			},
		}, nil,
	)
	mockClient.On("DescribeSpotInstanceRequests", mock.Anything).Return(
		&ec2.DescribeSpotInstanceRequestsOutput{
			SpotInstanceRequests: []*ec2.SpotInstanceRequest{
				// A spot request with tags and a corresponding instance.
				{
					SpotInstanceRequestId: aws.String("spot1"),
					State: aws.String(ec2.SpotInstanceStateActive),
					Tags: []*ec2.Tag{
						{
							Key:   aws.String(testNamespace),
							Value: aws.String(""),
						},
					},
					InstanceId: aws.String("inst1"),
				},
				// A spot request without tags, but with
				// a corresponding instance.
				{
					SpotInstanceRequestId: aws.String("spot2"),
					State: aws.String(
						ec2.SpotInstanceStateActive),
					InstanceId: aws.String("inst2"),
				},
				// A spot request that hasn't been booted yet.
				{
					SpotInstanceRequestId: aws.String("spot3"),
					State: aws.String(ec2.SpotInstanceStateOpen),
					Tags: []*ec2.Tag{
						{
							Key:   aws.String(testNamespace),
							Value: aws.String(""),
						},
					},
				},
				// A spot request in another namespace.
				{
					SpotInstanceRequestId: aws.String("spot4"),
					State: aws.String(ec2.SpotInstanceStateOpen),
					Tags: []*ec2.Tag{
						{
							Key:   aws.String("notOurs"),
							Value: aws.String(""),
						},
					},
				},
			},
		}, nil,
	)

	emptyClient := new(mocks.EC2Client)
	emptyClient.On("DescribeInstances", mock.Anything).Return(
		&ec2.DescribeInstancesOutput{}, nil,
	)
	emptyClient.On("DescribeSpotInstanceRequests", mock.Anything).Return(
		&ec2.DescribeSpotInstanceRequestsOutput{}, nil,
	)

	amazonCluster := newAmazonCluster(func(region string) EC2Client {
		if region == "us-west-1" {
			return mockClient
		}
		return emptyClient
	})

	amazonCluster.namespace = testNamespace
	spots, err := amazonCluster.List()

	assert.Nil(t, err)
	assert.Equal(t, []Machine{
		{
			ID:        "spot1",
			Provider:  db.Amazon,
			PublicIP:  "publicIP",
			PrivateIP: "privateIP",
			Size:      "size",
			Region:    "us-west-1",
		},
		{
			ID:       "spot2",
			Provider: db.Amazon,
			Region:   "us-west-1",
			Size:     "size2",
		},
		{
			ID:       "spot3",
			Provider: db.Amazon,
			Region:   "us-west-1",
		},
	}, spots)
}

func TestNewACLs(t *testing.T) {
	t.Parallel()

	mockClient := new(mocks.EC2Client)
	mockClient.On("DescribeSecurityGroups", mock.Anything).Return(
		&ec2.DescribeSecurityGroupsOutput{
			SecurityGroups: []*ec2.SecurityGroup{
				{
					IpPermissions: []*ec2.IpPermission{
						{
							IpRanges: []*ec2.IpRange{
								{CidrIp: aws.String(
									"foo")},
								{CidrIp: aws.String(
									"deleteMe")},
							},
							IpProtocol: aws.String("-1"),
						},
					},
					GroupId: aws.String(""),
				},
			},
		}, nil,
	)
	mockClient.On("RevokeSecurityGroupIngress", mock.Anything).Return(
		&ec2.RevokeSecurityGroupIngressOutput{}, nil,
	)
	mockClient.On("AuthorizeSecurityGroupIngress", mock.Anything).Return(
		&ec2.AuthorizeSecurityGroupIngressOutput{}, nil,
	)

	cluster := newAmazonCluster(func(region string) EC2Client {
		return mockClient
	})
	cluster.namespace = testNamespace

	err := cluster.SetACLs([]string{"foo", "bar"})

	assert.Nil(t, err)

	mockClient.AssertCalled(t, "AuthorizeSecurityGroupIngress",
		&ec2.AuthorizeSecurityGroupIngressInput{
			GroupName:               aws.String(testNamespace),
			SourceSecurityGroupName: aws.String(testNamespace),
		},
	)

	mockClient.AssertCalled(t, "AuthorizeSecurityGroupIngress",
		&ec2.AuthorizeSecurityGroupIngressInput{
			GroupName: aws.String(testNamespace),
			IpPermissions: []*ec2.IpPermission{
				{
					IpRanges: []*ec2.IpRange{
						{
							CidrIp: aws.String("bar"),
						},
					},
					IpProtocol: aws.String("-1"),
				},
			},
		},
	)

	mockClient.AssertCalled(t, "RevokeSecurityGroupIngress",
		&ec2.RevokeSecurityGroupIngressInput{
			GroupName: aws.String(testNamespace),
			IpPermissions: []*ec2.IpPermission{
				{
					IpRanges: []*ec2.IpRange{
						{
							CidrIp: aws.String("deleteMe"),
						},
					},
					IpProtocol: aws.String("-1"),
				},
			},
		},
	)
}

func TestBoot(t *testing.T) {
	t.Parallel()

	mockClient := new(mocks.EC2Client)
	mockClient.On("DescribeSecurityGroups", mock.Anything).Return(
		&ec2.DescribeSecurityGroupsOutput{
			SecurityGroups: []*ec2.SecurityGroup{
				{
					GroupId: aws.String("groupId"),
				},
			},
		}, nil,
	)
	mockClient.On("RequestSpotInstances", mock.Anything).Return(
		&ec2.RequestSpotInstancesOutput{
			SpotInstanceRequests: []*ec2.SpotInstanceRequest{
				{
					SpotInstanceRequestId: aws.String("spot1"),
				},
				{
					SpotInstanceRequestId: aws.String("spot2"),
				},
			},
		}, nil,
	)
	mockClient.On("CreateTags", mock.Anything).Return(
		&ec2.CreateTagsOutput{}, nil,
	)
	mockClient.On("DescribeInstances", mock.Anything).Return(
		&ec2.DescribeInstancesOutput{}, nil,
	)
	mockClient.On("DescribeSpotInstanceRequests", mock.Anything).Return(
		&ec2.DescribeSpotInstanceRequestsOutput{
			SpotInstanceRequests: []*ec2.SpotInstanceRequest{
				{
					SpotInstanceRequestId: aws.String("spot1"),
					State: aws.String(ec2.SpotInstanceStateActive),
					Tags: []*ec2.Tag{
						{
							Key:   aws.String(testNamespace),
							Value: aws.String(""),
						},
					},
				},
				{
					SpotInstanceRequestId: aws.String("spot2"),
					State: aws.String(ec2.SpotInstanceStateActive),
					Tags: []*ec2.Tag{
						{
							Key:   aws.String(testNamespace),
							Value: aws.String(""),
						},
					},
				},
			},
		}, nil,
	)

	amazonCluster := newAmazonCluster(func(region string) EC2Client {
		return mockClient
	})
	amazonCluster.namespace = testNamespace

	err := amazonCluster.Boot([]Machine{
		{
			Region:   "us-west-1",
			Size:     "m4.large",
			DiskSize: 32,
		},
		{
			Region:   "us-west-1",
			Size:     "m4.large",
			DiskSize: 32,
		},
	})
	assert.Nil(t, err)

	cfg := cloudConfigUbuntu(nil, "xenial")
	mockClient.AssertCalled(t, "RequestSpotInstances",
		&ec2.RequestSpotInstancesInput{
			SpotPrice: aws.String(spotPrice),
			LaunchSpecification: &ec2.RequestSpotLaunchSpecification{
				ImageId:      aws.String(amis["us-west-1"]),
				InstanceType: aws.String("m4.large"),
				UserData: aws.String(base64.StdEncoding.EncodeToString(
					[]byte(cfg))),
				SecurityGroupIds: aws.StringSlice([]string{"groupId"}),
				BlockDeviceMappings: []*ec2.BlockDeviceMapping{
					blockDevice(32)},
			},
			InstanceCount: aws.Int64(2),
		},
	)
	mockClient.AssertCalled(t, "CreateTags",
		&ec2.CreateTagsInput{
			Tags: []*ec2.Tag{
				{
					Key:   aws.String(testNamespace),
					Value: aws.String(""),
				},
			},
			Resources: aws.StringSlice([]string{"spot1", "spot2"}),
		},
	)
}

func TestStop(t *testing.T) {
	t.Parallel()

	mockClient := new(mocks.EC2Client)
	toStopIDs := []string{"spot1", "spot2"}
	// When we're getting information about what machines to stop.
	mockClient.On("DescribeSpotInstanceRequests",
		&ec2.DescribeSpotInstanceRequestsInput{
			SpotInstanceRequestIds: aws.StringSlice(toStopIDs),
		}).Return(
		&ec2.DescribeSpotInstanceRequestsOutput{
			SpotInstanceRequests: []*ec2.SpotInstanceRequest{
				{
					SpotInstanceRequestId: aws.String(toStopIDs[0]),
					InstanceId:            aws.String("inst1"),
					State: aws.String(
						ec2.SpotInstanceStateActive),
				},
				{
					SpotInstanceRequestId: aws.String(toStopIDs[1]),
					State: aws.String(ec2.SpotInstanceStateActive),
				},
			},
		}, nil,
	)
	// When we're listing machines to tell if they've stopped.
	mockClient.On("DescribeSpotInstanceRequests", mock.Anything).Return(
		&ec2.DescribeSpotInstanceRequestsOutput{}, nil,
	)
	mockClient.On("TerminateInstances", mock.Anything).Return(
		&ec2.TerminateInstancesOutput{}, nil,
	)
	mockClient.On("CancelSpotInstanceRequests", mock.Anything).Return(
		&ec2.CancelSpotInstanceRequestsOutput{}, nil,
	)
	mockClient.On("DescribeInstances", mock.Anything).Return(
		&ec2.DescribeInstancesOutput{}, nil,
	)

	amazonCluster := newAmazonCluster(func(region string) EC2Client {
		return mockClient
	})

	err := amazonCluster.Stop([]Machine{
		{
			Region: "us-west-1",
			ID:     toStopIDs[0],
		},
		{
			Region: "us-west-1",
			ID:     toStopIDs[1],
		},
	})
	assert.Nil(t, err)

	mockClient.AssertCalled(t, "TerminateInstances",
		&ec2.TerminateInstancesInput{
			InstanceIds: aws.StringSlice([]string{"inst1"}),
		},
	)

	mockClient.AssertCalled(t, "CancelSpotInstanceRequests",
		&ec2.CancelSpotInstanceRequestsInput{
			SpotInstanceRequestIds: aws.StringSlice(toStopIDs),
		},
	)
}
