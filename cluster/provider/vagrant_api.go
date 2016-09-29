package provider

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/NetSys/quilt/util"
	log "github.com/Sirupsen/logrus"
	homedir "github.com/mitchellh/go-homedir"
)

var vagrantCmd = "vagrant"
var shCmd = "sh"

type vagrantAPI struct{}

func newVagrantAPI() vagrantAPI {
	vagrant := vagrantAPI{}
	return vagrant
}

func (api vagrantAPI) Init(cloudConfig string, size string, id string) error {
	vdir, err := api.VagrantDir()
	if err != nil {
		return err
	}
	if _, err := os.Stat(vdir); os.IsNotExist(err) {
		os.Mkdir(vdir, os.ModeDir|os.ModePerm)
	}
	path := vdir + id
	os.Mkdir(path, os.ModeDir|os.ModePerm)

	_, stderr, err := api.Shell(id, `vagrant --machine-readable init coreos-beta`)
	if err != nil {
		log.Errorf("Failed to initialize Vagrant environment: %s", string(stderr))
		api.Destroy(id)
		return errors.New("unable to init machine")
	}

	err = util.WriteFile(path+"/user-data", []byte(cloudConfig), 0644)
	if err != nil {
		api.Destroy(id)
		return err
	}

	vagrant := vagrantFile()
	err = util.WriteFile(path+"/vagrantFile", []byte(vagrant), 0644)
	if err != nil {
		api.Destroy(id)
		return err
	}

	err = util.WriteFile(path+"/size", []byte(size), 0644)
	if err != nil {
		api.Destroy(id)
		return err
	}

	return nil
}

func (api vagrantAPI) Up(id string) error {
	_, stderr, err := api.Shell(id, `vagrant --machine-readable up`)
	if err != nil {
		log.Errorf("Failed to start Vagrant machine: %s", string(stderr))
		return errors.New("unable to check machine status")
	}
	return nil
}

func (api vagrantAPI) Destroy(id string) error {
	_, stderr, err := api.Shell(id,
		`vagrant --machine-readable destroy -f; cd ../; rm -rf %s`)
	if err != nil {
		log.Errorf("Failed to destroy Vagrant machine: %s", string(stderr))
		return errors.New("unable to destroy machine")
	}
	return nil
}

func (api vagrantAPI) PublicIP(id string) (string, error) {
	ip, stderr, err := api.Shell(id,
		`vagrant ssh -c "ip -f inet addr show enp0s8 | grep -Po 'inet \K[\d.]+'"`)
	if err != nil {
		log.Errorf("Failed to parse Vagrant machine IP: %s", string(stderr))
		return "", err
	}
	return strings.TrimSuffix(string(ip), "\n"), nil
}

func (api vagrantAPI) Status(id string) (string, error) {
	output, stderr, err := api.Shell(id, `vagrant --machine-readable status`)
	if err != nil {
		log.Errorf("Failed to retrieve Vagrant machine status: %s",
			string(stderr))
		return "", errors.New("unable to retrieve machine status")
	}
	lines := bytes.Split(output, []byte("\n"))
	for _, line := range lines {
		words := strings.Split(string(line[:]), ",")
		if len(words) >= 4 {
			if strings.Compare(words[2], "state") == 0 {
				return words[3], nil
			}
		}
	}
	return "", nil
}

func (api vagrantAPI) List() ([]string, error) {
	subdirs := []string{}
	vdir, err := api.VagrantDir()
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(vdir); os.IsNotExist(err) {
		return subdirs, nil
	}

	files, err := ioutil.ReadDir(vdir)
	if err != nil {
		return subdirs, err
	}
	for _, file := range files {
		subdirs = append(subdirs, file.Name())
	}
	return subdirs, nil
}

func (api vagrantAPI) AddBox(name string, provider string) error {
	/* Adding a box fails if it already exists, hence the check. */
	exists, err := api.ContainsBox(name)
	if err == nil && exists {
		return nil
	}
	err = exec.Command(vagrantCmd, []string{"--machine-readable", "box", "add",
		"--provider", provider, name}...).Run()
	if err != nil {
		return errors.New("unable to add box")
	}
	return nil
}

func (api vagrantAPI) ContainsBox(name string) (bool, error) {
	output, err := exec.Command(vagrantCmd, []string{"--machine-readable", "box",
		"list"}...).Output()
	if err != nil {
		return false, errors.New("unable to list machines")
	}
	lines := bytes.Split(output, []byte("\n"))
	for _, line := range lines {
		words := strings.Split(string(line[:]), ",")
		if words[len(words)-1] == name {
			return true, nil
		}
	}
	return false, nil
}

func (api vagrantAPI) Shell(id string, commands string) ([]byte, []byte, error) {
	chdir := `(cd %s; `
	vdir, err := api.VagrantDir()
	if err != nil {
		return nil, nil, err
	}
	chdir = fmt.Sprintf(chdir, vdir+id)
	shellCommand := chdir + strings.Replace(commands, "%s", id, -1) + ")"

	var outbuf, errbuf bytes.Buffer
	cmd := exec.Command(shCmd, []string{"-c", shellCommand}...)
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf
	err = cmd.Run()

	return outbuf.Bytes(), errbuf.Bytes(), err
}

func (api vagrantAPI) VagrantDir() (string, error) {
	dir, err := homedir.Dir()
	if err != nil {
		return "", err
	}
	vagrantDir := dir + "/.vagrant/"
	return vagrantDir, nil
}

func (api vagrantAPI) Size(id string) string {
	size, _, err := api.Shell(id, "cat size")
	if err != nil {
		return ""
	}
	return string(size)
}

func vagrantFile() string {
	vagrantfile := `CLOUD_CONFIG_PATH = File.join(File.dirname(__FILE__), "user-data")
SIZE_PATH = File.join(File.dirname(__FILE__), "size")
Vagrant.require_version ">= 1.6.0"

size = File.open(SIZE_PATH).read.strip.split(",")
Vagrant.configure(2) do |config|
  config.vm.box = "boxcutter/ubuntu1604"

  config.vm.network "private_network", type: "dhcp"

  ram=(size[0].to_f*1024).to_i
  cpus=size[1]
  config.vm.provider "virtualbox" do |v|
    v.memory = ram
    v.cpus = cpus
  end

  if File.exist?(CLOUD_CONFIG_PATH)
    config.vm.provision "shell", path: "#{CLOUD_CONFIG_PATH}"
  end
end
`
	return vagrantfile
}

// VagrantCreateSize creates an encoded string representing the amount of RAM
// and number of CPUs for an instance.
func (api vagrantAPI) CreateSize(ram, cpu float64) string {
	if ram < 1 {
		ram = 1
	}
	if cpu < 1 {
		cpu = 1
	}
	return fmt.Sprintf("%g,%g", ram, cpu)
}
