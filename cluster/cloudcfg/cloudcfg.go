package cloudcfg

import (
	"bytes"
	"html/template"
	"strings"
)

const (
	quiltImage = "quilt/quilt:latest"
)

// Ubuntu generates a cloud config file for the Ubuntu operating system with the
// corresponding `version`.
func Ubuntu(keys []string, version string) string {
	t := template.Must(template.New("cloudConfig").Parse(cfgTemplate))

	var cloudConfigBytes bytes.Buffer
	err := t.Execute(&cloudConfigBytes, struct {
		QuiltImage    string
		UbuntuVersion string
		SSHKeys       string
	}{
		QuiltImage:    quiltImage,
		UbuntuVersion: version,
		SSHKeys:       strings.Join(keys, "\n"),
	})
	if err != nil {
		panic(err)
	}

	return cloudConfigBytes.String()
}
