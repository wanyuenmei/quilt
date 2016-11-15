package command

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/robertkrimen/otto"

	"github.com/NetSys/quilt/stitch"
)

// Run contains the options for running Stitches.
type Run struct {
	stitch string

	common *commonFlags
}

// NewRunCommand creates a new Run command instance.
func NewRunCommand() *Run {
	return &Run{
		common: &commonFlags{},
	}
}

// InstallFlags sets up parsing for command line flags.
func (rCmd *Run) InstallFlags(flags *flag.FlagSet) {
	rCmd.common.InstallFlags(flags)

	flags.StringVar(&rCmd.stitch, "stitch", "", "the stitch to run")

	flags.Usage = func() {
		fmt.Println("usage: quilt run [-H=<daemon_host>] " +
			"[-stitch=<stitch>] <stitch>")
		fmt.Println("`run` compiles the provided stitch, and sends the " +
			"result to the Quilt daemon to be executed.")
		flags.PrintDefaults()
	}
}

// Parse parses the command line arguments for the run command.
func (rCmd *Run) Parse(args []string) error {
	if rCmd.stitch == "" {
		if len(args) == 0 {
			return errors.New("no spec specified")
		}
		rCmd.stitch = args[0]
	}

	return nil
}

// Run starts the run for the provided Stitch.
func (rCmd *Run) Run() int {
	c, err := getClient(rCmd.common.host)
	if err != nil {
		log.Error(err)
		return 1
	}
	defer c.Close()

	stitchPath := rCmd.stitch
	compiled, err := stitch.FromFile(stitchPath, stitch.DefaultImportGetter)
	if err != nil && os.IsNotExist(err) && !filepath.IsAbs(stitchPath) {
		// Automatically add the ".js" file suffix if it's not provided.
		if !strings.HasSuffix(stitchPath, ".js") {
			stitchPath += ".js"
		}
		compiled, err = stitch.FromFile(
			filepath.Join(stitch.GetQuiltPath(), stitchPath),
			stitch.DefaultImportGetter)
	}
	if err != nil {
		// Print the stacktrace if it's an Otto error.
		if ottoError, ok := err.(*otto.Error); ok {
			log.Error(ottoError.String())
		} else {
			log.Error(err)
		}
		return 1
	}

	err = c.Deploy(compiled.String())
	if err != nil {
		log.WithError(err).Error("Unable to start run.")
		return 1
	}

	fmt.Println("Successfully started run.")
	return 0
}
