package network

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"myself_docker/constant"
	"net"
	"os"
	"path"
	"path/filepath"
	"text/tabwriter"
)

var (
	defaultNetworkPath = "/var/lib/mydocker/network/network/"
	drivers            = map[string]Driver{}
)

// 加载网卡驱动到drivers里面，并创建网络相关文件
func init() {
	// 加载网络驱动
	var bridgeDriver = BridgeNetworkDriver{}
	drivers[bridgeDriver.Name()] = &bridgeDriver
	// 文件不存在则创建
	if _, err := os.Stat(defaultNetworkPath); err != nil {
		if !os.IsNotExist(err) {
			logrus.Errorf("check %s is exist failed,detail:%v", defaultNetworkPath, err)
			return
		}
		if err = os.MkdirAll(defaultNetworkPath, constant.Perm0644); err != nil {
			logrus.Errorf("create %s failed,detail:%v", defaultNetworkPath, err)
			return
		}
	}
}

func DeleteNetwork(networkName string) error {
	//读取相关数据
	networks, err := loadNetwork()
	if err != nil {
		return errors.WithMessage(err, "load network from file failed")
	}
	//获得与之匹配的相关数据
	net, ok := networks[networkName]
	if !ok {
		return fmt.Errorf("no Such Network: %s", networkName)
	}
	//释放的由该网段分配出去的IP
	if err = ipAllocator.Release(net.IPRange, &net.IPRange.IP); err != nil {
		return errors.Wrap(err, "remove Network gateway ip failed")
	}
	if err = drivers[net.Driver].Delete(net.Name); err != nil {
		return errors.Wrap(err, "remove Network DriverError failed")
	}
	// 最后从网络的配直目录中删除该网络对应的配置文件
	return net.remove(defaultNetworkPath)
}

//创建driver的时候会调用这个函数
func CreateNetwork(driver, subnet, name string) error {
	//获取非IP下的网段
	_, cidr, _ := net.ParseCIDR(subnet)
	//获得该网段下的第一个地址作为网关的地址
	ip, err := ipAllocator.Allocate(cidr)
	if err != nil {
		return err
	}
	cidr.IP = ip
	//注册驱动
	net, err := drivers[driver].Create(cidr.String(), name)
	if err != nil {
		return err
	}
	return net.dump(defaultNetworkPath)
}

func ListNetwork() {
	//1.加载Network数据
	networks, err := loadNetwork()
	if err != nil {
		logrus.Errorf("load network from file failed,detail: %v", err)
		return
	}
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	fmt.Fprint(w, "NAME\tIpRange\tDriver\n")
	for _, net := range networks {
		fmt.Fprintf(w, "%s\t%s\t%s\n",
			net.Name,
			net.IPRange.String(),
			net.Driver,
		)
	}
	if err = w.Flush(); err != nil {
		logrus.Errorf("Flush error %v", err)
		return
	}
}

func loadNetwork() (map[string]*Network, error) {
	networks := map[string]*Network{}
	// 检查网络配置目录中的所有文件,并执行第二个参数中的函数指针去处理目录下的每一个文件
	err := filepath.Walk(defaultNetworkPath, func(netPath string, info os.FileInfo, err error) error {
		// 如果是目录则跳过
		if info.IsDir() {
			return nil
		}
		//  加载文件名作为网络名
		_, netName := path.Split(netPath)
		net := &Network{
			Name: netName,
		}
		// 调用前面介绍的 Network.load 方法加载网络的配置信息
		if err = net.load(netPath); err != nil {
			logrus.Errorf("error load network: %s", err)
		}
		// 将网络的配置信息加入到 networks 字典中
		networks[netName] = net
		return nil
	})
	return networks, err
}

func (net *Network) dump(dumpPath string) error {
	if _, err := os.Stat(dumpPath); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		if err = os.MkdirAll(dumpPath, constant.Perm0622); err != nil {
			return errors.Wrapf(err, "create network dump path %s failed", dumpPath)
		}
	}
	netPath := path.Join(dumpPath, net.Name)
	// 打开保存的文件用于写入,后面打开的模式参数分别是存在内容则清空、只写入、不存在则创建
	netFile, err := os.OpenFile(netPath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, constant.Perm0644)
	if err != nil {
		return errors.Wrapf(err, "open file %s failed", dumpPath)
	}
	defer netFile.Close()
	//存入的是网络名，驱动名，以及网段
	netJson, err := json.Marshal(net)
	if err != nil {
		return errors.Wrapf(err, "Marshal %v failed", net)
	}

	_, err = netFile.Write(netJson)
	return errors.Wrapf(err, "write %s failed", netJson)
}

func (net *Network) load(dumpPath string) error {
	// 打开配置文件
	netConfigFile, err := os.Open(dumpPath)
	if err != nil {
		return err
	}
	defer netConfigFile.Close()
	// 从配置文件中读取网络 配置 json 符串
	netJson := make([]byte, 2000)
	n, err := netConfigFile.Read(netJson)
	if err != nil {
		return err
	}

	err = json.Unmarshal(netJson[:n], net)
	return errors.Wrapf(err, "unmarshal %s failed", netJson[:n])
}

func (net *Network) remove(dumpPath string) error {
	// 检查网络对应的配置文件状态，如果文件己经不存在就直接返回
	fullPath := path.Join(dumpPath, net.Name)
	if _, err := os.Stat(fullPath); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	// 否则删除这个网络对应的配置文件
	return os.Remove(fullPath)
}
