package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"myself_docker/cgroups/subsystems"
	"myself_docker/container"
	"os"
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
		imageName := cmdArray[0]
		cmdArray = cmdArray[1:]
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
		Run(tty, cmdArray, resConf, volume, containerName, imageName)
		return nil
	},
}

var commitCommand = cli.Command{
	Name:  "commit",
	Usage: "commit container to image,image,e.g. mydocker commit 123456789 myimage",
	Action: func(context *cli.Context) error {
		log.Infof("commit come on")
		if len(context.Args()) < 2 {
			return fmt.Errorf("missing image name")
		}
		containerID := context.Args().Get(0)
		imageName := context.Args().Get(1)
		log.Infof("command %s", imageName)
		err := container.CommitContainer(containerID, imageName)
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
		err := container.RunContainerInitProcess()
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
var logCommand = cli.Command{
	Name:  "logs",
	Usage: "print logs of a container",
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("please input your container name")
		}
		containerName := context.Args().Get(0)
		logContainer(containerName)
		return nil
	},
}
var execCommand = cli.Command{
	Name:  "exec",
	Usage: "exec a command into container",
	Action: func(context *cli.Context) error {
		if os.Getenv(EnvExecPid) != "" {
			log.Infof("pid callback pid %v", os.Getgid())
			return nil
		}
		if len(context.Args()) < 2 {
			return fmt.Errorf("missing container name or command")
		}
		containerName := context.Args().Get(0)
		// 将除了容器名之外的参数作为命令部分
		commandArray := context.Args().Tail()
		ExecContainer(containerName, commandArray)
		return nil
	},
}
var stopCommand = cli.Command{
	Name:  "stop",
	Usage: "stop a container,e.g. mydocker stop 1234567890",
	Action: func(context *cli.Context) error {
		// 期望输入是：mydocker stop 容器Id，如果没有指定参数直接打印错误
		if len(context.Args()) < 1 {
			return fmt.Errorf("missing container id")
		}
		containerName := context.Args().Get(0)
		stopContainer(containerName)
		return nil
	},
}

var removeCommand = cli.Command{
	Name:  "rm",
	Usage: "remove unused containers,e.g. mydocker rm 1234567890",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "f",
			Usage: "force delete running container",
		},
	},
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("missing container id")
		}
		containerId := context.Args().Get(0)
		force := context.Bool("f")
		removeContainer(containerId, force)
		return nil
	},
}
