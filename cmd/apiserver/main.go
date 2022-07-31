package main

import (
	"flag"

	"github.com/BurntSushi/toml"

	"github.com/ducesoft/overlord/pkg/log"
	"github.com/ducesoft/overlord/platform/api/model"
	"github.com/ducesoft/overlord/platform/api/server"
	"github.com/ducesoft/overlord/platform/api/service"
	"github.com/ducesoft/overlord/version"
)

var (
	confPath string
)

func main() {
	flag.StringVar(&confPath, "conf", "conf.toml", "scheduler conf")
	flag.Parse()

	if version.ShowVersion() {
		return
	}

	conf := new(model.ServerConfig)
	_, err := toml.DecodeFile(confPath, &conf)
	if err != nil {
		panic(err)
	}
	if log.Init(conf.Config) {
		defer log.Close()
	}
	svc := service.New(conf)
	server.Run(conf, svc)
}
