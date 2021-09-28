package network

import (
	"fmt"
	"net"
	"testing"

	"github.com/vishvananda/netlink"

	"github.com/devhg/ddocker/container"
)

func TestBridgeInit(t *testing.T) {
	d := BridgeNetworkDriver{}
	_, err := d.Create("testbridge", "192.168.0.1/24")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("err: %v", err)
}

func TestBridgeConnect(t *testing.T) {
	ep := Endpoint{
		ID: "testcontainer",
	}

	n := Network{
		Name: "testbridge",
	}

	d := BridgeNetworkDriver{}
	err := d.Connect(&n, &ep)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("err: %v", err)
}

func TestNetworkConnect(t *testing.T) {
	cInfo := &container.ContainerInfo{
		ID:  "testcontainer",
		PID: "15438",
	}

	d := BridgeNetworkDriver{}
	n, err := d.Create("testbridge", "192.168.0.2/24")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("network: %v", n)

	err = Init()
	if err != nil {
		t.Fatal(err)
	}

	networks[n.Name] = n
	err = Connect(n.Name, cInfo)
	t.Logf("err: %v", err)

	err = DeleteNetwork("testbridge")
	if err != nil {
		t.Fatal(err)
	}
}

func TestLoad(t *testing.T) {
	n := Network{
		Name: "testbridge",
	}

	err := n.load("/var/run/ddocker/network/network/testbridge")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("network: %v", n)
}

func TestParseCIDR(t *testing.T) {
	ip, ipNet, _ := net.ParseCIDR("192.168.0.2/16")
	fmt.Println("net.ParseCIDR", ip, ipNet.IP)

	ipNet2, _ := netlink.ParseIPNet("192.168.0.2/16")
	fmt.Println("netlink.ParseCIDR", ipNet2.IP)
}
