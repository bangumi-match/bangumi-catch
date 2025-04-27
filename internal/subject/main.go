package subject

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	. "bgm-catch/internal/basic"
)

const tips = "请选择以下选项：\n" +
	"CA（创建动画）\n" +
	"UA（更新动画）\n" +
	"DA（日期范围更新/下载动画）\n" +
	"R（重新映射ID）\n" +
	"CS（下载Staff）\n" +
	"AS（根据Anime Lite下载全部对应Staff）\n" +
	"US（使用动画ID更新Staff）\n" +
	"CR（使用动画ID下载关系数据）\n" +
	"AR（下载全部动画的关系数据）\n" +
	"UR（更新关系数据）"

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

	fmt.Print(tips)
	mode, _ := reader.ReadString('\n')
	mode = strings.TrimSpace(mode)

	switch strings.ToUpper(mode) {
	case "CA", "CREATE", "CREATE_ANIME":
		// 创建模式处理
		fmt.Print("请输入ID列表（例如：1,2,5-10,12）: ")
		idInput, _ := reader.ReadString('\n')
		idInput = strings.TrimSpace(idInput)

		ids, err := ParseIDList(idInput)
		if err != nil {
			log.Fatalf("ID列表解析失败: %v", err)
		}
		createMode(ids, token)
		existingList, err := readExistingSubjects()
		if err != nil {
			log.Fatalf("读取现有数据失败: %v", err)
		}
		updateRemap(existingList)

	case "UA", "UPDATE", "UPDATE_ANIME":
		// 更新模式处理
		fmt.Print("请输入ID列表（例如：1,2,5-10,12）或输入'all'更新全部条目: ")
		idInput, _ := reader.ReadString('\n')
		idInput = strings.TrimSpace(idInput)

		var ids []int
		if strings.ToLower(idInput) == "all" {
			// 读取现有数据获取所有ID
			existingList, err := readExistingSubjects()
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
		existingList, err := readExistingSubjects()
		if err != nil {
			log.Fatalf("读取现有数据失败: %v", err)
		}
		updateRemap(existingList)

	case "DA", "DATE", "DATE_ANIME":
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
		existingList, err := readExistingSubjects()
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
		if err := os.WriteFile("data/anime.json", output, 0644); err != nil {
			log.Fatalf("文件写入失败: %v", err)
		}
		fmt.Printf("日期范围更新成功！现有条目数: %d\n", len(existingList))
		updateRemap(existingList)
	case "R", "REMAP":
		// Fix project IDs mode
		fixProjectIDs()
		existingList, err := readExistingSubjects()
		if err != nil {
			log.Fatalf("读取现有数据失败: %v", err)
		}
		updateRemap(existingList)

	case "CS", "CREATE_STAFF":
		// 下载Person数据
		fmt.Print("请输入下载Staff的动画ID列表（例如：1,2,5-10,12）: ")
		idInput, _ := reader.ReadString('\n')
		idInput = strings.TrimSpace(idInput)

		ids, err := ParseIDList(idInput)
		if err != nil {
			log.Fatalf("ID列表解析失败: %v", err)
		}
		createSubjectPerson(ids, token)

	case "AS", "ALL_STAFF":
		// 根据Anime Lite下载Person数据
		existingList, err := readExistingSubjects()
		if err != nil {
			log.Fatalf("读取现有数据失败: %v", err)
		}

		var ids []int
		for _, item := range existingList {
			ids = append(ids, item.OriginalID)
		}
		createSubjectPerson(ids, token)

	case "US", "UPDATE_STAFF":
		// 更新Person数据
		fmt.Print("请输入ID列表（例如：1,2,5-10,12）: ")
		idInput, _ := reader.ReadString('\n')
		idInput = strings.TrimSpace(idInput)

		ids, err := ParseIDList(idInput)
		if err != nil {
			log.Fatalf("ID列表解析失败: %v", err)
		}
		updateSubjectPerson(ids, token)

	case "CR", "CREATE_RELATION":
		fmt.Print("请输入ID列表（例如：1,2,5-10,12）: ")
		idInput, _ := reader.ReadString('\n')
		idInput = strings.TrimSpace(idInput)

		ids, err := ParseIDList(idInput)
		if err != nil {
			log.Fatalf("ID列表解析失败: %v", err)
		}
		createSubjectRelations(ids, token)
	case "AR", "ALL_RELATIONS":
		existingList, err := readExistingSubjects()
		if err != nil {
			log.Fatalf("读取基础数据失败: %v", err)
		}

		var ids []int
		for _, item := range existingList {
			ids = append(ids, item.OriginalID)
		}
		createSubjectRelations(ids, token)
	case "UR", "UPDATE_RELATION":
		fmt.Print("请输入ID列表（例如：1,2,5-10,12）: ")
		idInput, _ := reader.ReadString('\n')
		idInput = strings.TrimSpace(idInput)

		ids, err := ParseIDList(idInput)
		if err != nil {
			log.Fatalf("ID列表解析失败: %v", err)
		}
		updateSubjectRelations(ids, token)

	default:
		println("无效模式选择，请选择以下选项：\n%s", tips)
		log.Fatal("无效模式选择")
	}
}
