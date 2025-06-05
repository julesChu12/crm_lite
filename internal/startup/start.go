package startup

import (
	"fmt"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
)

func Start(router *gin.Engine) {
	// 将启动的 pid 写入到文件中
	pid := os.Getpid()
	pidFile := fmt.Sprintf("/tmp/crm_lite_%d.pid", pid)
	os.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0644)
}
