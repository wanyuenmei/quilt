package minion

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/NetSys/quilt/db"
	"github.com/NetSys/quilt/stitch"
	"github.com/NetSys/quilt/util"
	"github.com/davecgh/go-spew/spew"
)

const testImage = "alpine"

func TestContainerTxn(t *testing.T) {
	conn := db.New()
	trigg := conn.Trigger(db.ContainerTable).C

	spec := ""
	if err := testContainerTxn(conn, spec); err != "" {
		t.Error(err)
	}
	if fired(trigg) {
		t.Error("Unexpected Database Change")
	}

	spec = `deployment.deploy(
		new Service("a", [new Container("alpine", ["tail"])])
	)`
	if err := testContainerTxn(conn, spec); err != "" {
		t.Error(err)
	}
	if !fired(trigg) {
		t.Error("Expected Database Change")
	}

	if err := testContainerTxn(conn, spec); err != "" {
		t.Error(err)
	}
	if fired(trigg) {
		t.Error("Unexpected Database Change")
	}

	spec = `var b = new Container("alpine", ["tail"]);
	deployment.deploy([
		new Service("b", [b]),
		new Service("a", [b, new Container("alpine", ["tail"])])
	]);`
	if err := testContainerTxn(conn, spec); err != "" {
		t.Error(err)
	}
	if !fired(trigg) {
		t.Error("Expected Database Change")
	}

	spec = `var b = new Service("b", [new Container("alpine", ["cat"])]);
	deployment.deploy([
		b,
		new Service("a",
			b.containers.concat([new Container("alpine", ["tail"])])),
	]);`
	if err := testContainerTxn(conn, spec); err != "" {
		t.Error(err)
	}
	if !fired(trigg) {
		t.Error("Expected Database Change")
	}

	spec = `var b = new Service("b", [new Container("ubuntu", ["cat"])]);
	deployment.deploy([
		b,
		new Service("a",
			b.containers.concat([new Container("alpine", ["tail"])])),
	]);`
	if err := testContainerTxn(conn, spec); err != "" {
		t.Error(err)
	}
	if !fired(trigg) {
		t.Error("Expected Database Change")
	}

	spec = `deployment.deploy(
		new Service("a", [
			new Container("alpine", ["cat"]),
			new Container("alpine", ["cat"])
		])
	);`
	if err := testContainerTxn(conn, spec); err != "" {
		t.Error(err)
	}
	if !fired(trigg) {
		t.Error("Expected Database Change")
	}

	spec = `deployment.deploy(
		new Service("a", [new Container("alpine")])
	)`
	if err := testContainerTxn(conn, spec); err != "" {
		t.Error(err)
	}
	if !fired(trigg) {
		t.Error("Expected Database Change")
	}

	spec = `var b = new Service("b", [new Container("alpine")]);
	var c = new Service("c", [new Container("alpine")]);
	deployment.deploy([
		b,
		c,
		new Service("a", b.containers.concat(c.containers)),
	])`
	if err := testContainerTxn(conn, spec); err != "" {
		t.Error(err)
	}
	if !fired(trigg) {
		t.Error("Expected Database Change")
	}

	if err := testContainerTxn(conn, spec); err != "" {
		t.Error(err)
	}
	if fired(trigg) {
		t.Error("Unexpected Database Change")
	}
}

func testContainerTxn(conn db.Conn, spec string) string {
	compiled, err := stitch.FromJavascript(spec, stitch.DefaultImportGetter)
	if err != nil {
		return err.Error()
	}

	var containers []db.Container
	conn.Transact(func(view db.Database) error {
		updatePolicy(view, db.Master, compiled.String())
		containers = view.SelectFromContainer(nil)
		return nil
	})

	for _, e := range queryContainers(compiled) {
		found := false
		for i, c := range containers {
			if e.Image == c.Image &&
				reflect.DeepEqual(e.Command, c.Command) &&
				util.EditDistance(c.Labels, e.Labels) == 0 {
				containers = append(containers[:i], containers[i+1:]...)
				found = true
				break
			}
		}

		if found == false {
			return fmt.Sprintf("Missing expected label set: %v\n%v",
				e, containers)
		}
	}

	if len(containers) > 0 {
		return spew.Sprintf("Unexpected containers: %s", containers)
	}

	return ""
}

func TestConnectionTxn(t *testing.T) {
	conn := db.New()
	trigg := conn.Trigger(db.ConnectionTable).C

	pre := `var a = new Service("a", [new Container("alpine")]);
	var b = new Service("b", [new Container("alpine")]);
	var c = new Service("c", [new Container("alpine")]);
	deployment.deploy([a, b, c]);`

	spec := ""
	if err := testConnectionTxn(conn, spec); err != "" {
		t.Error(err)
	}
	if fired(trigg) {
		t.Error("Unexpected Database Change")
	}

	spec = pre + `a.connect(80, a);`
	if err := testConnectionTxn(conn, spec); err != "" {
		t.Error(err)
	}
	if !fired(trigg) {
		t.Error("Expected Database Change")
	}
	if err := testConnectionTxn(conn, spec); err != "" {
		t.Error(err)
	}
	if fired(trigg) {
		t.Error("Unexpected Database Change")
	}

	spec = pre + `a.connect(90, a);`
	if err := testConnectionTxn(conn, spec); err != "" {
		t.Error(err)
	}
	if !fired(trigg) {
		t.Error("Expected Database Change")
	}
	if err := testConnectionTxn(conn, spec); err != "" {
		t.Error(err)
	}
	if fired(trigg) {
		t.Error("Unexpected Database Change")
	}

	spec = pre + `b.connect(90, a);
	b.connect(90, c);
	b.connect(100, b);
	c.connect(101, a);`
	if err := testConnectionTxn(conn, spec); err != "" {
		t.Error(err)
	}
	if !fired(trigg) {
		t.Error("Expected Database Change")
	}
	if err := testConnectionTxn(conn, spec); err != "" {
		t.Error(err)
	}
	if fired(trigg) {
		t.Error("Unexpected Database Change")
	}

	spec = pre
	if err := testConnectionTxn(conn, spec); err != "" {
		t.Error(err)
	}
	if !fired(trigg) {
		t.Error("Expected Database Change")
	}
	if err := testConnectionTxn(conn, spec); err != "" {
		t.Error(err)
	}
	if fired(trigg) {
		t.Error("Unexpected Database Change")
	}
}

func testConnectionTxn(conn db.Conn, spec string) string {
	compiled, err := stitch.FromJavascript(spec, stitch.DefaultImportGetter)
	if err != nil {
		return err.Error()
	}

	var connections []db.Connection
	conn.Transact(func(view db.Database) error {
		updatePolicy(view, db.Master, compiled.String())
		connections = view.SelectFromConnection(nil)
		return nil
	})

	exp := compiled.Connections
	for _, e := range exp {
		found := false
		for i, c := range connections {
			if e.From == c.From && e.To == c.To && e.MinPort == c.MinPort &&
				e.MaxPort == c.MaxPort {
				connections = append(
					connections[:i], connections[i+1:]...)
				found = true
				break
			}
		}

		if found == false {
			return fmt.Sprintf("Missing expected connection: %v", e)
		}
	}

	if len(connections) > 0 {
		return spew.Sprintf("Unexpected connections: %s", connections)
	}

	return ""
}

func fired(c chan struct{}) bool {
	time.Sleep(5 * time.Millisecond)
	select {
	case <-c:
		return true
	default:
		return false
	}
}

func TestPlacementTxn(t *testing.T) {
	conn := db.New()
	checkPlacement := func(spec string, exp ...db.Placement) {
		compiled, err := stitch.FromJavascript(spec,
			stitch.DefaultImportGetter)
		if err != nil {
			t.Errorf("Unexpected error while compiling: %s", err.Error())
		}

		placements := map[db.Placement]struct{}{}
		conn.Transact(func(view db.Database) error {
			updatePolicy(view, db.Master, compiled.String())
			res := view.SelectFromPlacement(nil)

			// Set the ID to 0 so that we can use reflect.DeepEqual.
			for _, p := range res {
				p.ID = 0
				placements[p] = struct{}{}
			}

			return nil
		})

		if len(placements) != len(exp) {
			t.Errorf("Placement error in %s. Expected %v, got %v",
				spec, exp, placements)
		}

		for _, p := range exp {
			if _, ok := placements[p]; !ok {
				t.Errorf("Placement error in %s. Expected %v, got %v",
					spec, exp, placements)
				break
			}
		}
	}

	pre := `var foo = new Service("foo", [new Container("foo")]);
	var bar = new Service("bar", [new Container("bar")]);
	var baz = new Service("baz", [new Container("bar")]);
	deployment.deploy([foo, bar, baz]);`

	// Create an exclusive placement.
	spec := pre + `bar.place(new LabelRule(true, foo));`
	checkPlacement(spec,
		db.Placement{
			TargetLabel: "bar",
			Exclusive:   true,
			OtherLabel:  "foo",
		},
	)

	// Change the placement from "exclusive" to "on".
	spec = pre + `bar.place(new LabelRule(false, foo));`
	checkPlacement(spec,
		db.Placement{
			TargetLabel: "bar",
			Exclusive:   false,
			OtherLabel:  "foo",
		},
	)

	// Add another placement constraint.
	spec = pre + `bar.place(new LabelRule(false, foo));
	bar.place(new LabelRule(true, bar));`
	checkPlacement(spec,
		db.Placement{
			TargetLabel: "bar",
			Exclusive:   false,
			OtherLabel:  "foo",
		},
		db.Placement{
			TargetLabel: "bar",
			Exclusive:   true,
			OtherLabel:  "bar",
		},
	)

	// Machine placement
	spec = pre + `foo.place(new MachineRule(false, {size: "m4.large"}));`
	checkPlacement(spec,
		db.Placement{
			TargetLabel: "foo",
			Exclusive:   false,
			Size:        "m4.large",
		},
	)

	// Port placement
	spec = pre + `publicInternet.connect(80, foo);
	publicInternet.connect(81, foo);`
	checkPlacement(spec,
		db.Placement{
			TargetLabel: "foo",
			Exclusive:   true,
			OtherLabel:  "foo",
		},
	)

	spec = pre + `publicInternet.connect(80, foo);
	publicInternet.connect(80, bar);
	(function() {
		publicInternet.connect(81, bar);
		publicInternet.connect(81, baz);
	})()`

	checkPlacement(spec,
		db.Placement{
			TargetLabel: "foo",
			Exclusive:   true,
			OtherLabel:  "foo",
		},

		db.Placement{
			TargetLabel: "bar",
			Exclusive:   true,
			OtherLabel:  "bar",
		},

		db.Placement{
			TargetLabel: "foo",
			Exclusive:   true,
			OtherLabel:  "bar",
		},

		db.Placement{
			TargetLabel: "bar",
			Exclusive:   true,
			OtherLabel:  "foo",
		},

		db.Placement{
			TargetLabel: "baz",
			Exclusive:   true,
			OtherLabel:  "baz",
		},

		db.Placement{
			TargetLabel: "bar",
			Exclusive:   true,
			OtherLabel:  "baz",
		},

		db.Placement{
			TargetLabel: "baz",
			Exclusive:   true,
			OtherLabel:  "bar",
		},
	)
}
