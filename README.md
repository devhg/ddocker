# ddocker

<<自己动手写docker>>

> Doc: <https://www.yuque.com/playgo/ddocker>

## call flow

![ddocker](imgs/ddocker.png)

## cgroup resource control

![ddocker](imgs/cgroup.png)

## create container with pipe

![ddocker](imgs/ddocker-pipe.png)

## run command

```bash
root@ubuntu1404:~/GoWork/src/github.com/devhg/ddocker# go build .

root@ubuntu1404:~/GoWork/src/github.com/devhg/ddocker# ./ddocker run -it -mm 100m stress --vm-bytes 200m --vm-keep -m 1

root@ubuntu1404:~/GoWork/src/github.com/devhg/ddocker# ./ddocker run -it -mm 100m -cpushare 512 stress --vm-bytes 200m --vm-keep -m 1

root@ubuntu1404:~/GoWork/src/github.com/devhg/ddocker# ./ddocker run -it ls -l

root@ubuntu1404:~/GoWork/src/github.com/devhg/ddocker# ./ddocker run -it bash
```

## Preview

create a busybox container
![](./imgs/create-busybox-container.png)

## Q & A

[Q&A](./problems.md)
