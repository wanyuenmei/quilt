package etcd

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"sort"
	"time"

	"github.com/NetSys/quilt/db"
	"github.com/NetSys/quilt/join"

	log "github.com/Sirupsen/logrus"
	"github.com/coreos/etcd/client"
)

const (
	minionDir       = "/minion"
	labelToIPStore  = minionDir + "/labelIP"
	containerStore  = minionDir + "/container"
	dockerToIPStore = minionDir + "/dockerIP"
)

// XXX: This really shouldn't live in here.  It's just a temporary measure, soon we'll
// disentangle the etcd logic from the IP allocation logic.  At that point, we can ditch
// this.
const gatewayIP = "10.0.0.1"

// We store rand.Uint32() in a variable so it's easily mocked out by the unit tests.
// Nondeterminism is hard to test.
var rand32 = rand.Uint32

// Keeping all the store data types in a struct makes it much less verbose to pass them
// around while operating on them
type storeData struct {
	containers []storeContainer
	dockerIPs  map[string]string
	multiHost  map[string]string
}

type storeContainer struct {
	DockerID string
	Command  []string
	Labels   []string
	Env      map[string]string
}

type storeContainerSlice []storeContainer

// wakeChan collapses the various channels these functions wait on into a single
// channel. Multiple redundant pings will be coalesced into a single message.
func wakeChan(conn db.Conn, store Store) chan struct{} {
	minionWatch := store.Watch(minionDir, 1*time.Second)
	trigg := conn.TriggerTick(30, db.MinionTable, db.ContainerTable, db.LabelTable,
		db.EtcdTable).C

	c := make(chan struct{}, 1)
	go func() {
		for {
			select {
			case <-minionWatch:
			case <-trigg:
			}

			select {
			case c <- struct{}{}:
			default: // There's a notification in queue, no need for another.
			}
		}
	}()

	return c
}

func runNetwork(conn db.Conn, store Store) {
	for range wakeChan(conn, store) {
		// If the etcd read failed, we only want to update the db if it
		// failed because a key was missing (has not been created yet).
		// In all other cases, we skip this iteration.
		etcdData, err := readEtcd(store)
		if err != nil {
			etcdErr, ok := err.(client.Error)
			if !ok || etcdErr.Code != client.ErrorCodeKeyNotFound {
				log.WithError(err).Error("Etcd transaction failed.")
				continue
			}
			log.WithError(err).Debug()
		}

		leader := false
		var containers []db.Container
		conn.Transact(func(view db.Database) error {
			leader = view.EtcdLeader()
			containers = view.SelectFromContainer(func(c db.Container) bool {
				return c.DockerID != ""
			})

			// It would likely be more efficient to perform the etcd write
			// outside of the DB transact. But, if we perform the writes
			// after the transact, there is no way to ensure that the writes
			// were successful before updating the DB with the information
			// produced by the updateEtcd* functions (not considering the
			// etcd writes they perform).
			if leader {
				etcdData, err = updateEtcd(store, etcdData, containers)
				if err != nil {
					log.WithError(err).Error("Etcd update failed.")
					return nil
				}
			}

			updateDBContainers(view, etcdData)
			updateDBLabels(view, etcdData)
			return nil
		})
	}
}

func readEtcd(store Store) (storeData, error) {
	containers, err := store.Get(containerStore)
	dockerIPs, err2 := store.Get(dockerToIPStore)
	if err2 != nil {
		err = err2
	}

	labels, err3 := store.Get(labelToIPStore)
	if err3 != nil {
		err = err3
	}

	etcdContainerSlice := []storeContainer{}
	dockerMap := map[string]string{}
	multiHostMap := map[string]string{}

	// Failed store reads will just be skipped by Unmarshal, which is fine
	// since an error is returned
	json.Unmarshal([]byte(containers), &etcdContainerSlice)
	json.Unmarshal([]byte(dockerIPs), &dockerMap)
	json.Unmarshal([]byte(labels), &multiHostMap)

	return storeData{etcdContainerSlice, dockerMap, multiHostMap}, err
}

func updateEtcd(s Store, etcdData storeData, containers []db.Container) (storeData,
	error) {

	if etcdData, err := updateEtcdContainer(s, etcdData, containers); err != nil {
		return etcdData, err
	}

	if etcdData, err := updateEtcdDocker(s, etcdData, containers); err != nil {
		return etcdData, err
	}

	if etcdData, err := updateEtcdLabel(s, etcdData, containers); err != nil {
		return etcdData, err
	}

	return etcdData, nil
}

func updateEtcdContainer(s Store, etcdData storeData, containers []db.Container) (
	storeData, error) {

	dbContainerSlice := []storeContainer{}
	for _, c := range containers {
		sc := storeContainer{
			DockerID: c.DockerID,
			Command:  c.Command,
			Labels:   c.Labels,
			Env:      c.Env,
		}
		dbContainerSlice = append(dbContainerSlice, sc)
	}
	sort.Sort(storeContainerSlice(dbContainerSlice))

	dbContainers, _ := json.Marshal(dbContainerSlice)
	jsonContainers, _ := json.Marshal(etcdData.containers)
	if string(dbContainers) == string(jsonContainers) {
		return etcdData, nil
	}

	if err := s.Set(containerStore, string(dbContainers)); err != nil {
		return etcdData, err
	}

	etcdData.containers = dbContainerSlice
	return etcdData, nil
}

func updateEtcdDocker(s Store, etcdData storeData, containers []db.Container) (storeData,
	error) {

	newDockerMap := map[string]string{}
	for _, c := range containers {
		newDockerMap[c.DockerID] = ""
	}

	// Etcd is the source of truth for IPs, so sync the DB and ensure that it is up
	// to date. It's simpler to update the db since it may have containers that etcd
	// doesn't know about.
	for id, ip := range etcdData.dockerIPs {
		if _, ok := newDockerMap[id]; ok {
			newDockerMap[id] = ip
		}
	}
	syncIPs(newDockerMap, net.IPv4(10, 0, 0, 0))

	if stringMapEquals(etcdData.dockerIPs, newDockerMap) {
		return etcdData, nil
	}

	dbDockerIP, _ := json.Marshal(newDockerMap)
	if err := s.Set(dockerToIPStore, string(dbDockerIP)); err != nil {
		return etcdData, err
	}

	etcdData.dockerIPs = newDockerMap
	return etcdData, nil
}

func updateEtcdLabel(s Store, etcdData storeData, containers []db.Container) (storeData,
	error) {

	// Collect a map of labels to all of the containers that have that label.
	labelContainers := map[string][]db.Container{}
	for _, c := range containers {
		for _, l := range c.Labels {
			labelContainers[l] = append(labelContainers[l], c)
		}
	}

	newMultiHosts := map[string]string{}

	// Gather the multihost containers and set the IPs of non-multihost containers
	// at the same time. The single host IPs are retrieved from the map of container
	// IPs that updateEtcdDocker created.
	for label, cs := range labelContainers {
		if len(cs) > 1 {
			newMultiHosts[label] = ""
		}
	}

	// Etcd is the source of truth for IPs. If the label exists in both etcd and the
	// db and it is a multihost label, then assign it the IP that etcd has.
	// Otherwise, it stays unassigned and syncIPs will take care of it.
	for id := range newMultiHosts {
		if ip, ok := etcdData.multiHost[id]; ok {
			newMultiHosts[id] = ip
		}
	}

	// No need to sync the SingleHost IPs, since they get their IPs from the dockerIP
	// map, which was already synced in updateEtcdDocker
	syncIPs(newMultiHosts, net.IPv4(10, 1, 0, 0))

	if stringMapEquals(newMultiHosts, etcdData.multiHost) {
		return etcdData, nil
	}

	newLabelJSON, _ := json.Marshal(newMultiHosts)
	if err := s.Set(labelToIPStore, string(newLabelJSON)); err != nil {
		return etcdData, err
	}

	etcdData.multiHost = newMultiHosts
	return etcdData, nil
}

func updateDBContainers(view db.Database, etcdData storeData) {
	minion, err := view.MinionSelf()
	worker := err == nil && minion.Role == db.Worker

	etcdKey := func(sc interface{}) interface{} {
		return sc.(storeContainer).DockerID
	}
	dbKey := func(c interface{}) interface{} {
		return c.(db.Container).DockerID
	}

	pairs, dbcs, _ := join.HashJoin(db.ContainerSlice(view.SelectFromContainer(nil)),
		storeContainerSlice(etcdData.containers), dbKey, etcdKey)

	// If etcd hasn't heard of the container, clear its IP
	for _, dbc := range dbcs {
		c := dbc.(db.Container)
		c.IP = ""
		view.Commit(c)
	}

	for _, pair := range pairs {
		dbc := pair.L.(db.Container)
		dbc.IP = ""

		// Workers get their info from Etcd
		if worker {
			etcdc := pair.R.(storeContainer)
			dbc.Labels = etcdc.Labels
			dbc.Command = etcdc.Command
			dbc.Env = etcdc.Env
		}

		// Workers and masters get their IP from Etcd
		if storeIP, ok := etcdData.dockerIPs[dbc.DockerID]; ok {
			dbc.IP = storeIP
			ip := net.ParseIP(storeIP).To4()
			if ip != nil {
				dbc.Mac = fmt.Sprintf("02:00:%02x:%02x:%02x:%02x",
					ip[0], ip[1], ip[2], ip[3])
			}
		}

		view.Commit(dbc)
	}
}

func updateDBLabels(view db.Database, etcdData storeData) {
	// Gather all of the label keys and the IPs for single host labels
	labelIPs := map[string]string{}
	labelKeys := map[string]struct{}{}
	for _, c := range etcdData.containers {
		for _, l := range c.Labels {
			labelKeys[l] = struct{}{}
			if _, ok := etcdData.multiHost[l]; !ok {
				labelIPs[l] = etcdData.dockerIPs[c.DockerID]
			}
		}
	}

	labelKeyFunc := func(val interface{}) interface{} {
		return val.(db.Label).Label
	}

	labelKeySlice := join.StringSlice{}
	for l := range labelKeys {
		labelKeySlice = append(labelKeySlice, l)
	}

	pairs, dbls, dirKeys := join.HashJoin(db.LabelSlice(view.SelectFromLabel(nil)),
		labelKeySlice, labelKeyFunc, nil)

	for _, dbl := range dbls {
		view.Remove(dbl.(db.Label))
	}

	for _, key := range dirKeys {
		pairs = append(pairs, join.Pair{L: view.InsertLabel(), R: key})
	}

	for _, pair := range pairs {
		dbl := pair.L.(db.Label)
		dbl.Label = pair.R.(string)
		if _, ok := etcdData.multiHost[dbl.Label]; ok {
			dbl.IP = etcdData.multiHost[dbl.Label]
			dbl.MultiHost = true
		} else {
			dbl.IP = labelIPs[dbl.Label]
			dbl.MultiHost = false
		}

		view.Commit(dbl)
	}
}

// syncIPs takes a map of IDs to IPs and creates an IP address for every entry that's
// missing one.
func syncIPs(ipMap map[string]string, prefixIP net.IP) {
	prefix := binary.BigEndian.Uint32(prefixIP.To4())
	mask := uint32(0xffff0000)

	var unassigned []string
	ipSet := map[uint32]struct{}{}
	for k, ipString := range ipMap {
		ip := parseIP(ipString, prefix, mask)
		if ip != 0 {
			ipSet[ip] = struct{}{}
		} else {
			unassigned = append(unassigned, k)
		}
	}

	// Don't assign the IP of the default gateway
	ipSet[parseIP(gatewayIP, prefix, mask)] = struct{}{}
	for _, k := range unassigned {
		ip32 := randomIP(ipSet, prefix, mask)
		if ip32 == 0 {
			log.Errorf("Failed to allocate IP for %s.", k)
			ipMap[k] = ""
			continue
		}

		b := make([]byte, 4)
		binary.BigEndian.PutUint32(b, ip32)

		ipMap[k] = net.IP(b).String()
		ipSet[ip32] = struct{}{}
	}
}

func parseIP(ipStr string, prefix, mask uint32) uint32 {
	ip := net.ParseIP(ipStr).To4()
	if ip == nil {
		return 0
	}

	ip32 := binary.BigEndian.Uint32(ip)
	if ip32&mask != prefix {
		return 0
	}

	return ip32
}

// Choose a random IP address in prefix/mask that isn't in 'conflicts'.
// Returns 0 on failure.
func randomIP(conflicts map[uint32]struct{}, prefix, mask uint32) uint32 {
	for i := 0; i < 256; i++ {
		ip32 := (rand32() & ^mask) | (prefix & mask)
		if _, ok := conflicts[ip32]; !ok {
			return ip32
		}
	}

	return 0
}

func stringMapEquals(first, second map[string]string) bool {
	if len(first) != len(second) {
		return false
	}
	for k, f := range first {
		if s, ok := second[k]; !ok || f != s {
			return false
		}
	}
	return true
}

func stringSliceEquals(first, second []string) bool {
	if len(first) != len(second) {
		return false
	}
	for i, s := range first {
		if s != second[i] {
			return false
		}
	}
	return true
}

func (cs storeContainerSlice) Len() int {
	return len(cs)
}

func (cs storeContainerSlice) Less(i, j int) bool {
	return cs[i].DockerID < cs[j].DockerID
}

func (cs storeContainerSlice) Swap(i, j int) {
	cs[i], cs[j] = cs[j], cs[i]
}

func (cs storeContainerSlice) Get(i int) interface{} {
	return cs[i]
}
