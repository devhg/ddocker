package container

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/sirupsen/logrus"
)

// RunContainerInitProcess 是在容器内部执行的，也就是说代码执行到这里后，
// 容器所在的进程其实就已经创建完成了，这是本容器执行的第一个进程。
//
// syscall.Exec()最终调用了kernel的int execve(const char* filename, char* const argv[], char* const emvp[])
// 这个系统调用完成了初始化动作并将用户进程运行起来的动作，
// 在前面的代码中，容器的第一个进程是init初始化的进程，我们希望是我们自己的进程。但是PID=1的进程是不能kill的，
// 这个系统调用的作用就是，将原来的init进程替换成用户自己的进程，这样当进入容器的时候，PID=1的程序就是我们指定的进程了。
// 容器 === 进程。这其实也是目前docker使用容器引擎runC的实现方式之一
func RunContainerInitProcess() error {
	cmdArray := readUserCommand()
	// should omit nil check; len() for nil slices is defined as zero (S1009)   good!
	if len(cmdArray) == 0 {
		return fmt.Errorf("run container get user command error, cmdArray is nil")
	}
	setUpMount()

	// 在系统PATH中寻找命令的绝对路径
	cmdPath, err := exec.LookPath(cmdArray[0])
	if err != nil {
		logrus.Errorf("exec loop error %v", err)
		return err
	}
	logrus.Infof("found path is %s", cmdPath)
	logrus.Infoln(cmdPath, cmdArray)
	if err := syscall.Exec(cmdPath, cmdArray, os.Environ()); err != nil {
		logrus.Errorln(err.Error())
	}
	return nil
}

func readUserCommand() []string {
	// 一个进程的创建，默认有三个文件描述符，[标准输入 标准输出 标准错误]
	// uintptr(3) 是指index=3的文件描述符，也就是传进来管道的一端(readPipe)
	pipe := os.NewFile(uintptr(3), "pipe")
	msg, err := ioutil.ReadAll(pipe)
	if err != nil {
		logrus.Errorf("init read pipe error %v", err)
		return nil
	}
	msgStr := string(msg)
	return strings.Split(msgStr, " ")
}

// setUpMount Init 挂载点
// 使用mount先去挂在proc文件系统，以后后面通过ps等系统命令去查看进程的资源占用情况
//
// syscall.MS_NOEXEC 本文件系统中不允许运行其他程序
// syscall.MS_NOSUID 本文件系统运行程序，禁止set-user-ID或set-group-ID
// syscall.MS_NODEV  所有mount的系统都会默认设定的参数
//
func setUpMount() {
	pwd, err := os.Getwd()
	if err != nil {
		logrus.Errorf("get current location error: %v", err)
		return
	}

	logrus.Info("current location is: ", pwd)
	if err = pivotRoot(pwd); err != nil {
		logrus.Warn(err)
	}

	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	_ = syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
	_ = syscall.Mount("tmpfs", "/dev", "tmpfs", syscall.MS_NOSUID|syscall.MS_STRICTATIME, "mode=755")
}

// pivotRoot 改变当前的root文件系统，对应pivot_root系统调用
// 可以将当前进程的root文件系统移动到put_old文件夹，然后使new_root成为新的root文件系统。
// pivotRoot和chroot的主要区别：
// 	  pivot_root是把整个系统切换到一个新的root目录，移除对之前root的依赖，以便随时umount原来的文件系统
//    chroot是针对某个进程，系统的其他部分依旧运行于老的root目录中
func pivotRoot(root string) error {
	// 为了使当前root的 老root 和 新root 不在同一个文件系统下，我们把root重新mount了一次
	// bind mount是把相同的内容换了一个挂载点的挂载方法
	if err := syscall.Mount(root, root, "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("func[pivotRoot] mount rootfs to itself error: %v", err)
	}

	// 创建rootfs/.pivot_root 存储 old_root
	pivotDir := filepath.Join(root, ".pivot_root")
	if err := os.Mkdir(pivotDir, 0777); err != nil {
		return err
	}

	// pivot_root 到新的rootfs，old_root 现在挂载在rootfs/.pivot_root上
	// 挂载点目前依然可以在mount命令中看到
	if err := syscall.PivotRoot(root, pivotDir); err != nil {
		return fmt.Errorf("func[pivotRoot] syscall.PivotRoot error: %v", err)
	}

	// 修改当前的工作目录到根目录
	if err := syscall.Chdir("/"); err != nil {
		return fmt.Errorf("func[pivotRoot] syscall.Chdir / error: %v", err)
	}

	// umount rootfs/.pivot_root
	pivotDir = filepath.Join("/", ".pivot_root")
	if err := syscall.Unmount(pivotDir, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("func[pivotRoot] umount pivot_root dir error: %v", err)
	}

	return os.Remove(pivotDir)
}

// NewWorkSpace creates a workspace
func NewWorkSpace(rootURL, mntURL, volume string) {
	createReadOnlyLayer(rootURL)
	createWriteLayer(rootURL)
	createMountPoint(rootURL, mntURL)

	if volume != "" {
		volumeURLs := strings.Split(volume, ":")
		if len(volumeURLs) == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
			mountVolume(rootURL, mntURL, volumeURLs)
			return
		}
		logrus.Errorf("bad volume: %v", volume)
	}
}

func createReadOnlyLayer(rootURL string) {
	busyboxURL := path.Join(rootURL, "busybox")
	busyboxTar := path.Join(rootURL, "busybox.tar")
	if exist, _ := pathExist(busyboxURL); exist {
		return
	}

	exist, err := pathExist(busyboxTar)
	if err != nil {
		logrus.Infof("failed to judge whether %v exists. %v", busyboxTar, err)
	}

	if exist {
		if err := os.Mkdir(busyboxURL, 0777); err != nil {
			logrus.Errorf("mkdir %v error: %v", busyboxURL, err)
		}

		if _, err := exec.Command("tar", "-xvf", busyboxTar, "-C", busyboxURL).CombinedOutput(); err != nil {
			logrus.Errorf("unTar %v error: %v", busyboxTar, err)
		}
	}
}

func createWriteLayer(rootURL string) {
	writeURL := rootURL + "writeLayer/"
	if err := os.Mkdir(writeURL, 0777); err != nil {
		logrus.Errorf("mkdir %v error: %v", writeURL, err)
	}

	// work是overlay必须的，具体为什么？？？？ 暂时放这里吧
	workURL := rootURL + "work/"
	if err := os.Mkdir(workURL, 0777); err != nil {
		logrus.Errorf("mkdir %v error: %v", workURL, err)
	}
}

func createMountPoint(rootURL, mntURL string) {
	if err := os.Mkdir(mntURL, 0777); err != nil {
		logrus.Errorf("mkdir %v error: %v", mntURL, err)
	}

	// mount: unknown filesystem type 'aufs' aufs已经过时了，改成overlay
	// cat /proc/filesystems 查看支持的文件系统类型
	//
	// mount -t overlay overlay -o lowerdir=./lower,upperdir=./upper,workdir=./work ./merged

	// 把writeLayer目录和busybox目录 mount 到 mnt 目录
	// todo mount ？？？？？？
	dirs := fmt.Sprintf("lowerdir=%sbusybox,upperdir=%swriteLayer,workdir=%swork", rootURL, rootURL, rootURL)
	cmd := exec.Command("mount", "-t", "overlay", "overlay", "-o", dirs, mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logrus.Errorf("func[CreateMountPoint] %v", err)
	}
}

func pathExist(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// mountVolume 对用户要挂载进来的路径进行挂载
func mountVolume(rootURL, mntURL string, volumeURLs []string) {
	// 创建宿主机文件目录
	parentURL := volumeURLs[0]
	if err := os.Mkdir(parentURL, 0777); err != nil {
		logrus.Errorf("mkdir parent dir %v error: %v", parentURL, err)
	}

	// 在容器目录创建挂载点目录
	containerURL := path.Join(mntURL, volumeURLs[1])
	if err := os.Mkdir(containerURL, 0777); err != nil {
		logrus.Errorf("mkdir container dir %v error: %v", containerURL, err)
	}

	// 把宿主机文件目录挂在到容器内挂载点
	dirs := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%swork", containerURL, parentURL, rootURL)
	cmd := exec.Command("mount", "-t", "overlay", "overlay", "-o", dirs, containerURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logrus.Errorf("func[mountVolume] %v", err)
	}
}

// DeleteWorkSpace .
func DeleteWorkSpace(rootURL, mntURL, volume string) {
	if volume != "" {
		volumeURLs := strings.Split(volume, ":")
		if len(volumeURLs) == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
			deleteMountPointWithVolume(mntURL, volumeURLs)
			deleteWritePlayer(rootURL)
			return
		}
	}

	deleteMountPoint(mntURL)
	deleteWritePlayer(rootURL)
}

func deleteMountPoint(mntURL string) {
	cmd := exec.Command("umount", mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		logrus.Errorf("umount %v error: %v", mntURL, err)
	}

	if err := os.RemoveAll(mntURL); err != nil {
		logrus.Errorf("remove dir %v error: %v", mntURL, err)
	}
}

func deleteMountPointWithVolume(mntURL string, volumeURLs []string) {
	// 卸载容器里面volome挂载点的文件系统
	containerURL := path.Join(mntURL, volumeURLs[1])
	cmd := exec.Command("umount", containerURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logrus.Errorf("umount volume %v:%v error: %v", volumeURLs[0], containerURL, err)
	}

	// 卸载整个容器文件系统的挂载点
	deleteMountPoint(mntURL)
}

func deleteWritePlayer(rootURL string) {
	writeURL := rootURL + "writeLayer/"
	if err := os.RemoveAll(writeURL); err != nil {
		logrus.Errorf("remove dir %v error: %v", writeURL, err)
	}

	// work是overlay必须的，具体为什么？？？？
	workURL := rootURL + "work/"
	if err := os.RemoveAll(workURL); err != nil {
		logrus.Errorf("remove dir %v error: %v", workURL, err)
	}
}
