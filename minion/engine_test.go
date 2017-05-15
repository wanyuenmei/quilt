package minion

import (
	"testing"
	"time"

	"github.com/quilt/quilt/db"
	"github.com/quilt/quilt/join"
	"github.com/quilt/quilt/stitch"
	"github.com/stretchr/testify/assert"
)

const testImage = "alpine"

func TestContainerTxn(t *testing.T) {
	conn := db.New()
	trigg := conn.Trigger(db.ContainerTable).C

	testContainerTxn(t, conn, stitch.Stitch{})
	assert.False(t, fired(trigg))

	stc := stitch.Stitch{
		Containers: []stitch.Container{
			{
				ID:      "f133411ac23f45342a7b8b89bbe5e9efd0e711e5",
				Image:   stitch.Image{Name: "alpine"},
				Command: []string{"tail"},
			},
		},
		Labels: []stitch.Label{
			{
				Name: "a",
				IDs: []string{
					"f133411ac23f45342a7b8b89bbe5e9efd0e711e5",
				},
			},
		},
	}
	testContainerTxn(t, conn, stc)
	assert.True(t, fired(trigg))

	testContainerTxn(t, conn, stc)
	assert.False(t, fired(trigg))

	testContainerTxn(t, conn, stitch.Stitch{
		Containers: []stitch.Container{
			{
				ID:      "f133411ac23f45342a7b8b89bbe5e9efd0e711e5",
				Image:   stitch.Image{Name: "alpine"},
				Command: []string{"tail"},
			},
			{
				ID:      "6e24c8cbeb63dbffcc82730d01b439e2f5085f59",
				Image:   stitch.Image{Name: "alpine"},
				Command: []string{"tail"},
			},
		},
		Labels: []stitch.Label{
			{
				Name: "b",
				IDs: []string{
					"f133411ac23f45342a7b8b89bbe5e9efd0e711e5",
				},
			},
			{
				Name: "a",
				IDs: []string{
					"f133411ac23f45342a7b8b89bbe5e9efd0e711e5",
					"6e24c8cbeb63dbffcc82730d01b439e2f5085f59",
				},
			},
		},
	})
	assert.True(t, fired(trigg))

	testContainerTxn(t, conn, stitch.Stitch{
		Containers: []stitch.Container{
			{
				ID:      "0b8a2ed7d14d78a388375025223b05d072bbaec3",
				Image:   stitch.Image{Name: "alpine"},
				Command: []string{"cat"},
			},
			{
				ID:      "f133411ac23f45342a7b8b89bbe5e9efd0e711e5",
				Image:   stitch.Image{Name: "alpine"},
				Command: []string{"tail"},
			},
		},
		Labels: []stitch.Label{
			{
				Name: "b",
				IDs: []string{
					"0b8a2ed7d14d78a388375025223b05d072bbaec3",
				},
			},
			{
				Name: "a",
				IDs: []string{
					"0b8a2ed7d14d78a388375025223b05d072bbaec3",
					"f133411ac23f45342a7b8b89bbe5e9efd0e711e5",
				},
			},
		},
	})
	assert.True(t, fired(trigg))

	testContainerTxn(t, conn, stitch.Stitch{
		Containers: []stitch.Container{
			{
				ID:      "7a6244b8d2bfa10ee2fcbe6836a0519e116aee31",
				Image:   stitch.Image{Name: "ubuntu"},
				Command: []string{"cat"},
			},
			{
				ID:      "f133411ac23f45342a7b8b89bbe5e9efd0e711e5",
				Image:   stitch.Image{Name: "alpine"},
				Command: []string{"tail"},
			},
		},
		Labels: []stitch.Label{
			{
				Name: "b",
				IDs: []string{
					"7a6244b8d2bfa10ee2fcbe6836a0519e116aee31",
				},
			},
			{
				Name: "a",
				IDs: []string{
					"7a6244b8d2bfa10ee2fcbe6836a0519e116aee31",
					"f133411ac23f45342a7b8b89bbe5e9efd0e711e5",
				},
			},
		},
	})
	assert.True(t, fired(trigg))

	testContainerTxn(t, conn, stitch.Stitch{
		Containers: []stitch.Container{
			{
				ID:      "0b8a2ed7d14d78a388375025223b05d072bbaec3",
				Image:   stitch.Image{Name: "alpine"},
				Command: []string{"cat"},
			},
			{
				ID:      "d1c9f501efd7a348e54388358c5fe29690fb147d",
				Image:   stitch.Image{Name: "alpine"},
				Command: []string{"cat"},
			},
		},
		Labels: []stitch.Label{
			{
				Name: "a",
				IDs: []string{
					"0b8a2ed7d14d78a388375025223b05d072bbaec3",
					"d1c9f501efd7a348e54388358c5fe29690fb147d",
				},
			},
		},
	})
	assert.True(t, fired(trigg))

	testContainerTxn(t, conn, stitch.Stitch{
		Containers: []stitch.Container{
			{
				ID:    "018e4ee517d85640d9bf0adb4579d2ac9bd358af",
				Image: stitch.Image{Name: "alpine"},
			},
		},
		Labels: []stitch.Label{
			{
				Name: "a",
				IDs: []string{
					"018e4ee517d85640d9bf0adb4579d2ac9bd358af",
				},
			},
		},
	})
	assert.True(t, fired(trigg))

	stc = stitch.Stitch{
		Containers: []stitch.Container{
			{
				ID:    "018e4ee517d85640d9bf0adb4579d2ac9bd358af",
				Image: stitch.Image{Name: "alpine"},
			},
			{
				ID:    "ac4693f0b7fc17aa0e885aa03dc8f7cd6017f496",
				Image: stitch.Image{Name: "alpine"},
			},
		},
		Labels: []stitch.Label{
			{
				Name: "b",
				IDs: []string{
					"018e4ee517d85640d9bf0adb4579d2ac9bd358af",
				},
			},
			{
				Name: "c",
				IDs: []string{
					"ac4693f0b7fc17aa0e885aa03dc8f7cd6017f496",
				},
			},
			{
				Name: "a",
				IDs: []string{
					"018e4ee517d85640d9bf0adb4579d2ac9bd358af",
					"ac4693f0b7fc17aa0e885aa03dc8f7cd6017f496",
				},
			},
		},
	}
	testContainerTxn(t, conn, stc)
	assert.True(t, fired(trigg))

	testContainerTxn(t, conn, stc)
	assert.False(t, fired(trigg))
}

func testContainerTxn(t *testing.T, conn db.Conn, stc stitch.Stitch) {
	var containers []db.Container
	conn.Txn(db.AllTables...).Run(func(view db.Database) error {
		updatePolicy(view, stc.String())
		containers = view.SelectFromContainer(nil)
		return nil
	})

	for _, e := range queryContainers(stc) {
		found := false
		for i, c := range containers {
			if e.StitchID == c.StitchID {
				containers = append(containers[:i], containers[i+1:]...)
				found = true
				break
			}
		}

		assert.True(t, found)
	}

	assert.Empty(t, containers)
}

func TestConnectionTxn(t *testing.T) {
	conn := db.New()
	trigg := conn.Trigger(db.ConnectionTable).C

	testConnectionTxn(t, conn, stitch.Stitch{})
	assert.False(t, fired(trigg))

	stc := stitch.Stitch{
		Containers: []stitch.Container{
			{
				ID:    "018e4ee517d85640d9bf0adb4579d2ac9bd358af",
				Image: stitch.Image{Name: "alpine"},
			},
			{
				ID:    "ac4693f0b7fc17aa0e885aa03dc8f7cd6017f496",
				Image: stitch.Image{Name: "alpine"},
			},
			{
				ID:    "6c1423bb4006da48cd36aa664afec85f05575702",
				Image: stitch.Image{Name: "alpine"},
			},
		},
		Labels: []stitch.Label{
			{
				Name: "a",
				IDs: []string{
					"018e4ee517d85640d9bf0adb4579d2ac9bd358af",
				},
			},
			{
				Name: "b",
				IDs: []string{
					"ac4693f0b7fc17aa0e885aa03dc8f7cd6017f496",
				},
			},
			{
				Name: "c",
				IDs: []string{
					"6c1423bb4006da48cd36aa664afec85f05575702",
				},
			},
		},
		Connections: []stitch.Connection{
			{From: "a", To: "a", MinPort: 80, MaxPort: 80},
		},
	}
	testConnectionTxn(t, conn, stc)
	assert.True(t, fired(trigg))

	testConnectionTxn(t, conn, stc)
	assert.False(t, fired(trigg))

	stc.Connections = []stitch.Connection{
		{From: "a", To: "a", MinPort: 90, MaxPort: 90},
	}
	testConnectionTxn(t, conn, stc)
	assert.True(t, fired(trigg))

	testConnectionTxn(t, conn, stc)
	assert.False(t, fired(trigg))

	stc.Connections = []stitch.Connection{
		{From: "b", To: "a", MinPort: 90, MaxPort: 90},
		{From: "b", To: "c", MinPort: 90, MaxPort: 90},
		{From: "b", To: "a", MinPort: 100, MaxPort: 100},
		{From: "c", To: "a", MinPort: 101, MaxPort: 101},
	}
	testConnectionTxn(t, conn, stc)
	assert.True(t, fired(trigg))

	testConnectionTxn(t, conn, stc)
	assert.False(t, fired(trigg))

	stc.Connections = nil
	testConnectionTxn(t, conn, stc)
	assert.True(t, fired(trigg))

	testConnectionTxn(t, conn, stc)
	assert.False(t, fired(trigg))
}

func testConnectionTxn(t *testing.T, conn db.Conn, stc stitch.Stitch) {
	var connections []db.Connection
	conn.Txn(db.AllTables...).Run(func(view db.Database) error {
		updatePolicy(view, stc.String())
		connections = view.SelectFromConnection(nil)
		return nil
	})

	exp := stc.Connections
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

		assert.True(t, found)
	}

	assert.Empty(t, connections)
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
	checkPlacement := func(stc stitch.Stitch, exp ...db.Placement) {
		placements := map[db.Placement]struct{}{}
		conn.Txn(db.AllTables...).Run(func(view db.Database) error {
			updatePolicy(view, stc.String())
			res := view.SelectFromPlacement(nil)

			// Set the ID to 0 so that we can use reflect.DeepEqual.
			for _, p := range res {
				p.ID = 0
				placements[p] = struct{}{}
			}

			return nil
		})

		assert.Equal(t, len(exp), len(placements))
		for _, p := range exp {
			_, ok := placements[p]
			assert.True(t, ok)
		}
	}

	stc := stitch.Stitch{
		Containers: []stitch.Container{
			{
				ID:    "1a84f87eebbf7dc7edda83a38f34a49f2116240b",
				Image: stitch.Image{Name: "foo"},
			},
			{
				ID:    "1806739e57b7678db83f0a5c6c63b16325c54242",
				Image: stitch.Image{Name: "bar"},
			},
			{
				ID:    "af771c5f8cd87c550263b011541cbf6a14051976",
				Image: stitch.Image{Name: "bar"},
			},
		},
		Labels: []stitch.Label{
			{
				Name: "foo",
				IDs: []string{
					"1a84f87eebbf7dc7edda83a38f34a49f2116240b",
				},
			},
			{
				Name: "bar",
				IDs: []string{
					"1806739e57b7678db83f0a5c6c63b16325c54242",
				},
			},
			{
				Name: "baz",
				IDs: []string{
					"af771c5f8cd87c550263b011541cbf6a14051976",
				},
			},
		},
		Placements: []stitch.Placement{
			{
				TargetLabel: "bar",
				Exclusive:   true,
				OtherLabel:  "foo",
			},
		},
	}

	// Create an exclusive placement.
	checkPlacement(stc,
		db.Placement{
			TargetLabel: "bar",
			Exclusive:   true,
			OtherLabel:  "foo",
		},
	)

	// Change the placement from "exclusive" to "on".
	stc.Placements = []stitch.Placement{
		{TargetLabel: "bar", Exclusive: false, OtherLabel: "foo"},
	}
	checkPlacement(stc,
		db.Placement{
			TargetLabel: "bar",
			Exclusive:   false,
			OtherLabel:  "foo",
		},
	)

	// Add another placement constraint.
	stc.Placements = []stitch.Placement{
		{TargetLabel: "bar", Exclusive: false, OtherLabel: "foo"},
		{TargetLabel: "bar", Exclusive: true, OtherLabel: "bar"},
	}
	checkPlacement(stc,
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
	stc.Placements = []stitch.Placement{
		{TargetLabel: "foo", Exclusive: false, Size: "m4.large"},
	}
	checkPlacement(stc,
		db.Placement{
			TargetLabel: "foo",
			Exclusive:   false,
			Size:        "m4.large",
		},
	)

	// XXX: Port placement belongs in Stitch unit tests.
	// Port placement
	stc.Placements = nil
	stc.Connections = []stitch.Connection{
		{From: stitch.PublicInternetLabel, To: "foo", MinPort: 80, MaxPort: 80},
		{From: stitch.PublicInternetLabel, To: "foo", MinPort: 81, MaxPort: 81},
	}
	checkPlacement(stc,
		db.Placement{
			TargetLabel: "foo",
			Exclusive:   true,
			OtherLabel:  "foo",
		},
	)

	stc.Connections = []stitch.Connection{
		{From: stitch.PublicInternetLabel, To: "foo", MinPort: 80, MaxPort: 80},
		{From: stitch.PublicInternetLabel, To: "bar", MinPort: 80, MaxPort: 80},
		{From: stitch.PublicInternetLabel, To: "bar", MinPort: 81, MaxPort: 81},
		{From: stitch.PublicInternetLabel, To: "baz", MinPort: 81, MaxPort: 81},
	}
	checkPlacement(stc,
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

func checkImage(t *testing.T, conn db.Conn, stc stitch.Stitch, exp ...db.Image) {
	var images []db.Image
	conn.Txn(db.AllTables...).Run(func(view db.Database) error {
		updatePolicy(view, stc.String())
		images = view.SelectFromImage(nil)
		return nil
	})

	key := func(intf interface{}) interface{} {
		im := intf.(db.Image)
		im.ID = 0
		return im
	}
	_, lonelyLeft, lonelyRight := join.HashJoin(
		db.ImageSlice(images), db.ImageSlice(exp), key, key)
	assert.Empty(t, lonelyLeft, "unexpected images")
	assert.Empty(t, lonelyRight, "missing images")
}

func TestImageTxn(t *testing.T) {
	t.Parallel()

	// Regular image that isn't built by Quilt.
	checkImage(t, db.New(), stitch.Stitch{
		Containers: []stitch.Container{
			{
				ID:    "475c40d6070969839ba0f88f7a9bd0cc7936aa30",
				Image: stitch.Image{Name: "image"},
			},
		},
		Labels: []stitch.Label{
			{
				Name: "foo",
				IDs: []string{
					"475c40d6070969839ba0f88f7a9bd0cc7936aa30",
				},
			},
		},
	})

	conn := db.New()
	checkImage(t, conn, stitch.Stitch{
		Containers: []stitch.Container{
			{
				ID:    "96189e4ea36c80171fd842ccc4c3438d06061991",
				Image: stitch.Image{Name: "a", Dockerfile: "1"},
			},
			{
				ID:    "c51d206a1414f1fadf5020e5db35feee91410f79",
				Image: stitch.Image{Name: "a", Dockerfile: "1"},
			},
			{
				ID:    "ede1e03efba48e66be3e51aabe03ec77d9f9def9",
				Image: stitch.Image{Name: "b", Dockerfile: "1"},
			},
			{
				ID:    "133c61c61ef4b49ea26717efe0f0468d455fd317",
				Image: stitch.Image{Name: "c"},
			},
		},
		Labels: []stitch.Label{
			{
				Name: "foo",
				IDs: []string{
					"96189e4ea36c80171fd842ccc4c3438d06061991",
				},
			},
			{
				Name: "foo2",
				IDs: []string{
					"c51d206a1414f1fadf5020e5db35feee91410f79",
				},
			},
			{
				Name: "foo3",
				IDs: []string{
					"ede1e03efba48e66be3e51aabe03ec77d9f9def9",
				},
			},
			{
				Name: "foo4",
				IDs: []string{
					"133c61c61ef4b49ea26717efe0f0468d455fd317",
				},
			},
		},
	},
		db.Image{
			Name:       "a",
			Dockerfile: "1",
		},
		db.Image{
			Name:       "b",
			Dockerfile: "1",
		},
	)

	// Ensure existing images are preserved.
	conn.Txn(db.ImageTable).Run(func(view db.Database) error {
		img := view.SelectFromImage(func(img db.Image) bool {
			return img.Name == "a"
		})[0]
		img.DockerID = "id"
		view.Commit(img)
		return nil
	})
	checkImage(t, conn, stitch.Stitch{
		Containers: []stitch.Container{
			{
				ID:    "96189e4ea36c80171fd842ccc4c3438d06061991",
				Image: stitch.Image{Name: "a", Dockerfile: "1"},
			},
			{
				ID:    "18c2c81fb48a2a481af58ba5ad6da0e2b244f060",
				Image: stitch.Image{Name: "b", Dockerfile: "2"},
			},
		},
		Labels: []stitch.Label{
			{
				Name: "foo",
				IDs: []string{
					"96189e4ea36c80171fd842ccc4c3438d06061991",
				},
			}, {
				Name: "foo2",
				IDs: []string{
					"18c2c81fb48a2a481af58ba5ad6da0e2b244f060",
				},
			},
		},
	},
		db.Image{
			Name:       "a",
			Dockerfile: "1",
			DockerID:   "id",
		},
		db.Image{
			Name:       "b",
			Dockerfile: "2",
		},
	)
}
