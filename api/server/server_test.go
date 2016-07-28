package server

import (
	"testing"

	"golang.org/x/net/context"

	"github.com/NetSys/quilt/api/pb"
	"github.com/NetSys/quilt/db"
)

func checkQuery(t *testing.T, s server, table db.TableType, exp string) {
	reply, err := s.Query(context.Background(),
		&pb.DBQuery{Table: string(table)})
	if err != nil {
		t.Errorf("Unexpected error: %s\n", err.Error())
		return
	}

	if exp != reply.TableContents {
		t.Errorf(`Bad query response: expected "%s", got "%s".`,
			exp, reply.TableContents)
	}
}

func TestMachineResponse(t *testing.T) {
	t.Parallel()

	conn := db.New()
	conn.Transact(func(view db.Database) error {
		m := view.InsertMachine()
		m.Role = db.Master
		m.Provider = db.Amazon
		m.Size = "size"
		m.PublicIP = "8.8.8.8"
		m.PrivateIP = "9.9.9.9"
		view.Commit(m)

		return nil
	})

	exp := `[{"ID":1,"Role":"Master","Provider":"Amazon","Region":"",` +
		`"Size":"size","DiskSize":0,"SSHKeys":null,"CloudID":"",` +
		`"PublicIP":"8.8.8.8","PrivateIP":"9.9.9.9"}]`

	checkQuery(t, server{conn}, db.MachineTable, exp)
}

func TestContainerResponse(t *testing.T) {
	t.Parallel()

	conn := db.New()
	conn.Transact(func(view db.Database) error {
		c := view.InsertContainer()
		c.DockerID = "docker-id"
		c.Image = "image"
		c.Command = []string{"cmd", "arg"}
		c.Labels = []string{"labelA", "labelB"}
		view.Commit(c)

		return nil
	})

	exp := `[{"ID":1,"Pid":0,"IP":"","Mac":"","Minion":"",` +
		`"DockerID":"docker-id","StitchID":0,"Image":"image",` +
		`"Command":["cmd","arg"],"Labels":["labelA","labelB"],"Env":null}]`

	checkQuery(t, server{conn}, db.ContainerTable, exp)
}
