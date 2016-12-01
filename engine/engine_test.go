package engine

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/NetSys/quilt/db"
	"github.com/NetSys/quilt/join"
	"github.com/NetSys/quilt/stitch"
	"github.com/davecgh/go-spew/spew"
)

func TestEngine(t *testing.T) {
	spew := spew.NewDefaultConfig()
	spew.MaxDepth = 2

	pre := `var deployment = createDeployment({
		namespace: "namespace",
		adminACL: ["1.2.3.4/32"],
	});
	var baseMachine = new Machine({provider: "Amazon", size: "m4.large"});`
	conn := db.New()

	code := pre + `deployment.deploy(baseMachine.asMaster().replicate(2));
		deployment.deploy(baseMachine.asWorker().replicate(3));`

	UpdatePolicy(conn, prog(t, code))
	err := conn.Transact(func(view db.Database) error {
		acl, err := view.GetACL()
		masters := view.SelectFromMachine(func(m db.Machine) bool {
			return m.Role == db.Master
		})
		workers := view.SelectFromMachine(func(m db.Machine) bool {
			return m.Role == db.Worker
		})

		if err != nil {
			return err
		} else if len(acl.Admin) != 1 {
			return fmt.Errorf("bad acls: %s", spew.Sdump(acl))
		}

		if len(masters) != 2 {
			return fmt.Errorf("bad masters: %s", spew.Sdump(masters))
		}

		if len(workers) != 3 {
			return fmt.Errorf("bad workers: %s", spew.Sdump(workers))
		}
		return nil
	})
	if err != nil {
		t.Error(err.Error())
	}

	/* Verify master increase. */
	code = pre + `deployment.deploy(baseMachine.asMaster().replicate(4));
		deployment.deploy(baseMachine.asWorker().replicate(5));`

	UpdatePolicy(conn, prog(t, code))
	err = conn.Transact(func(view db.Database) error {
		masters := view.SelectFromMachine(func(m db.Machine) bool {
			return m.Role == db.Master
		})
		workers := view.SelectFromMachine(func(m db.Machine) bool {
			return m.Role == db.Worker
		})

		if len(masters) != 4 {
			return fmt.Errorf("bad masters: %s", spew.Sdump(masters))
		}

		if len(workers) != 5 {
			return fmt.Errorf("bad workers: %s", spew.Sdump(workers))
		}
		return nil
	})
	if err != nil {
		t.Error(err.Error())
	}

	/* Verify that external writes stick around. */
	err = conn.Transact(func(view db.Database) error {
		masters := view.SelectFromMachine(func(m db.Machine) bool {
			return m.Role == db.Master
		})
		workers := view.SelectFromMachine(func(m db.Machine) bool {
			return m.Role == db.Worker
		})

		for _, master := range masters {
			master.CloudID = "1"
			master.PublicIP = "2"
			master.PrivateIP = "3"
			view.Commit(master)
		}

		for _, worker := range workers {
			worker.CloudID = "1"
			worker.PublicIP = "2"
			worker.PrivateIP = "3"
			view.Commit(worker)
		}

		return nil
	})

	/* Also verify that masters and workers decrease properly. */
	code = pre + `deployment.deploy(baseMachine.asMaster());
		deployment.deploy(baseMachine.asWorker());`
	UpdatePolicy(conn, prog(t, code))
	err = conn.Transact(func(view db.Database) error {
		masters := view.SelectFromMachine(func(m db.Machine) bool {
			return m.Role == db.Master
		})
		workers := view.SelectFromMachine(func(m db.Machine) bool {
			return m.Role == db.Worker
		})

		if len(masters) != 1 || masters[0].CloudID != "1" ||
			masters[0].PublicIP != "2" || masters[0].PrivateIP != "3" {
			return fmt.Errorf("bad masters: %s", spew.Sdump(masters))
		}

		if len(workers) != 1 || workers[0].CloudID != "1" ||
			workers[0].PublicIP != "2" || workers[0].PrivateIP != "3" {
			return fmt.Errorf("bad workers: %s", spew.Sdump(workers))
		}
		return nil
	})
	if err != nil {
		t.Error(err.Error())
	}

	/* Empty Namespace does nothing. */
	code = pre + `deployment.namespace = "";
		deployment.deploy(baseMachine.asMaster());
		deployment.deploy(baseMachine.asWorker());`
	UpdatePolicy(conn, prog(t, code))
	err = conn.Transact(func(view db.Database) error {
		masters := view.SelectFromMachine(func(m db.Machine) bool {
			return m.Role == db.Master
		})
		workers := view.SelectFromMachine(func(m db.Machine) bool {
			return m.Role == db.Worker
		})

		if len(masters) != 1 || masters[0].CloudID != "1" ||
			masters[0].PublicIP != "2" || masters[0].PrivateIP != "3" {
			return fmt.Errorf("bad masters: %s", spew.Sdump(masters))
		}

		if len(workers) != 1 || workers[0].CloudID != "1" ||
			workers[0].PublicIP != "2" || workers[0].PrivateIP != "3" {
			return fmt.Errorf("bad workers: %s", spew.Sdump(workers))
		}
		return nil
	})
	if err != nil {
		t.Error(err.Error())
	}

	/* Verify things go to zero. */
	code = pre + `deployment.deploy(baseMachine.asWorker())`
	UpdatePolicy(conn, prog(t, code))
	err = conn.Transact(func(view db.Database) error {
		masters := view.SelectFromMachine(func(m db.Machine) bool {
			return m.Role == db.Master
		})
		workers := view.SelectFromMachine(func(m db.Machine) bool {
			return m.Role == db.Worker
		})

		if len(masters) != 0 {
			return fmt.Errorf("bad masters: %s", spew.Sdump(masters))
		}

		if len(workers) != 0 {
			return fmt.Errorf("bad workers: %s", spew.Sdump(workers))
		}
		return nil
	})
	if err != nil {
		t.Error(err.Error())
	}

	// This function checks whether there is a one-to-one mapping for each machine
	// in `slice` to a provider in `providers`.
	providersInSlice := func(slice db.MachineSlice, providers db.ProviderSlice) bool {
		lKey := func(left interface{}) interface{} {
			return left.(db.Machine).Provider
		}
		rKey := func(right interface{}) interface{} {
			return right.(db.Provider)
		}
		_, l, r := join.HashJoin(slice, providers, lKey, rKey)
		return len(l) == 0 && len(r) == 0
	}

	/* Test mixed providers. */
	code = `deployment.deploy([
		new Machine({provider: "Amazon", size: "m4.large", role: "Master"}),
		new Machine({provider: "Vagrant", size: "v.large", role: "Master"}),
		new Machine({provider: "Amazon", size: "m4.large", role: "Worker"}),
		new Machine({provider: "Google", size: "g.large", role: "Worker"})]);`
	UpdatePolicy(conn, prog(t, code))
	err = conn.Transact(func(view db.Database) error {
		masters := view.SelectFromMachine(func(m db.Machine) bool {
			return m.Role == db.Master
		})
		workers := view.SelectFromMachine(func(m db.Machine) bool {
			return m.Role == db.Worker
		})

		if !providersInSlice(masters, db.ProviderSlice{db.Amazon, db.Vagrant}) {
			return fmt.Errorf("bad masters: %s", spew.Sdump(masters))
		}

		if !providersInSlice(workers, db.ProviderSlice{db.Amazon, db.Google}) {
			return fmt.Errorf("bad workers: %s", spew.Sdump(workers))
		}
		return nil
	})
	if err != nil {
		t.Error(err.Error())
	}

	/* Test that machines with different providers don't match. */
	code = `deployment.deploy([
		new Machine({provider: "Amazon", size: "m4.large", role: "Master"}),
		new Machine({provider: "Amazon", size: "m4.large", role: "Worker"})]);`
	UpdatePolicy(conn, prog(t, code))
	err = conn.Transact(func(view db.Database) error {
		masters := view.SelectFromMachine(func(m db.Machine) bool {
			return m.Role == db.Master
		})

		if !providersInSlice(masters, db.ProviderSlice{db.Amazon}) {
			return fmt.Errorf("bad masters: %s", spew.Sdump(masters))
		}
		return nil
	})
	if err != nil {
		t.Error(err.Error())
	}
}

func TestSort(t *testing.T) {
	spew := spew.NewDefaultConfig()
	spew.MaxDepth = 2

	pre := `var baseMachine = new Machine({provider: "Amazon", size: "m4.large"});`
	conn := db.New()

	UpdatePolicy(conn, prog(t, pre+`
	deployment.deploy(baseMachine.asMaster().replicate(3));
	deployment.deploy(baseMachine.asWorker().replicate(1));`))
	err := conn.Transact(func(view db.Database) error {
		machines := view.SelectFromMachine(func(m db.Machine) bool {
			return m.Role == db.Master
		})

		if len(machines) != 3 {
			return fmt.Errorf("bad machines: %s", spew.Sdump(machines))
		}

		machines[2].PublicIP = "a"
		machines[2].PrivateIP = "b"
		view.Commit(machines[2])

		machines[1].PrivateIP = "c"
		view.Commit(machines[1])

		return nil
	})
	if err != nil {
		t.Error(err.Error())
	}

	UpdatePolicy(conn, prog(t, pre+`
	deployment.deploy(baseMachine.asMaster().replicate(2));
	deployment.deploy(baseMachine.asWorker().replicate(1));`))
	err = conn.Transact(func(view db.Database) error {
		machines := view.SelectFromMachine(func(m db.Machine) bool {
			return m.Role == db.Master
		})

		if len(machines) != 2 {
			return fmt.Errorf("bad machines: %s", spew.Sdump(machines))
		}

		for _, m := range machines {
			if m.PublicIP == "" && m.PrivateIP == "" {
				return fmt.Errorf("bad machine: %s",
					spew.Sdump(machines))
			}
		}

		return nil
	})
	if err != nil {
		t.Error(err.Error())
	}

	UpdatePolicy(conn, prog(t, pre+`
	deployment.deploy(baseMachine.asMaster().replicate(1));
	deployment.deploy(baseMachine.asWorker().replicate(1));`))
	err = conn.Transact(func(view db.Database) error {
		machines := view.SelectFromMachine(func(m db.Machine) bool {
			return m.Role == db.Master
		})

		if len(machines) != 1 {
			return fmt.Errorf("bad machines: %s", spew.Sdump(machines))
		}

		for _, m := range machines {
			if m.PublicIP == "" || m.PrivateIP == "" {
				return fmt.Errorf("bad machine: %s",
					spew.Sdump(machines))
			}
		}

		return nil
	})
	if err != nil {
		t.Error(err.Error())
	}
}

func TestACLs(t *testing.T) {
	spew := spew.NewDefaultConfig()
	spew.MaxDepth = 2

	conn := db.New()

	code := `createDeployment({adminACL: ["1.2.3.4/32", "local"]})
		.deploy([
			new Machine({provider: "Amazon", role: "Master"}),
			new Machine({provider: "Amazon", role: "Worker"})
		]);`

	myIP = func() (string, error) {
		return "5.6.7.8", nil
	}
	UpdatePolicy(conn, prog(t, code))
	err := conn.Transact(func(view db.Database) error {
		acl, err := view.GetACL()

		if err != nil {
			return err
		}

		if !reflect.DeepEqual(acl.Admin,
			[]string{"1.2.3.4/32", "5.6.7.8/32"}) {
			return fmt.Errorf("bad ACLs: %s",
				spew.Sdump(acl.Admin))
		}

		return nil
	})
	if err != nil {
		t.Error(err.Error())
	}

	myIP = func() (string, error) {
		return "", errors.New("")
	}
	UpdatePolicy(conn, prog(t, code))
	err = conn.Transact(func(view db.Database) error {
		acl, err := view.GetACL()

		if err != nil {
			return err
		}

		if !reflect.DeepEqual(acl.Admin, []string{"1.2.3.4/32"}) {
			return fmt.Errorf("bad ACLs: %s",
				spew.Sdump(acl.Admin))
		}

		return nil
	})
	if err != nil {
		t.Error(err.Error())
	}
}

func prog(t *testing.T, code string) stitch.Stitch {
	result, err := stitch.FromJavascript(code, stitch.DefaultImportGetter)
	if err != nil {
		t.Error(err.Error())
		return stitch.Stitch{}
	}

	return result
}
