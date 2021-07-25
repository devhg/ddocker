## 遇到的问题

1. pstree -pl 进程树

sh: pstree: 未找到命令
`yum install -y psmisc`

2. stress 安装（done）

`sudo yum install -y epel-release`

`sudo yum install -y stress`

3. cpuset limit（doing）

`{"level":"info","msg":"set cgroup proc failed write /sys/fs/cgroup/cpuset/ddocker-cgroup/tasks: no space left on device","time":"2021-07-25T16:33:09+08:00"}`
