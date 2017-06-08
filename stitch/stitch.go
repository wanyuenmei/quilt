package stitch

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
)

// A Stitch is an abstract representation of the policy language.
type Stitch struct {
	Containers  []Container  `json:",omitempty"`
	Labels      []Label      `json:",omitempty"`
	Connections []Connection `json:",omitempty"`
	Placements  []Placement  `json:",omitempty"`
	Machines    []Machine    `json:",omitempty"`

	AdminACL  []string `json:",omitempty"`
	MaxPrice  float64  `json:",omitempty"`
	Namespace string   `json:",omitempty"`

	Invariants []invariant `json:",omitempty"`
}

// A Placement constraint guides where containers may be scheduled, either relative to
// the labels of other containers, or the machine the container will run on.
type Placement struct {
	TargetLabel string `json:",omitempty"`

	Exclusive bool `json:",omitempty"`

	// Label Constraint
	OtherLabel string `json:",omitempty"`

	// Machine Constraints
	Provider   string `json:",omitempty"`
	Size       string `json:",omitempty"`
	Region     string `json:",omitempty"`
	FloatingIP string `json:",omitempty"`
}

// An Image represents a Docker image that can be run. If the Dockerfile is non-empty,
// the image should be built and hosted by Quilt.
type Image struct {
	Name       string `json:",omitempty"`
	Dockerfile string `json:",omitempty"`
}

// A Container may be instantiated in the stitch and queried by users.
type Container struct {
	ID                string            `json:",omitempty"`
	Image             Image             `json:",omitempty"`
	Command           []string          `json:",omitempty"`
	Env               map[string]string `json:",omitempty"`
	FilepathToContent map[string]string `json:",omitempty"`
	Hostname          string            `json:",omitempty"`
}

// A Label represents a logical group of containers.
type Label struct {
	Name        string   `json:",omitempty"`
	IDs         []string `json:",omitempty"`
	Annotations []string `json:",omitempty"`
}

// A Connection allows containers implementing the From label to speak to containers
// implementing the To label in ports in the range [MinPort, MaxPort]
type Connection struct {
	From    string `json:",omitempty"`
	To      string `json:",omitempty"`
	MinPort int    `json:",omitempty"`
	MaxPort int    `json:",omitempty"`
}

// A ConnectionSlice allows for slices of Collections to be used in joins
type ConnectionSlice []Connection

// A Machine specifies the type of VM that should be booted.
type Machine struct {
	ID          string   `json:",omitempty"`
	Provider    string   `json:",omitempty"`
	Role        string   `json:",omitempty"`
	Size        string   `json:",omitempty"`
	CPU         Range    `json:",omitempty"`
	RAM         Range    `json:",omitempty"`
	DiskSize    int      `json:",omitempty"`
	Region      string   `json:",omitempty"`
	SSHKeys     []string `json:",omitempty"`
	FloatingIP  string   `json:",omitempty"`
	Preemptible bool     `json:",omitempty"`
}

// A Range defines a range of acceptable values for a Machine attribute
type Range struct {
	Min float64 `json:",omitempty"`
	Max float64 `json:",omitempty"`
}

// PublicInternetLabel is a magic label that allows connections to or from the public
// network.
const PublicInternetLabel = "public"

// Accepts returns true if `x` is within the range specified by `stitchr` (include),
// or if no max is specified and `x` is larger than `stitchr.min`.
func (stitchr Range) Accepts(x float64) bool {
	return stitchr.Min <= x && (stitchr.Max == 0 || x <= stitchr.Max)
}

var lookPath = exec.LookPath

// FromFile gets a Stitch handle from a file on disk.
func FromFile(filename string) (Stitch, error) {
	if _, err := lookPath("node"); err != nil {
		return Stitch{}, errors.New(
			"failed to locate Node.js. Is it installed and in your PATH?")
	}

	stderr := bytes.NewBuffer(nil)
	cmd := exec.Command("node", "-p",
		fmt.Sprintf(
			`require("%s");
			JSON.stringify(global._quiltDeployment.toQuiltRepresentation());`,
			filename,
		),
	)
	cmd.Stderr = stderr
	out, err := cmd.Output()
	if err != nil {
		return Stitch{}, errors.New(stderr.String())
	}
	// If there wasn't an error, still print stderr, in case there were any
	// warnings or other non-fatal errors.
	fmt.Fprint(os.Stderr, stderr.String())

	return FromJSON(string(out))
}

// FromJSON gets a Stitch handle from the deployment representation.
func FromJSON(jsonStr string) (stc Stitch, err error) {
	err = json.Unmarshal([]byte(jsonStr), &stc)
	if err != nil {
		return Stitch{}, err
	}

	if len(stc.Invariants) == 0 {
		return stc, nil
	}

	graph, err := InitializeGraph(stc)
	if err != nil {
		return Stitch{}, err
	}

	if err := checkInvariants(graph, stc.Invariants); err != nil {
		return Stitch{}, err
	}

	return stc, nil
}

// String returns the Stitch in its deployment representation.
func (stitch Stitch) String() string {
	jsonBytes, err := json.Marshal(stitch)
	if err != nil {
		panic(err)
	}
	return string(jsonBytes)
}

// Get returns the value contained at the given index
func (cs ConnectionSlice) Get(ii int) interface{} {
	return cs[ii]
}

// Len returns the number of items in the slice
func (cs ConnectionSlice) Len() int {
	return len(cs)
}
