package inspect

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/quilt/quilt/stitch"
)

func TestSlug(t *testing.T) {
	test := map[string]string{
		"slug.spec":       "slug",
		"a/b/c/slug.spec": "a/b/c/slug",
		"foo":             "err",
	}

	for inp, expect := range test {
		if sl, err := getSlug(inp); err != nil {
			if expect != "err" {
				t.Error(err)
			}
		} else if sl != expect {
			t.Error(sl)
		}
	}
}

// The expected graphviz graph returned by inspect when run on `testStitch`.
const expGraph = `strict digraph {
    subgraph cluster_0 {
        3c1a5738512a43c3122608ab32dbf9f84a14e5f9;
        54be1283e837c6e40ac79709aca8cdb8ec5f31f5;
        cb129f8a27df770b1dac70955c227a57bc5c4af6;
        public;
    }
    3c1a5738512a43c3122608ab32dbf9f84a14e5f9 -> cb129f8a27df770b1dac70955c227a57bc5c4af6
    54be1283e837c6e40ac79709aca8cdb8ec5f31f5 -> 3c1a5738512a43c3122608ab32dbf9f84a14e5f9
}`

func isGraphEqual(a, b string) bool {
	a = strings.Replace(a, "\n", "", -1)
	a = strings.Replace(a, " ", "", -1)
	b = strings.Replace(b, "\n", "", -1)
	b = strings.Replace(b, " ", "", -1)
	return a == b
}

func TestViz(t *testing.T) {
	t.Parallel()

	spec := stitch.Stitch{
		Containers: []stitch.Container{
			{
				ID:    "54be1283e837c6e40ac79709aca8cdb8ec5f31f5",
				Image: stitch.Image{Name: "ubuntu"},
			},
			{
				ID:    "3c1a5738512a43c3122608ab32dbf9f84a14e5f9",
				Image: stitch.Image{Name: "ubuntu"},
			},
			{
				ID:    "cb129f8a27df770b1dac70955c227a57bc5c4af6",
				Image: stitch.Image{Name: "ubuntu"},
			},
		},
		Labels: []stitch.Label{
			{
				Name: "a",
				IDs: []string{
					"54be1283e837c6e40ac79709aca8cdb8ec5f31f5",
				},
			},
			{
				Name: "b",
				IDs: []string{
					"3c1a5738512a43c3122608ab32dbf9f84a14e5f9",
				},
			},
			{
				Name: "c",
				IDs: []string{
					"cb129f8a27df770b1dac70955c227a57bc5c4af6",
				},
			},
		},
		Connections: []stitch.Connection{
			{From: "a", To: "b", MinPort: 22, MaxPort: 22},
			{From: "b", To: "c", MinPort: 22, MaxPort: 22},
		},
	}

	graph, err := stitch.InitializeGraph(spec)
	if err != nil {
		panic(err)
	}

	gv := makeGraphviz(graph)
	if !isGraphEqual(gv, expGraph) {
		t.Error(gv + "\n" + expGraph)
	}
}

func TestMainArgErr(t *testing.T) {
	t.Parallel()

	exitCode := Main([]string{"test.js"})
	assert.NotZero(t, exitCode)
}
