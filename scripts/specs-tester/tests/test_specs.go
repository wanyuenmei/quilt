package tests

import (
	"bytes"
	"errors"
	"os"
	"os/exec"

	log "github.com/Sirupsen/logrus"

	"github.com/quilt/quilt/stitch"
)

// workDir is the directory specs are placed during testing.
const workDir = "/tmp/quilt-spec-test"

func tryRunSpec(s spec) error {
	os.Mkdir(workDir, 0755)
	defer os.RemoveAll(workDir)
	os.Chdir(workDir)

	if err := run("git", "clone", s.repo, "."); err != nil {
		return err
	}

	if err := run("npm", "install", "."); err != nil {
		return err
	}

	_, err := stitch.FromFile(s.path)
	return err
}

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	stderr := bytes.NewBuffer(nil)
	cmd.Stderr = stderr
	if cmd.Run() != nil {
		return errors.New(stderr.String())
	}
	return nil
}

type spec struct {
	repo, path string
}

// TestSpecs checks that the listed Quilt specs compile.
func TestSpecs() error {
	specs := []spec{
		{"https://github.com/quilt/tester", "./tests/100-logs/logs.js"},
		{"https://github.com/quilt/tester",
			"./tests/61-duplicate-cluster/duplicate-cluster.js"},
		{"https://github.com/quilt/tester",
			"./tests/60-duplicate-cluster-setup/duplicate-cluster-setup.js"},
		{"https://github.com/quilt/tester", "./tests/40-stop/stop.js"},
		{"https://github.com/quilt/tester", "./tests/30-mean/mean.js"},
		{"https://github.com/quilt/tester", "./tests/20-spark/spark.js"},
		{"https://github.com/quilt/tester", "./tests/15-bandwidth/bandwidth.js"},
		{"https://github.com/quilt/tester", "./tests/10-network/network.js"},
		{"https://github.com/quilt/tester",
			"./tests/outbound-public/outbound-public.js"},
		{"https://github.com/quilt/tester",
			"./tests/inbound-public/inbound-public.js"},
		{"https://github.com/quilt/tester",
			"./tests/elasticsearch/elasticsearch.js"},
		{"https://github.com/quilt/tester",
			"./tests/build-dockerfile/build-dockerfile.js"},
		{"https://github.com/quilt/tester", "./tests/etcd/etcd.js"},
		{"https://github.com/quilt/tester", "./tests/zookeeper/zookeeper.js"},

		{"https://github.com/quilt/nginx", "./main.js"},
		{"https://github.com/quilt/spark", "./sparkPI.js"},
		{"https://github.com/quilt/wordpress", "./wordpress-example.js"},
		{"https://github.com/quilt/etcd", "./etcd-example.js"},
		{"https://github.com/quilt/zookeeper", "./zookeeper-example.js"},
		{"https://github.com/quilt/redis", "./redis-example.js"},
		{"https://github.com/quilt/mean", "./example.js"},
		{"https://github.com/quilt/elasticsearch", "./main.js"},
		{"https://github.com/quilt/kibana", "./main.js"},
		{"https://github.com/quilt/django", "./django-example.js"},
		{"https://github.com/quilt/php-apache", "./main.js"},
		{"https://github.com/quilt/mongo", "./example.js"},
		{"https://github.com/quilt/tester", "./tester-runner-example.js"},
		{"https://github.com/quilt/lobsters", "./lobsters-example.js"},
		{"https://github.com/quilt/infrastructure", "./floating-ip.js"},
	}

	for _, s := range specs {
		log.Infof("Testing %s in %s", s.path, s.repo)
		if err := tryRunSpec(s); err != nil {
			return err
		}
	}
	return nil
}
