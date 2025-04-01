package user

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strconv"
)

// ------------------------- 文件操作 -------------------------
func readExistingData() ([]JsonUserFile, error) {
	data, err := os.ReadFile(userOutputFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []JsonUserFile{}, nil
		}
		return nil, err
	}

	var users []JsonUserFile
	if err := json.Unmarshal(data, &users); err != nil {
		return nil, err
	}
	return users, nil
}

func loadAnimeMap() error {
	file, err := os.Open(animeMapFile)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	animeIDMap = make(map[int]int)

	// 跳过标题行
	if _, err := reader.Read(); err != nil {
		return err
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		projectID, _ := strconv.Atoi(record[0])
		originalID, _ := strconv.Atoi(record[1])
		animeIDMap[originalID] = projectID
	}
	return nil
}

func saveUserData(users []JsonUserFile) error {
	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON序列化失败: %v", err)
	}

	if err := os.WriteFile(userOutputFile, data, 0644); err != nil {
		return fmt.Errorf("文件写入失败: %v", err)
	}
	return nil
}

func generateUserMap() {
	// 读取现有的 JSON 数据
	users, err := readExistingData()
	if err != nil {
		log.Fatal(err)
	}

	// 按照 UserID 排序
	sort.Slice(users, func(i, j int) bool {
		return users[i].UserID < users[j].UserID
	})

	// 重新赋予 project_id
	for i := range users {
		users[i].ProjectID = i + 1
	}

	// 将更新后的数据保存回 JSON 文件
	saveUserData(users)

	// 生成 CSV 文件
	file, err := os.Create(userMapFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	writer.Write([]string{"project_id", "user_id"})

	for _, u := range users {
		writer.Write([]string{
			strconv.Itoa(u.ProjectID),
			strconv.Itoa(u.UserID),
		})
	}
	writer.Flush()
}
