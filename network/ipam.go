package network

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path"
	"strings"

	"github.com/sirupsen/logrus"
)

// 用于网络IP地址的分配和释放
type IPAM struct {
	// 分配配置文件存储位置
	SubnetAllocatorPath string

	Subnets map[string]string
}

const (
	ipamDefaultAllocatorPath = "/var/run/ddocker/network/ipam/subnet.json"
)

var ipAllocator = &IPAM{
	SubnetAllocatorPath: ipamDefaultAllocatorPath,
}

func (ia *IPAM) load() error {
	if _, err := os.Stat(ia.SubnetAllocatorPath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	f, err := os.Open(ia.SubnetAllocatorPath)
	if err != nil {
		return err
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	// 反序列化分配的IP信息
	return json.Unmarshal(b, &ia.Subnets)
}

func (ia *IPAM) dump() error {
	// 检查存储文件所在的文件夹是否存在，如果不存在则创建
	dir, _ := path.Split(ia.SubnetAllocatorPath)
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			if err = os.MkdirAll(dir, 0644); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	// 打开存储文件
	f, err := os.OpenFile(ia.SubnetAllocatorPath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	// 序列化并保存
	b, _ := json.Marshal(ia.Subnets)
	_, err = f.Write(b)
	return err
}

// 通过网段去分配一个可用的IP地址
func (ia *IPAM) Allocate(subnet *net.IPNet) (ip net.IP, err error) {
	ia.Subnets = make(map[string]string)

	if err = ia.load(); err != nil {
		fmt.Println(err)
		logrus.Infof("load IPAM allocation info error: %v", err)
	}

	// 返回网段的子网掩码的总长度 和 网段前面的固定定位的长度
	// "127.0.0.1/8" 子网掩码是 "255.0.0.0"。返回值就是 8 和 32
	ones, bits := subnet.Mask.Size()

	if _, exist := ia.Subnets[subnet.String()]; !exist {
		// 用"0"填充满这个网段的配置
		// 1<<uint8(bits-ones) 表示此网段有 2^(bits-ones) 个IP数目
		ia.Subnets[subnet.String()] = strings.Repeat("0", 1<<uint8(bits-ones))
	}

	bitmap := []byte(ia.Subnets[subnet.String()])
	for i, c := range bitmap {
		if c == '0' {
			bitmap[i] = '1'
			ia.Subnets[subnet.String()] = string(bitmap)

			ip = subnet.IP

			// 通过网段的IP和上面的偏移相加计算出分配的IP地址，IP地址是一个uint8的数组
			// 需要通过的数组的每一项加所需的值，比如网段是172.16.0.0/12，数组序号65555
			// 那么需要在[172.16.0.0]基础上，计算变成
			// [uint8(65555>>24), uint8(65555>>16), uint8(65555>>8), uint8(65555>>0)]
			// [172, 17, 0, 19]
			for t := 4; t > 0; t-- {
				[]byte(ip)[4-t] += uint8(i >> ((t - 1) * 8))
			}
			// 由于IP是从1开始分配的，因此需要最后加1。最后变成172.17.0.20
			[]byte(ip)[3]++

			break
		}
	}

	if err = ia.dump(); err != nil {
		err = fmt.Errorf("dump IPAM info error: %v after allocate", err)
	}
	return ip, err
}

func (ia *IPAM) Release(subnet *net.IPNet, ipaddr net.IP) error {
	ia.Subnets = make(map[string]string)

	if err := ia.load(); err != nil {
		logrus.Infof("load IPAM allocation info error: %v", err)
	}

	releaseIP := ipaddr.To4()
	offset := 0 // 位图	偏移

	// 对应分配时候的加1
	releaseIP[3]--
	for t := 4; t > 0; t-- {
		offset += int(releaseIP[t-1]-subnet.IP[t-1]) << ((4 - t) * 8)
	}

	// 将分配位图数组中的索引位置的值置0
	bitmap := []byte(ia.Subnets[subnet.String()])
	bitmap[offset] = '0'
	ia.Subnets[subnet.String()] = string(bitmap)

	// 重新保存分配位图信息
	if err := ia.dump(); err != nil {
		return fmt.Errorf("dump IPAM info error: %v after release", err)
	}
	return nil
}
