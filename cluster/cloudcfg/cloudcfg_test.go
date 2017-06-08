package cloudcfg

import (
	"testing"

	"github.com/quilt/quilt/db"

	log "github.com/Sirupsen/logrus"
)

func TestCloudConfig(t *testing.T) {
	cfgTemplate = "({{.QuiltImage}}) ({{.SSHKeys}}) ({{.UbuntuVersion}}) " +
		"({{.Role}}) ({{.LogLevel}})"

	log.SetLevel(log.InfoLevel)
	ver = "master"
	res := Ubuntu([]string{"a", "b"}, db.Master)
	exp := "(quilt/quilt:master) (a\nb) (xenial) (Master) (info)"
	if res != exp {
		t.Errorf("res: %s\nexp: %s", res, exp)
	}

	log.SetLevel(log.DebugLevel)
	ver = "1.2.3"
	res = Ubuntu([]string{"a", "b"}, db.Worker)
	exp = "(quilt/quilt:1.2.3) (a\nb) (xenial) (Worker) (debug)"
	if res != exp {
		t.Errorf("res: %s\nexp: %s", res, exp)
	}
}
