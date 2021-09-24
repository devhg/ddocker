package network

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"net"
	"os"
	"path"
	"path/filepath"
	"text/tabwriter"

	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"

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
	defaultNetworkPath = "/var/run/mydocker/network/network/"
	drivers            = map[string]NetworkDriver{}
	networks           = map[string]*Network{}
)

// Init 加载网路驱动
func Init() error {
	var bridgeDriver = BridgeNetworkDriver{}
	drivers[bridgeDriver.Name()] = &bridgeDriver

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
	fmt.Fprintf(w, "Name\tIPRange\tDriver")
	for _, v := range networks {
		fmt.Fprintf(w, "%s\t%s\t%s\t", v.Name, v.IPRange, v.Driver)
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

	cidr.IP = net.IP(gatewayIP)

	network, err := drivers[driver].Create(cidr.String(), name)
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

	if err := drivers[nw.Driver].Delete(*nw); err != nil {
		return fmt.Errorf("remove network error: %v", err)
	}

	return nw.remove(defaultNetworkPath)
}

func Connect(networkName string, cinfo *container.ContainerInfo) error {
	network, ok := networks[networkName]
	if !ok {
		return fmt.Errorf("no such network: %s", networkName)
	}

	ip, err := ipAllocator.Allocate(network.IPRange)
	if err != nil {
		return err
	}

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

	// 进入容器的网络namespace，配置容器网络设备的 IP 地址和路由
	if err := configEndpointIPAddrAndRoute(endpoint, cinfo); err != nil {
		return err
	}

	// 配置容器到宿主机的端口映射
	return configPortMapping(endpoint, cinfo)
}

func configEndpointIPAddrAndRoute(endpoint *Endpoint, cinfo *container.ContainerInfo) error {
	return nil
}

func configPortMapping(endpoint *Endpoint, cinfo *container.ContainerInfo) error {
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
