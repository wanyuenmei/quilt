package command

import (
	"flag"

	"github.com/quilt/quilt/api"
	"github.com/quilt/quilt/api/client"
)

type connectionFlags struct {
	host string
}

func (cf *connectionFlags) InstallFlags(flags *flag.FlagSet) {
	flags.StringVar(&cf.host, "H", api.DefaultSocket, "the host to connect to")
}

type connectionHelper struct {
	client client.Client

	connectionFlags
}

func (ch *connectionHelper) BeforeRun() error {
	return ch.setupClient(client.New)
}

func (ch *connectionHelper) AfterRun() error {
	return ch.client.Close()
}

func (ch *connectionHelper) setupClient(getter client.Getter) (err error) {
	ch.client, err = getter(ch.host)
	return err
}
