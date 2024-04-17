package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"myself_docker/container"
)

var RunCommand = &cli.Command{
	Name:  "run",
	Usage: "启动一个容器 gocker run -it [command]",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "it",
			Usage: "是否启用命令行交互模式",
		},
	},
	Action: func(context *cli.Context) error {
		if context.Args().Len() < 1 {
			return fmt.Errorf(" Missing container command ")
		}
		cmd := context.Args().Get(0)
		tty := context.Bool("it")
		Run(tty, cmd)
		return nil
	},
}

var InitCommand = &cli.Command{
	Name:  "init",
	Usage: "Init container process run user's process in container. Do not call it outside",
	Action: func(context *cli.Context) error {
		log.Infof("init come on")
		cmd := context.Args().Get(0)
		log.Infof("command %s", cmd)
		err := container.RunContainerInitProcess(cmd, nil)
		return err
	},
}
