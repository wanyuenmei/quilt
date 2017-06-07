package command

import (
	"flag"

	"github.com/quilt/quilt/api"
)

type commonFlags struct {
	host string
}

func (cf *commonFlags) InstallFlags(flags *flag.FlagSet) {
	flags.StringVar(&cf.host, "H", api.DefaultSocket, "the host to connect to")
}

func (cf *commonFlags) BeforeRun() error {
	return nil
}

func (cf *commonFlags) AfterRun() error {
	return nil
}
