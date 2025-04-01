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

func resolveUserID(userID string) (fetchID string, userName string, err error) { // 通过数字ID获取用户唯一urlID
	// 读取本地 JSON 文件
	users, err := readExistingData()
	if err != nil {
		return "", "", fmt.Errorf("读取本地数据失败: %v", err)
	}

	// 检查本地是否存在该用户ID
	for _, user := range users {
		if strconv.Itoa(user.UserID) == userID {
			if user.UserName != "" {
				return userID, user.UserName, nil
			}
			break
		}
	}

	// 如果输入是数字ID，尝试获取用户名
	if _, err := strconv.Atoi(userID); err == nil {
		// 重试机制
		for attempt := 0; attempt < 3; attempt++ {
			userName = getUserName(userID)
			if userName != "" {
				return userName, userName, nil
			}
			time.Sleep(500 * time.Millisecond)
		}
		// 获取不到用户名时仍使用数字ID
		return userID, "", nil
	}
	// 如果输入本身就是用户名，直接使用
	return userID, userID, nil
}
