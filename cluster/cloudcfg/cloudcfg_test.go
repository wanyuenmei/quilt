package cloudcfg

import "testing"

func TestCloudConfig(t *testing.T) {
	cfgTemplate = "({{.QuiltImage}}) ({{.SSHKeys}}) ({{.UbuntuVersion}})"

	res := Ubuntu([]string{"a", "b"}, "1")
	exp := "(quilt/quilt:latest) (a\nb) (1)"
	if res != exp {
		t.Errorf("res: %s\nexp: %s", res, exp)
	}
}
