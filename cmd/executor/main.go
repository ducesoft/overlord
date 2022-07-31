package main

import (
	"context"
	"flag"

	"github.com/ducesoft/overlord/pkg/log"
	"github.com/ducesoft/overlord/platform/mesos"
	"github.com/ducesoft/overlord/version"
)

func main() {
	flag.Parse()
	if version.ShowVersion() {
		return
	}

	ec := mesos.New()
	log.InitHandle(log.NewStdHandler())
	ec.Run(context.Background())
}
