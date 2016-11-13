package scheduler

import (
	"sync"

	"github.com/NetSys/quilt/db"
	"github.com/NetSys/quilt/join"
	"github.com/NetSys/quilt/minion/docker"
	"github.com/NetSys/quilt/util"
	log "github.com/Sirupsen/logrus"
)

const labelKey = "quilt"
const labelValue = "scheduler"
const labelPair = labelKey + "=" + labelValue
const concurrencyLimit = 1

func runWorker(conn db.Conn, dk docker.Client, myIP string) {
	if myIP == "" {
		return
	}

	filter := map[string][]string{"label": {labelPair}}

	var toBoot, toKill []interface{}
	for i := 0; i < 2; i++ {
		dkcs, err := dk.List(filter)
		if err != nil {
			log.WithError(err).Warning("Failed to list docker containers.")
			return
		}

		conn.Transact(func(view db.Database) error {
			dbcs := view.SelectFromContainer(func(dbc db.Container) bool {
				return dbc.Minion == myIP
			})

			var changed []db.Container
			changed, toBoot, toKill = syncWorker(dbcs, dkcs)
			for _, dbc := range changed {
				view.Commit(dbc)
			}

			// XXX: We do the actual booting and destruction in the
			// transaction so that we can prevent the network from running
			// while the containers are being created. This can cause the
			// occasional deadlock.  This should be undone once we have the
			// network driver in place.
			doContainers(dk, toBoot, dockerRun)
			doContainers(dk, toKill, dockerKill)
			return nil
		})

	}
}

func syncWorker(dbcs []db.Container, dkcs []docker.Container) (changed []db.Container,
	toBoot, toKill []interface{}) {

	pairs, dbci, dkci := join.Join(dbcs, dkcs, syncJoinScore)

	for _, i := range dkci {
		toKill = append(toKill, i.(docker.Container))
	}

	for _, i := range dbci {
		toBoot = append(toBoot, i.(db.Container))
	}

	for _, pair := range pairs {
		dbc := pair.L.(db.Container)
		dkc := pair.R.(docker.Container)

		if dbc.DockerID != dkc.ID {
			dbc.DockerID = dkc.ID
			dbc.Pid = dkc.Pid
			changed = append(changed, dbc)
		}
	}

	return changed, toBoot, toKill
}

func doContainers(dk docker.Client, containers []interface{},
	do func(docker.Client, chan interface{})) {

	in := make(chan interface{})
	var wg sync.WaitGroup
	wg.Add(concurrencyLimit)
	for i := 0; i < concurrencyLimit; i++ {
		go func() {
			do(dk, in)
			wg.Done()
		}()
	}

	for _, dbc := range containers {
		in <- dbc
	}
	close(in)
	wg.Wait()
}

func dockerRun(dk docker.Client, in chan interface{}) {
	for i := range in {
		dbc := i.(db.Container)
		log.WithField("container", dbc).Info("Start container")
		_, err := dk.Run(docker.RunOptions{
			Image:  dbc.Image,
			Args:   dbc.Command,
			Env:    dbc.Env,
			Labels: map[string]string{labelKey: labelValue},
		})
		if err != nil {
			log.WithFields(log.Fields{
				"error":     err,
				"container": dbc,
			}).WithError(err).Warning("Failed to run container", dbc)
			continue
		}
	}
}

func dockerKill(dk docker.Client, in chan interface{}) {
	for i := range in {
		dkc := i.(docker.Container)
		log.WithField("container", dkc.ID).Info("Remove container")
		if err := dk.RemoveID(dkc.ID); err != nil {
			log.WithFields(log.Fields{
				"error": err,
				"id":    dkc.ID,
			}).Warning("Failed to remove container.")
		}
	}
}

func syncJoinScore(left, right interface{}) int {
	dbc := left.(db.Container)
	dkc := right.(docker.Container)

	// Depending on the container, the command in the database could be
	// either The command plus it's arguments, or just it's arguments.  To
	// handle that case, we check both.
	cmd1 := dkc.Args
	cmd2 := append([]string{dkc.Path}, dkc.Args...)
	dbcCmd := dbc.Command

	for key, value := range dbc.Env {
		if dkc.Env[key] != value {
			return -1
		}
	}

	switch {
	case dbc.Image != dkc.Image:
		return -1
	case len(dbcCmd) != 0 &&
		!util.StrSliceEqual(dbcCmd, cmd1) &&
		!util.StrSliceEqual(dbcCmd, cmd2):
		return -1
	case dbc.DockerID == dkc.ID:
		return 0
	default:
		return 1
	}
}
