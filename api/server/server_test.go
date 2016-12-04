package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
		`"PublicIP":"8.8.8.8","PrivateIP":"9.9.9.9","Connected":false}]`

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
		`"StitchID":0,"DockerID":"docker-id","Image":"image",` +
		`"Command":["cmd","arg"],"Labels":["labelA","labelB"],"Env":null}]`

	checkQuery(t, server{conn}, db.ContainerTable, exp)
}

func TestBadDeployment(t *testing.T) {
	conn := db.New()
	s := server{dbConn: conn}

	badDeployment := `{`

	_, err := s.Deploy(context.Background(),
		&pb.DeployRequest{Deployment: badDeployment})

	if err == nil {
		t.Error("Expected error from bad deployment.")
		return
	}

	expErr := "unexpected end of JSON input"
	if err.Error() != expErr {
		t.Errorf("Expected deployment error %s, but got %s\n",
			expErr, err.Error())
	}
}

func TestDeploy(t *testing.T) {
	conn := db.New()
	s := server{dbConn: conn}

	createMachineDeployment := `
	{"Machines":[
		{"Provider":"Amazon",
		"Role":"Master",
		"Size":"m4.large"
	}, {"Provider":"Amazon",
		"Role":"Worker",
		"Size":"m4.large"
	}]}`

	_, err := s.Deploy(context.Background(),
		&pb.DeployRequest{Deployment: createMachineDeployment})

	if err != nil {
		t.Errorf("Unexpected error when deploying stitch: %s\n", err.Error())
		return
	}

	var machines []db.Machine
	conn.Transact(func(view db.Database) error {
		machines = view.SelectFromMachine(nil)
		return nil
	})

	if len(machines) != 2 {
		t.Errorf("Two machines should have been created by the deployment, "+
			"but we found: %v\n", machines)
	}
}

func TestSwitchNamespace(t *testing.T) {
	conn := db.New()
	s := server{dbConn: conn}

	conn.Transact(func(view db.Database) error {
		clst := view.InsertCluster()
		clst.Namespace = "old-namespace"
		view.Commit(clst)
		return nil
	})

	newNamespaceDeployment := `{"namespace": "new-namespace"}`

	_, err := s.Deploy(context.Background(),
		&pb.DeployRequest{Deployment: newNamespaceDeployment})

	assert.NotNil(t, err)
}
