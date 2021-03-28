// +build !windows

package main

import (
	"io"

	"github.com/sirupsen/logrus"
	u "github.com/docker/docker/utils"
)

func runDaemon(opts *daemonOptions) error {
	u.Info("enter cmd runDaemon")
	defer u.Duration(u.Track("cmd runDaemon"))
	
	daemonCli := NewDaemonCli()
	return daemonCli.start(opts)
}

func initLogging(_, stderr io.Writer) {
	logrus.SetOutput(stderr)
}
