package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"myself_docker/cgroups/subsystems"
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
		&cli.StringFlag{
			Name:  "mem", // 限制进程内存使用量，为了避免和 stress 命令的 -m 参数冲突 这里使用 -mem,到时候可以看下解决冲突的方法
			Usage: "memory limit,e.g.: -mem 100m",
		},
		&cli.StringFlag{
			Name:  "cpu",
			Usage: "cpu quota,e.g.: -cpu 100", // 限制进程 cpu 使用率
		},
		&cli.StringFlag{
			Name:  "cpuset",
			Usage: "cpuset limit,e.g.: -cpuset 2,4", // 限制进程 cpu 使用率
		},
	},
	Action: func(context *cli.Context) error {
		if context.Args().Len() < 1 {
			return fmt.Errorf(" Missing container command ")
		}
		var cmdArray []string
		for _, arg := range context.Args() {
			cmdArray = append(cmdArray, arg)
		}
		tty := context.Bool("it")
		resConf := &subsystems.ResourceConfig{
			MemoryLimit: context.String("mem"),
			CpuSet:      context.String("cpuset"),
			CpuCfsQuota: context.Int("cpu"),
		}
		log.Info("resConf:", resConf)
		Run(tty, cmdArray, resConf)
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
