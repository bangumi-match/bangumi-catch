package main

import (
	"bgm-catch/internal/subject"
	"bgm-catch/internal/user"
	"flag"
	"fmt"
	"os"
	"strings"
)

func main() {
	// 解析命令行参数
	mode := flag.String("mode", "", "启动模式: subject 或 user")
	flag.Parse()

	// 检查环境变量
	if *mode == "" {
		*mode = os.Getenv("START_MODE")
	}

	// 如果没有通过命令行参数或环境变量指定模式，则询问用户
	if *mode == "" {
		fmt.Println("请选择下载模式: subject 或 user")
		fmt.Print("输入模式: ")
		fmt.Scanln(mode)
	}

	// 根据模式启动相应的模块
	switch strings.ToLower(*mode) {
	case "subject", "s":
		startSubjectModule()
	case "user", "u":
		startUserModule()
	default:
		fmt.Println("无效的模式，请选择 subject 或 user")
	}
}

func startSubjectModule() {
	fmt.Println("启动 subject 模块...")
	subject.Main()
}

func startUserModule() {
	fmt.Println("启动 user 模块...")
	user.Main()
}
