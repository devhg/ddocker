package network

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"text/tabwriter"

	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"

	"github.com/devhg/ddocker/container"
)

// 用于网络管理
type NetworkDriver interface {
	Name() string

	Create(name, subnet string) (*Network, error)

	// 删除网络
	Delete(network Network) error

	// 连接容器网络端点到网络
	Connect(network *Network, endpoint *Endpoint) error

	// 从网络上移除容器的网络端点
	Disconnect(network Network, endpoint *Endpoint) error
}

type Endpoint struct {
	ID          string
	Device      netlink.Veth
	IPaddr      net.IP
	MACaddr     net.HardwareAddr
	PortMapping []string
	Network     *Network
}

var (
	defaultNetworkPath = "/var/run/ddocker/network/network/"
	drivers            = map[string]NetworkDriver{}
	networks           = map[string]*Network{}
)

// Init 加载网路驱动
func Init() error {
	var bridgeDriver = &BridgeNetworkDriver{}
	drivers[bridgeDriver.Name()] = bridgeDriver

	if _, err := os.Stat(defaultNetworkPath); err != nil {
		if os.IsNotExist(err) {
			if err = os.MkdirAll(defaultNetworkPath, 0644); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	_ = filepath.Walk(defaultNetworkPath, func(nwpath string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		_, fileName := path.Split(nwpath)
		n := &Network{Name: fileName}

		if err := n.load(nwpath); err != nil {
			logrus.Errorf("error load network: %v", err)
		}

		networks[fileName] = n
		return nil
	})

	return nil
}

func ListNetwork() {
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	fmt.Fprintf(w, "Name\tIPRange\tDriver\n")
	for _, v := range networks {
		fmt.Fprintf(w, "%s\t%s\t%s\n", v.Name, v.IPRange, v.Driver)
	}

	if err := w.Flush(); err != nil {
		logrus.Errorf("flush error: %v", err)
	}
}

func CreateNetwork(driver, subnet, name string) error {
	_, cidr, err := net.ParseCIDR(subnet)
	if err != nil {
		panic(err)
	}

	// IPAM 分配网关ip，获取到网段的第一个ip作为网关的ip
	gatewayIP, err := ipAllocator.Allocate(cidr)
	if err != nil {
		return err
	}

	cidr.IP = gatewayIP

	network, err := drivers[driver].Create(name, cidr.String())
	if err != nil {
		return err
	}

	// 保存网络信息到文件系统中，以便查询和在网络上连接端点
	return network.dump(defaultNetworkPath)
}

func DeleteNetwork(networkName string) error {
	nw, ok := networks[networkName]
	if !ok {
		return fmt.Errorf("no such network: %s", networkName)
	}

	if err := ipAllocator.Release(nw.IPRange, nw.IPRange.IP); err != nil {
		return err
	}

	if err := drivers["bridge"].Delete(*nw); err != nil {
		return fmt.Errorf("remove network error: %v", err)
	}

	return nw.remove(defaultNetworkPath)
}

func Connect(networkName string, cinfo *container.ContainerInfo) error {
	// 通过networkName获取对应已经创建的network
	network, ok := networks[networkName]
	if !ok {
		return fmt.Errorf("no such network: %s", networkName)
	}

	// 分配容器的IP地址
	ip, err := ipAllocator.Allocate(network.IPRange)
	if err != nil {
		return err
	}

	// 创建容器的 网络端点，设置网络端点的IP，端口的映射信息
	endpoint := &Endpoint{
		ID:          cinfo.ID + "-" + networkName,
		IPaddr:      ip,
		Network:     network,
		PortMapping: cinfo.PortMapping,
	}

	// 调用网络驱动的 Connect 方法去连接和配置网络端点
	if err := drivers[network.Driver].Connect(network, endpoint); err != nil {
		return err
	}

	// 进入容器的网络namespace，配置容器网络、设备的IP地址和路由
	if err := configEndpointIPAddrAndRoute(endpoint, cinfo); err != nil {
		return err
	}

	// 配置容器到宿主机的端口映射
	return configPortMapping(endpoint)
}

// 进入容器的网络namespace，配置容器网络设备的 IP 地址和路由
func configEndpointIPAddrAndRoute(ep *Endpoint, cinfo *container.ContainerInfo) error {
	// 通过name获取已经接入Linux Bridge的veth
	peerLink, err := netlink.LinkByName(ep.Device.PeerName)
	if err != nil {
		return fmt.Errorf("fail config endpoint: %v", err)
	}

	// 将上面获取到的网络端点veth，加入到容器的net namespace中
	// 并使这个函数下面的操作都在这个网络空间中进行，执行完恢复默认的网络空间
	defer enterContainerNetns(&peerLink, cinfo)()

	interfaceIP := ep.Network.IPRange
	interfaceIP.IP = ep.IPaddr

	// 1. 设置容器内veth端点的IP
	if err = setInterfaceIP(ep.Device.PeerName, interfaceIP.String()); err != nil {
		return fmt.Errorf("%v,%s", ep.Network, err)
	}

	// 2. 启动容器内的veth端点
	if err = setInterfaceUP(ep.Device.PeerName); err != nil {
		return err
	}

	// 3. 容器 net namespace 中默认本地地址是127.0.0.1的“lo”网卡的状态默认是关闭的
	// 启动它以保证容器访问自己的请求
	if err = setInterfaceUP("lo"); err != nil {
		return err
	}

	// 4. 设置容器内的所有外部请求都通过容器的veth端点访问
	// 0.0.0.0/0网段，表示所有的IP地址段
	_, cidr, _ := net.ParseCIDR("0.0.0.0/0")

	// 构建要添加的路由数据，包括网络设备、网关IP及目的网段
	defaultRoute := &netlink.Route{
		LinkIndex: peerLink.Attrs().Index,
		Gw:        ep.Network.IPRange.IP,
		Dst:       cidr,
	}

	// 添加路由到容器的网络空间
	// $(route add -net 0.0.0.0/0 gw ${Bridge网桥地址} dev ${容器内的veth端点设备})
	return netlink.RouteAdd(defaultRoute)
}

func enterContainerNetns(enLink *netlink.Link, cinfo *container.ContainerInfo) func() {
	// 找到容器的 net namespace    /proc/[pid]/ns/net
	cnsnet := fmt.Sprintf("/proc/%s/ns/net", cinfo.PID)
	f, err := os.OpenFile(cnsnet, os.O_RDONLY, 0)
	if err != nil {
		logrus.Errorf("error get container net namespace, %v", err)
	}

	nsFD := f.Fd()

	// 锁定当前程序的线程，如果不锁定，goroutine可能会调度到别的线程上去，
	// 就不能保证一直在所需要的网络空间中了。
	runtime.LockOSThread()

	// 1. 修改veth peer 另外一端移到容器的 net namespace 中
	if err = netlink.LinkSetNsFd(*enLink, int(nsFD)); err != nil {
		logrus.Errorf("error set link netns to container namespace, %v", err)
	}

	// 2. 获取当前的网络namespace，以便以后从容器net namespace退出到当前namespace
	origns, err := netns.Get()
	if err != nil {
		logrus.Errorf("error get current netns, %v", err)
	}

	// 3. 设置当前进程到新的网络namespace，并在函数执行完成之后再恢复到之前的namespace
	if err = netns.Set(netns.NsHandle(nsFD)); err != nil {
		logrus.Errorf("error set netns, %v", err)
	}

	return func() {
		// 在容器net namespace中，执行此函数恢复到原来的net namespace
		_ = netns.Set(origns)
		origns.Close()
		runtime.UnlockOSThread()
		f.Close()
	}
}

func configPortMapping(ep *Endpoint) error {
	for _, pm := range ep.PortMapping {
		portMapping := strings.Split(pm, ":")
		if len(portMapping) != 2 {
			logrus.Errorf("port mapping format error, %v", pm)
			continue
		}

		// 由于iptables没有go语言的实现，采用exec.Command的方式直接调用命令配置
		// 在iptables的PREROUTING中添加DNAT规则，将宿主机端口转发到容器的地址端口上
		iptablesCmd := fmt.Sprintf("-t nat -A PREROUTING -p tcp -m tcp --dport %s -j DNAT --to-destination %s:%s",
			portMapping[0], ep.IPaddr.String(), portMapping[1])

		subcmds := strings.Split(iptablesCmd, " ")
		cmd := exec.Command("iptables", subcmds...)
		if output, err := cmd.Output(); err != nil {
			logrus.Errorf("iptables Output, %v", output)
			continue
		}
	}
	return nil
}

type Network struct {
	Name    string
	IPRange *net.IPNet
	Driver  string
}

// 将网络信息保存到网络配置目录的文件中
func (n *Network) dump(dumpPath string) error {
	if _, err := os.Stat(dumpPath); err != nil {
		if os.IsNotExist(err) {
			if err = os.MkdirAll(dumpPath, 0644); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	// 网络信息保存的目标文件
	dst := path.Join(dumpPath, n.Name)

	f, err := os.OpenFile(dst, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	b, _ := json.Marshal(n)
	_, err = f.Write(b)
	return err
}

// 从网络的配置文件中加载网络信息
func (n *Network) load(dst string) error {
	f, err := os.Open(dst)
	if err != nil {
		return err
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, n)
}

// 从网络配置目录中删除网络的配置文件
func (n *Network) remove(dumpPath string) error {
	dst := path.Join(dumpPath, n.Name)

	if _, err := os.Stat(dst); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	return os.Remove(dst)
}
