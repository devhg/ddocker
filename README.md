# ddocker

<<自己动手写docker>>

> Doc: <https://www.yuque.com/playgo/ddocker>

## call flow

![ddocker](imgs/ddocker.png)

## run command

```bash
root@ubuntu1404:~/GoWork/src/github.com/devhg/ddocker# go build .

root@ubuntu1404:~/GoWork/src/github.com/devhg/ddocker# ./ddocker run -it -mm 100m stress --vm-bytes 200m --vm-keep -m 1
```
