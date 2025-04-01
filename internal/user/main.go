package user

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	. "bgm-catch/internal/basic"
)

var (
	animeIDMap map[int]int
	userIDMap  map[int]int
)

func Main() {
	logFile, err := initLog()
	if err != nil {
		log.Fatalf("初始化日志失败: %v", err)
	}
	defer logFile.Close()

	if err := loadAnimeMap(); err != nil {
		log.Fatalf("加载动画映射表失败: %v", err)
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("请选择模式(C=创建 / U=更新 / A=添加 / R=重新映射): ")
	mode, _ := reader.ReadString('\n')
	mode = strings.ToUpper(strings.TrimSpace(mode))

	switch mode {
	case "C":
		fmt.Print("请输入用户ID或范围（例如：1001 或 1001-2000）: ")
		input, _ := reader.ReadString('\n')
		userIDs, err := ParseIDList(strings.TrimSpace(input))
		if err != nil {
			log.Fatal("输入解析失败:", err)
		}
		createMode(userIDs)

	case "U":
		fmt.Print("请输入要更新的用户ID或范围（输入'all'更新所有用户，输入'empty'更新所有Data为空的用户）: ")
		input, _ := reader.ReadString('\n')
		input = strings.ToUpper(strings.TrimSpace(input))
		var userIDs []int
		if strings.Contains(input, "ALL") {
			userIDs, err = getAllUserIDs()
			if err != nil {
				log.Fatal("获取所有用户ID失败:", err)
			}
		} else if strings.Contains(input, "EMPTY") {
			userIDs, err = getUsersWithEmptyData()
			if err != nil {
				log.Fatal("获取Data为空的用户ID失败:", err)
			}
		} else {
			userIDs, err = ParseIDList(input)
			if err != nil {
				log.Fatal("输入解析失败:", err)
			}
		}
		updateMode(userIDs)
	case "A":
		fmt.Print("请输入要添加的用户ID或范围: ")
		input, _ := reader.ReadString('\n')
		userIDs, err := ParseIDList(strings.TrimSpace(input))
		if err != nil {
			log.Fatal("输入解析失败:", err)
		}
		addMode(userIDs)
	case "R":
		generateUserMap()
		fmt.Println("用户映射表已重新生成")
	default:
		log.Fatal("无效模式选择")
	}
}
