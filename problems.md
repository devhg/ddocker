## 遇到的问题

1. pstree -pl 进程树

sh: pstree: 未找到命令
`yum install -y psmisc`

2. stress 安装（done）

`sudo yum install -y epel-release`

`sudo yum install -y stress`

3. cpuset limit（doing）

`{"level":"info","msg":"set cgroup proc failed write /sys/fs/cgroup/cpuset/ddocker-cgroup/tasks: no space left on device","time":"2021-07-25T16:33:09+08:00"}`

4. 打开/proc/self/mountinfo只有首次有效，第二次就会出现如下的错误
`{"level":"warning","msg":"open /proc/self/mountinfo: no such file or directory","time":"2021-07-25T21:11:39+08:00"}`
