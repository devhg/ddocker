package enterns

/*
#define _GNU_SOURCE
#include <unistd.h>
#include <errno.h>
#include <sched.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <fcntl.h>

// __attribute__((constructor))指的是，一旦这个包被引用，那么这个函数
// 就会被自动执行，类似构造函数，会在程序启动时执行
__attribute__((constructor)) void enter_namespace(void) {
	// 从环境变量获取 需要进入的 PID 和 command
	char *ddocker_pid;
	ddocker_pid = getenv("ddocker_pid");
	if (ddocker_pid) {
		// fprintf(stdout, "got ddocker_pid=%s\n", ddocker_pid);
	} else {
		// fprintf(stdout, "missing ddocker_pid env skip enterns");
		return;
	}

	char *ddocker_cmd;
	ddocker_cmd = getenv("ddocker_cmd");
	if (ddocker_cmd) {
		// fprintf(stdout, "got ddocker_cmd=%s\n", ddocker_cmd);
	} else {
		// fprintf(stdout, "missing ddocker_cmd env skip enterns");
		return;
	}

	int i;
	char nspath[1024];
	char *namespaces[] = { "ipc", "uts", "net", "pid", "mnt" };
	for (i = 0; i < 5; i++) {
		sprintf(nspath, "/proc/%s/ns/%s", ddocker_pid, namespaces[i]);
		int fd = open(nspath, O_RDONLY);
		// 调用setns系统调用，进入对应的 namespace
		if (setns(fd, 0) == -1) {
			// fprintf(stderr, "setns on %s namespace failed: %s\n", namespaces[i], strerror(errno));
		} else {
			// fprintf(stdout, "setns on %s namespace succeeded\n", namespaces[i]);
		}
		close(fd);
	}
	int res = system(ddocker_cmd);
	exit(0);
	return;
}
*/
import "C"
