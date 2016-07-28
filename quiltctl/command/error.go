package command

import (
	"fmt"
)

// DaemonResponseError represents when the Quilt daemon responds with an error.
type DaemonResponseError struct {
	responseError error
}

func (err DaemonResponseError) Error() string {
	return fmt.Sprintf("Bad response from the Quilt daemon: %s.",
		err.responseError.Error())
}

// DaemonConnectError represents when we are unable to connect to the Quilt daemon.
type DaemonConnectError struct {
	host         string
	connectError error
}

func (err DaemonConnectError) Error() string {
	return fmt.Sprintf("Unable to connect to the Quilt daemon at %s: %s.",
		err.host, err.connectError.Error())
}
