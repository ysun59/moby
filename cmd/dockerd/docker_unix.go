// +build !windows

package main

import (
	"io"

	"github.com/sirupsen/logrus"
	u "github.com/YesZhen/superlog_go"
)

func runDaemon(opts *daemonOptions) error {
	u.Info("enter cmd runDaemon")
	defer u.LogEnd(u.LogBegin("cmd runDaemon"))
	
	daemonCli := NewDaemonCli()
	return daemonCli.start(opts)
}

func initLogging(_, stderr io.Writer) {
	logrus.SetOutput(stderr)
}
