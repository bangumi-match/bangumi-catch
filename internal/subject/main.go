package subject

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	. "bgm-catch/internal/basic"
)

func Main() {
	logFile, err := initLog()
	if err != nil {
		log.Fatalf("初始化日志失败: %v", err)
	}
	defer logFile.Close()

	token := os.Getenv("TOKEN")
	if token == "" {
		log.Println("警告：未设置TOKEN环境变量，可能无法获取完整数据")
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("请选择模式(C=创建 / U=更新 / D=日期范围更新 / R=重新映射 / F=Fix Project IDs / P=下载Person / AP=根据Anime Lite下载Person / UP=update person): ")
	mode, _ := reader.ReadString('\n')
	mode = strings.TrimSpace(mode)

	switch strings.ToUpper(mode) {
	case "C", "CREATE":
		// 创建模式处理
		fmt.Print("请输入ID列表（例如：1,2,5-10,12）: ")
		idInput, _ := reader.ReadString('\n')
		idInput = strings.TrimSpace(idInput)

		ids, err := ParseIDList(idInput)
		if err != nil {
			log.Fatalf("ID列表解析失败: %v", err)
		}
		createMode(ids, token)
		existingList, err := readExistingData()
		if err != nil {
			log.Fatalf("读取现有数据失败: %v", err)
		}
		updateRemap(existingList)

	case "U", "UPDATE":
		// 更新模式处理
		fmt.Print("请输入ID列表（例如：1,2,5-10,12）或输入'all'更新全部条目: ")
		idInput, _ := reader.ReadString('\n')
		idInput = strings.TrimSpace(idInput)

		var ids []int
		if strings.ToLower(idInput) == "all" {
			// 读取现有数据获取所有ID
			existingList, err := readExistingData()
			if err != nil {
				log.Fatalf("读取现有数据失败: %v", err)
			}
			for _, item := range existingList {
				ids = append(ids, item.OriginalID)
			}
		} else {
			// 正常解析ID列表
			parsedIDs, err := ParseIDList(idInput)
			if err != nil {
				log.Fatalf("ID列表解析失败: %v", err)
			}
			ids = parsedIDs
		}
		updateMode(ids, token)
		existingList, err := readExistingData()
		if err != nil {
			log.Fatalf("读取现有数据失败: %v", err)
		}
		updateRemap(existingList)

	case "D", "DATE":
		fmt.Print("请输入起始年月（格式：YYYY-MM）: ")
		startInput, _ := reader.ReadString('\n')
		startInput = strings.TrimSpace(startInput)

		fmt.Print("请输入结束年月（格式：YYYY-MM）: ")
		endInput, _ := reader.ReadString('\n')
		endInput = strings.TrimSpace(endInput)

		dates, err := parseDateRange(startInput, endInput)
		if err != nil {
			log.Fatalf("日期解析失败: %v", err)
		}

		fmt.Println("开始抓取日期范围数据...")
		newSubjects := fetchByDateRange(dates, token)

		// 合并到现有数据
		existingList, err := readExistingData()
		if err != nil {
			log.Fatalf("读取现有数据失败: %v", err)
		}

		maxProjectID := 0
		for _, item := range existingList {
			if item.ProjectID > maxProjectID {
				maxProjectID = item.ProjectID
			}
		}

		for _, newSubj := range newSubjects {
			found := false
			for i := range existingList {
				if existingList[i].OriginalID == newSubj.OriginalID {
					updateExistingFields(&existingList[i], &newSubj)
					found = true
					break
				}
			}

			if !found {
				maxProjectID++
				newSubj.ProjectID = maxProjectID
				existingList = append(existingList, newSubj)
			}
		}

		output, err := json.MarshalIndent(existingList, "", "  ")
		if err != nil {
			log.Fatalf("JSON生成失败: %v", err)
		}
		if err := ioutil.WriteFile("data/anime_lite.json", output, 0644); err != nil {
			log.Fatalf("文件写入失败: %v", err)
		}
		fmt.Printf("日期范围更新成功！现有条目数: %d\n", len(existingList))
		updateRemap(existingList)

	case "F", "FIX":
		// Fix project IDs mode
		fixProjectIDs()
		existingList, err := readExistingData()
		if err != nil {
			log.Fatalf("读取现有数据失败: %v", err)
		}
		updateRemap(existingList)

	case "R", "REMAP":
		// 重新映射模式处理
		existingList, err := readExistingData()
		if err != nil {
			log.Fatalf("读取现有数据失败: %v", err)
		}
		updateRemap(existingList)

	case "P", "PERSON":
		// 下载Person数据
		fmt.Print("请输入ID列表（例如：1,2,5-10,12）: ")
		idInput, _ := reader.ReadString('\n')
		idInput = strings.TrimSpace(idInput)

		ids, err := ParseIDList(idInput)
		if err != nil {
			log.Fatalf("ID列表解析失败: %v", err)
		}
		createSubjectPerson(ids, token)

	case "AP", "ANIME_PERSON":
		// 根据Anime Lite下载Person数据
		existingList, err := readExistingData()
		if err != nil {
			log.Fatalf("读取现有数据失败: %v", err)
		}

		var ids []int
		for _, item := range existingList {
			ids = append(ids, item.OriginalID)
		}
		createSubjectPerson(ids, token)

	case "UP", "UPDATE_PERSON":
		// 更新Person数据
		fmt.Print("请输入ID列表（例如：1,2,5-10,12）: ")
		idInput, _ := reader.ReadString('\n')
		idInput = strings.TrimSpace(idInput)

		ids, err := ParseIDList(idInput)
		if err != nil {
			log.Fatalf("ID列表解析失败: %v", err)
		}
		updateSubjectPerson(ids, token)

	default:
		log.Fatal("无效模式选择，请选择C（创建）/U（更新）/D（日期范围更新）/R（重新映射）/F（Fix Project IDs）/P（下载Person）/AP（根据Anime Lite下载Person）/UP（更新Person）")
	}
}
