//go:generate protoc ./minion/pb/pb.proto --go_out=plugins=grpc:.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	l_mod "log"
	"os"
	"path/filepath"
	"strings"
	"text/scanner"
	"time"

	"github.com/NetSys/quilt/api"
	"github.com/NetSys/quilt/api/server"
	"github.com/NetSys/quilt/cluster"
	"github.com/NetSys/quilt/db"
	"github.com/NetSys/quilt/engine"
	"github.com/NetSys/quilt/minion"
	"github.com/NetSys/quilt/quiltctl"
	"github.com/NetSys/quilt/stitch"
	"github.com/NetSys/quilt/util"

	"google.golang.org/grpc/grpclog"

	log "github.com/Sirupsen/logrus"
)

func main() {
	flag.Usage = func() {
		fmt.Println("Usage: quilt " +
			"[-log-level=<level> | -l=<level> | -H=<listen_address>] " +
			"[inspect <stitch> | run <stitch> | minion | " +
			"stop <namespace> | get <import_path> | " +
			"machines | containers]")
		fmt.Println("\nWhen provided a stitch, quilt takes responsibility\n" +
			"for deploying it as specified.  Alternatively, quilt may be\n" +
			"instructed to stop all deployments in a given namespace,\n" +
			"or the default namespace if none is provided.\n")
		flag.PrintDefaults()
		fmt.Println("        Valid logger levels are:\n" +
			"            debug, info, warn, error, fatal or panic.")
	}

	var logLevel = flag.String("log-level", "info", "level to set logger to")
	flag.StringVar(logLevel, "l", "info", "level to set logger to")
	var lAddr = flag.String("H", api.DefaultSocket,
		"Socket to listen for API requests on.")
	flag.Parse()

	level, err := parseLogLevel(*logLevel)
	if err != nil {
		fmt.Println(err)
		usage()
	}
	log.SetLevel(level)
	log.SetFormatter(util.Formatter{})

	// GRPC spews a lot of useless log messages so we tell to eat its logs, unless
	// we are in debug mode
	grpclog.SetLogger(l_mod.New(ioutil.Discard, "", 0))
	if level == log.DebugLevel {
		grpclog.SetLogger(log.StandardLogger())
	}

	conn := db.New()
	nArgs := len(flag.Args())
	if nArgs < 1 || (flag.Arg(0) != "minion" && nArgs < 2) {
		usage()
	}

	subcommand := flag.Arg(0)
	switch {
	case subcommand == "run":
		go configLoop(conn, flag.Arg(1))
	case subcommand == "stop":
		stop(conn, flag.Arg(1))
	case subcommand == "minion":
		minion.Run()
		return
	case quiltctl.HasSubcommand(subcommand):
		quiltctl.Run(flag.Args())
		return
	default:
		usage()
	}

	go server.Run(conn, *lAddr)
	cluster.Run(conn)
}

func stop(conn db.Conn, namespace string) {
	specStr := "(define AdminACL (list))"
	if namespace != "" {
		specStr += fmt.Sprintf(` (define Namespace "%s")`, namespace)
	}

	var sc scanner.Scanner
	spec, err := stitch.New(*sc.Init(strings.NewReader(specStr)), "", false)
	if err != nil {
		panic(err)
	}

	err = engine.UpdatePolicy(conn, spec)
	if err != nil {
		panic(err)
	}
}

func configLoop(conn db.Conn, stitchPath string) {
	tick := time.Tick(5 * time.Second)
	for {
		if err := updateConfig(conn, stitchPath); err != nil {
			log.WithError(err).Warn("Failed to update configuration.")
		}
		<-tick
	}
}

func usage() {
	flag.Usage()
	os.Exit(1)
}

const quiltPath = "QUILT_PATH"

func updateConfig(conn db.Conn, configPath string) error {
	pathStr, _ := os.LookupEnv(quiltPath)
	if pathStr == "" {
		pathStr = stitch.GetQuiltPath()
	}

	f, err := util.Open(configPath)
	if err != nil {
		f, err = util.Open(filepath.Join(pathStr, configPath))
		if err != nil {
			return err
		}
	}

	defer f.Close()

	sc := scanner.Scanner{
		Position: scanner.Position{
			Filename: configPath,
		},
	}

	spec, err := stitch.New(*sc.Init(bufio.NewReader(f)), pathStr, false)
	if err != nil {
		return err
	}

	return engine.UpdatePolicy(conn, spec)
}

// parseLogLevel returns the log.Level type corresponding to the given string
// (case insensitive).
// If no such matching string is found, it returns log.InfoLevel (default) and an error.
func parseLogLevel(logLevel string) (log.Level, error) {
	logLevel = strings.ToLower(logLevel)
	switch logLevel {
	case "debug":
		return log.DebugLevel, nil
	case "info":
		return log.InfoLevel, nil
	case "warn":
		return log.WarnLevel, nil
	case "error":
		return log.ErrorLevel, nil
	case "fatal":
		return log.FatalLevel, nil
	case "panic":
		return log.PanicLevel, nil
	}
	return log.InfoLevel, fmt.Errorf("bad log level: '%v'", logLevel)
}
