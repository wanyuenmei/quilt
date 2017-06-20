package cloudcfg

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/quilt/quilt/db"
	"github.com/quilt/quilt/version"

	log "github.com/Sirupsen/logrus"
)

const (
	quiltImage = "quilt/quilt"
)

// Allow mocking out for the unit tests.
var ver = version.Version

// Ubuntu generates a cloud config file for the Ubuntu operating system with the
// corresponding `version`.
func Ubuntu(opts Options) string {
	t := template.Must(template.New("cloudConfig").Parse(cfgTemplate))

	img := fmt.Sprintf("%s:%s", quiltImage, ver)

	var cloudConfigBytes bytes.Buffer
	err := t.Execute(&cloudConfigBytes, struct {
		QuiltImage    string
		UbuntuVersion string
		SSHKeys       string
		LogLevel      string
		MinionOpts    string
	}{
		QuiltImage:    img,
		UbuntuVersion: "xenial",
		SSHKeys:       strings.Join(opts.SSHKeys, "\n"),
		LogLevel:      log.GetLevel().String(),
		MinionOpts:    opts.MinionOpts.String(),
	})
	if err != nil {
		panic(err)
	}

	return cloudConfigBytes.String()
}

// Options defines configuration for the cloud config.
type Options struct {
	SSHKeys    []string
	MinionOpts MinionOptions
}

// MinionOptions defines the command line flags the minion should be invoked with.
type MinionOptions struct {
	Role db.Role
}

func (opts MinionOptions) String() string {
	optsMap := map[string]string{
		"role": string(opts.Role),
	}

	var optsList []string
	for name, val := range optsMap {
		if val != "" {
			optsList = append(optsList, fmt.Sprintf("--%s %q", name, val))
		}
	}

	return strings.Join(optsList, " ")
}
