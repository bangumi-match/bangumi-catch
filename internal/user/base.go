package user

import (
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"time"
)

// ------------------------- 工具函数 -------------------------
func initLog() (*os.File, error) {
	logDir := "logs"
	if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("创建日志目录失败: %v", err)
	}
	logFile, err := os.Create(fmt.Sprintf("%s/log_user_%s.txt", logDir, time.Now().Format("20060102_150405")))
	if err != nil {
		return nil, fmt.Errorf("创建日志文件失败: %v", err)
	}
	log.SetOutput(io.MultiWriter(logFile, os.Stdout))
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	return logFile, nil
}

func resolveUserID(userID int) (fetchID string, err error) { // 通过数字ID获取用户唯一urlID
	// 读取本地 JSON 文件
	user, err := readUserData(userID)
	if err != nil && os.IsNotExist(err) {
		return strconv.Itoa(userID), nil
	}

	// 如果输入是数字ID且本地没有设置字符串，尝试获取用户名
	if user.UserName == "" {
		user.UserName = getUserName(userID)
		if user.UserName == "" {
			// 获取不到用户名时仍使用数字ID
			return strconv.Itoa(userID), nil
		}
		return user.UserName, nil
	}
	return user.UserName, nil
}
