## 遇到的问题

> <https://www.yuque.com/playgo/ddocker>

1. pstree -pl 进程树

sh: pstree: 未找到命令
`yum install -y psmisc`

2. stress 安装（done）

`sudo yum install -y epel-release`

`sudo yum install -y stress`

3. cpuset limit（doing 似乎Ubuntu14背锅）

`{"level":"info","msg":"set cgroup proc failed write /sys/fs/cgroup/cpuset/ddocker-cgroup/tasks: no space left on device","time":"2021-07-25T16:33:09+08:00"}`

4. 打开/proc/self/mountinfo只有首次有效，第二次就会出现如下的错误（doing 似乎Ubuntu14背锅）

`{"level":"warning","msg":"open /proc/self/mountinfo: no such file or directory","time":"2021-07-25T21:11:39+08:00"}`

5. pivot_root 系统调用似乎没有生效（doing）
![](https://cdn.nlark.com/yuque/0/2021/png/2977669/1627832811471-57c6677a-94a8-4d94-955a-57ca40755a30.png)