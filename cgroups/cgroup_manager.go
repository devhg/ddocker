package cgroups

import (
	"github.com/devhg/ddocker/cgroups/subsystems"
	"github.com/sirupsen/logrus"
)

type CgroupManager struct {
	Path     string
	Resource *subsystems.ResourceConfig
}

func NewCgroupManager(path string) *CgroupManager {
	return &CgroupManager{
		Path: path,
	}
}

// 将PID加入到每个cgroup中
func (c *CgroupManager) Apply(pid int) error {
	for _, subSysIns := range subsystems.SubsystemIns {
		if err := subSysIns.Apply(c.Path, pid); err != nil {
			return err
		}
	}
	return nil
}

// 设置每个subsystem挂载中的cgroup资源限制
func (c *CgroupManager) Set(res *subsystems.ResourceConfig) error {
	c.Resource = res
	for _, subSysIns := range subsystems.SubsystemIns {
		if err := subSysIns.Set(c.Path, c.Resource); err != nil {
			return err
		}
	}
	return nil
}

// 释放各个subsystem挂载的cgroup
func (c *CgroupManager) Destroy() {
	for _, subSysIns := range subsystems.SubsystemIns {
		if err := subSysIns.Remove(c.Path); err != nil {
			logrus.Warnln(err)
			panic(err)
		}
	}
}
