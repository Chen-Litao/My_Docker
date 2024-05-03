package network

import (
	"encoding/json"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"myself_docker/constant"
	"net"
	"os"
	"path"
	"strings"
)

const ipamDefaultAllocatorPath = "/var/lib/mydocker/network/ipam/subnet.json"

type IPAM struct {
	SubnetAllocatorPath string             // 分配文件存放位置
	Subnets             *map[string]string // 网段和位图算法的数组 map, key 是网段， value 是分配的位图数组
}

// 初始化一个IPAM的对象，默认使用/var/lib/mydocker/network/ipam/subnet.json作为分配信息存储位置
var ipAllocator = &IPAM{
	SubnetAllocatorPath: ipamDefaultAllocatorPath,
}

//输入：网段xxx.xxx.xxx.xxx\xx
//输出：分配的IP
func (ipam *IPAM) Allocate(subnet *net.IPNet) (ip net.IP, err error) {
	// 存放网段中地址分配信息的数组
	ipam.Subnets = &map[string]string{}
	err = ipam.load()
	if err != nil {
		return nil, errors.Wrap(err, "load subnet allocation info error")
	}
	//获得网段以及子网掩码
	_, subnet, _ = net.ParseCIDR(subnet.String())
	one, size := subnet.Mask.Size()
	if _, exist := (*ipam.Subnets)[subnet.String()]; !exist {
		// ／用“0”填满这个网段的配置，uint8(size - one ）表示这个网段中有多少个可用地址
		// size - one是子网掩码后面的网络位数，2^(size - one)表示网段中的可用IP数
		// 而2^(size - one)等价于1 << uint8(size - one)
		// 左移一位就是扩大两倍
		(*ipam.Subnets)[subnet.String()] = strings.Repeat("0", 1<<uint8(size-one))
	}
	for c := range (*ipam.Subnets)[subnet.String()] {
		// 找到数组中为“0”的项和数组序号，即可以分配的 IP
		if (*ipam.Subnets)[subnet.String()][c] == '0' {
			// Go 的字符串，创建之后就不能修改 所以通过转换成 byte 数组，修改后再转换成字符串赋值
			ipalloc := []byte((*ipam.Subnets)[subnet.String()])
			ipalloc[c] = '1'
			(*ipam.Subnets)[subnet.String()] = string(ipalloc)
			ip = subnet.IP
			for t := uint(4); t > 0; t -= 1 {
				[]byte(ip)[4-t] += uint8(c >> ((t - 1) * 8))
			}
			ip[3] += 1
			break
		}
	}
	err = ipam.dump()
	if err != nil {
		log.Error("Allocate：dump ipam error", err)
	}
	return
}

//load 加载网段地址分配信息
func (ipam *IPAM) load() error {
	if _, err := os.Stat(ipam.SubnetAllocatorPath); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	// 读取文件，加载配置信息
	subnetConfigFile, err := os.Open(ipam.SubnetAllocatorPath)
	if err != nil {
		return err
	}
	//将读取出来的位图信息解码存储到ipam.Subnets
	defer subnetConfigFile.Close()
	subnetJson := make([]byte, 2000)
	n, err := subnetConfigFile.Read(subnetJson)
	if err != nil {
		return errors.Wrap(err, "read subnet config file error")
	}
	err = json.Unmarshal(subnetJson[:n], ipam.Subnets)
	return errors.Wrap(err, "err dump allocation info")
}

func (ipam *IPAM) dump() error {
	ipamConfigFileDir, _ := path.Split(ipam.SubnetAllocatorPath)
	//判断当前目录是否存在
	if _, err := os.Stat(ipamConfigFileDir); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		if err = os.MkdirAll(ipamConfigFileDir, constant.Perm0644); err != nil {
			return err
		}
	}
	// 打开存储文件 O_TRUNC 表示如果存在则消空，os.O_WRONLY：以只写模式打开文件。os.O_CREATE：如果指定的文件不存在，就创建它。指定文件的权限位
	subnetConfigFile, err := os.OpenFile(ipam.SubnetAllocatorPath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, constant.Perm0644)
	if err != nil {
		return err
	}
	defer subnetConfigFile.Close()
	ipamConfigJson, err := json.Marshal(ipam.Subnets)
	if err != nil {
		return err
	}
	_, err = subnetConfigFile.Write(ipamConfigJson)
	return err
}
