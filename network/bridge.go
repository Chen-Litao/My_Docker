package network

import (
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"net"
	"os/exec"
	"strings"
	"time"
)

type BridgeNetworkDriver struct {
}

func (d *BridgeNetworkDriver) Name() string {
	return "bridge"
}

func (d *BridgeNetworkDriver) Create(subnet string, name string) (*Network, error) {
	ip, ipRange, _ := net.ParseCIDR(subnet)
	ipRange.IP = ip
	n := &Network{
		Name:    name,
		IPRange: ipRange,
		Driver:  d.Name(),
	}
	err := d.initBridge(n)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to create bridge network")
	}
	return n, err
}

// Delete 删除网络
func (d *BridgeNetworkDriver) Delete(name string) error {
	// 根据名字找到对应的Bridge设备
	br, err := netlink.LinkByName(name)
	if err != nil {
		return err
	}
	// 删除网络对应的 Linux Bridge 设备
	return netlink.LinkDel(br)
}

func (d *BridgeNetworkDriver) Connect(networkName string, endpoint *Endpoint) error {
	bridgeName := networkName
	br, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return err
	}
	// 创建 Veth 接口的配置
	la := netlink.NewLinkAttrs()
	// 由于 Linux 接口名的限制,取 endpointID 的前
	la.Name = endpoint.ID[:5]
	// 通过设置 Veth 接口 master 属性，设置这个Veth的一端挂载到网络对应的 Linux Bridge
	la.MasterIndex = br.Attrs().Index
	endpoint.Device = netlink.Veth{
		LinkAttrs: la,
		PeerName:  "cif-" + endpoint.ID[:5],
	}
	// 调用netlink的LinkAdd方法创建出这个Veth接口
	// 因为上面指定了link的MasterIndex是网络对应的Linux Bridge
	// 所以Veth的一端就已经挂载到了网络对应的LinuxBridge.上
	if err = netlink.LinkAdd(&endpoint.Device); err != nil {
		return fmt.Errorf("error Add Endpoint Device: %v", err)
	}
	// 调用netlink的LinkSetUp方法，设置Veth启动
	// 相当于ip link set xxx up命令
	if err = netlink.LinkSetUp(&endpoint.Device); err != nil {
		return fmt.Errorf("error Add Endpoint Device: %v", err)
	}
	return nil
}

func (d *BridgeNetworkDriver) Disconnect(endpointID string) error {
	// 根据名字找到对应的 Veth 设备
	vethNme := endpointID[:5] // 由于 Linux 接口名的限制,取 endpointID 的前 5 位
	veth, err := netlink.LinkByName(vethNme)
	if err != nil {
		return err
	}
	// 从网桥解绑
	err = netlink.LinkSetNoMaster(veth)
	if err != nil {
		return errors.WithMessagef(err, "find veth [%s] failed", vethNme)
	}
	// 删除 veth-pair
	// 一端为 xxx,另一端为 cif-xxx
	err = netlink.LinkDel(veth)
	if err != nil {
		return errors.WithMessagef(err, "delete veth [%s] failed", vethNme)
	}
	veth2Name := "cif-" + vethNme
	veth2, err := netlink.LinkByName(veth2Name)
	if err != nil {
		return errors.WithMessagef(err, "find veth [%s] failed", veth2Name)
	}
	err = netlink.LinkDel(veth2)
	if err != nil {
		return errors.WithMessagef(err, "delete veth [%s] failed", veth2Name)
	}

	return nil
}

/*
  Linux Bridge 初始化流程
	1.创建bridge 虚拟设备
	2.设置bridge 设备地址
	3.启动
	4.设置规则
*/

func (d *BridgeNetworkDriver) initBridge(n *Network) error {
	bridgeName := n.Name
	// 1）创建 Bridge 虚拟设备
	if err := createBridgeInterface(bridgeName); err != nil {
		return errors.Wrapf(err, "Failed to create bridge %s", bridgeName)
	}
	// 2）设置 Bridge 设备地址和路由
	gatewayIP := *n.IPRange
	gatewayIP.IP = n.IPRange.IP

	if err := setInterfaceIP(bridgeName, gatewayIP.String()); err != nil {
		return errors.Wrapf(err, "Error set bridge ip: %s on bridge: %s", gatewayIP.String(), bridgeName)
	}
	// 3）启动 Bridge 设备
	if err := setInterfaceUP(bridgeName); err != nil {
		return errors.Wrapf(err, "Failed to set %s up", bridgeName)
	}
	// 4）设置 iptables SNAT 规则
	if err := setupIPTables(bridgeName, n.IPRange); err != nil {
		return errors.Wrapf(err, "Failed to set up iptables for %s", bridgeName)
	}
	return nil
}

func createBridgeInterface(bridgeName string) error {
	// 先检查是否己经存在了这个同名的Bridge设备
	/*
	* net.InterfaceByName是用于查找是否存在该网络接口
	 */
	_, err := net.InterfaceByName(bridgeName)
	if err == nil || !strings.Contains(err.Error(), "no such network interface") {
		return err
	}
	// create *netlink.Bridge object
	la := netlink.NewLinkAttrs()
	la.Name = bridgeName
	// 使用刚才创建的Link的属性创netlink.Bridge对象
	br := &netlink.Bridge{LinkAttrs: la}
	// 调用 net link Linkadd 方法，创 Bridge 虚拟网络设备
	// netlink.LinkAdd 方法是用来创建虚拟网络设备的，相当于 ip link add xxxx
	if err = netlink.LinkAdd(br); err != nil {
		return errors.Wrapf(err, "create bridge %s error", bridgeName)
	}
	return nil
}

func setInterfaceIP(name, rawIP string) error {
	retries := 2
	var iface netlink.Link
	var err error
	//循环两次。适应网络环境的微小波动，同时也避免在遇到持久错误时造成更大的问题。
	for i := 0; i < retries; i++ {
		// 通过LinkByName方法找到需要设置的网络接口
		iface, err = netlink.LinkByName(name)
		if err == nil {
			break
		}
		log.Debugf("error retrieving new bridge netlink link [ %s ]... retrying", name)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return errors.Wrap(err, "abandoning retrieving the new bridge link from netlink, Run [ ip link ] to troubleshoot")
	}
	//获取网段新消息和原IP信息
	ipNet, err := netlink.ParseIPNet(rawIP)
	if err != nil {
		return err
	}
	//通过  netlink.AddrAdd给网络接口配置地址，相当于ip addr add xxx命令
	//配置路由表 192.168.0.0/24 转发到这 testbridge 的网络接口上
	addr := &netlink.Addr{IPNet: ipNet}
	return netlink.AddrAdd(iface, addr)
}

func setInterfaceUP(interfaceName string) error {
	link, err := netlink.LinkByName(interfaceName)
	if err != nil {
		return errors.Wrapf(err, "error retrieving a link named [ %s ]:", link.Attrs().Name)
	}
	// 等价于 ip link set xxx up 命令
	if err = netlink.LinkSetUp(link); err != nil {
		return errors.Wrapf(err, "nabling interface for %s", interfaceName)
	}
	return nil
}

// setupIPTables 设置 iptables 对应 bridge MASQUERADE 规则
func setupIPTables(bridgeName string, subnet *net.IPNet) error {
	// 拼接命令
	iptablesCmd := fmt.Sprintf("-t nat -A POSTROUTING -s %s ! -o %s -j MASQUERADE", subnet.String(), bridgeName)
	cmd := exec.Command("iptables", strings.Split(iptablesCmd, " ")...)
	// 执行该命令
	output, err := cmd.Output()
	if err != nil {
		log.Errorf("iptables Output, %v", output)
	}
	return err
}
