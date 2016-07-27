package etcd

import (
	"time"

	"github.com/NetSys/quilt/db"

	log "github.com/Sirupsen/logrus"
	"github.com/coreos/etcd/client"
)

// Run synchronizes state in `conn` with the Etcd cluster.
func Run(conn db.Conn) {
	store := NewStore()
	<-store.BootWait()

	createMinionDir(store)

	go runElection(conn, store)
	runNetwork(conn, store)
}

func createMinionDir(store Store) {
	for {
		err := store.Mkdir(minionDir)
		if err == nil {
			return
		}

		// If the directory already exists, no need to create it
		etcdErr, ok := err.(client.Error)
		if ok && etcdErr.Code == client.ErrorCodeNodeExist {
			log.WithError(etcdErr).Debug()
			return
		}

		log.WithError(err).Warn()
		time.Sleep(5 * time.Second)
	}
}
