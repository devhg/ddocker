package network

type BridgeNetworkDriver struct {
}

func (b *BridgeNetworkDriver) Name() string {
	return "bridge"
}

func (b *BridgeNetworkDriver) Create(name string, subnet string) (*Network, error) {
	panic("not implemented") // TODO: Implement
}

// 删除网络
func (b *BridgeNetworkDriver) Delete(network Network) error {
	panic("not implemented") // TODO: Implement
}

// 连接容器网络端点到网络
func (b *BridgeNetworkDriver) Connect(network *Network, endpoint *Endpoint) error {
	panic("not implemented") // TODO: Implement
}

// 从网络上移除容器的网络端点
func (b *BridgeNetworkDriver) Disconnect(network Network, endpoint *Endpoint) error {
	panic("not implemented") // TODO: Implement
}
