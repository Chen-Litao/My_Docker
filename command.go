package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"myself_docker/cgroups/subsystems"
	"myself_docker/container"
)

var runCommand = cli.Command{
	Name:  "run",
	Usage: "启动一个容器 gocker run -it [command]",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "it",
			Usage: "是否启用命令行交互模式",
		},
		cli.StringFlag{
			Name:  "mem", // 限制进程内存使用量，为了避免和 stress 命令的 -m 参数冲突 这里使用 -mem,到时候可以看下解决冲突的方法
			Usage: "memory limit,e.g.: -mem 100m",
		},
		cli.StringFlag{
			Name:  "cpu",
			Usage: "cpu quota,e.g.: -cpu 100", // 限制进程 cpu 使用率
		},
		cli.StringFlag{
			Name:  "cpuset",
			Usage: "cpuset limit,e.g.: -cpuset 2,4", // 限制进程 cpu 使用率
		},
		cli.StringFlag{
			Name:  "v",
			Usage: "volume limit,e.g.: -v /source/config:/target/config,", // 限制进程 cpu 使用率
		},
		cli.BoolFlag{
			Name:  "d",
			Usage: "detach container,run background",
		},
		cli.StringFlag{
			Name:  "name",
			Usage: "container name",
		},
	},
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf(" Missing container command ")
		}
		//获取的是非flag后面的值
		var cmdArray []string
		for _, arg := range context.Args() {
			cmdArray = append(cmdArray, arg)
		}
		tty := context.Bool("it")
		detach := context.Bool("d")
		if tty && detach {
			return fmt.Errorf("it and d flag can not both provided")
		}
		if !detach { // 如果不是指定后台运行，就默认前台运行
			tty = true
		}
		log.Infof("createTty %v", tty)
		//flag后面的所有参数会被记录到这里
		resConf := &subsystems.ResourceConfig{
			MemoryLimit: context.String("mem"),
			CpuSet:      context.String("cpuset"),
			CpuCfsQuota: context.Int("cpu"),
		}
		log.Info("resConf:", resConf)
		volume := context.String("v")
		containerName := context.String("name")
		Run(tty, cmdArray, resConf, volume, containerName)
		return nil
	},
}

var commitCommand = cli.Command{
	Name:  "commit",
	Usage: "commit container to image",
	Action: func(context *cli.Context) error {
		log.Infof("commit come on")
		if len(context.Args()) < 1 {
			return fmt.Errorf("missing image name")
		}
		imageName := context.Args().Get(0)
		log.Infof("command %s", imageName)
		err := container.CommitContainer(imageName)
		return err
	},
}

var initCommand = cli.Command{
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

var listCommand = cli.Command{
	Name:  "ps",
	Usage: "list all the containers",
	Action: func(context *cli.Context) error {
		ListContainers()
		return nil
	},
}
