package user

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var userRemap []JsonUserFile

// ------------------------- 文件操作 -------------------------
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

// ------------------------- 文件操作 -------------------------

func init() {
	os.MkdirAll(usersDir, 0755)
}

func saveUserData(user JsonUserFile) error {
	user.CatchTime = time.Now().Format("2006-01-02 15:04:05")
	data, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("JSON序列化失败: %v", err)
	}
	return os.WriteFile(filepath.Join(usersDir, fmt.Sprintf("%d.json", user.UserID)), data, 0644)
}

func readUserData(userID int) (JsonUserFile, error) {
	data, err := os.ReadFile(filepath.Join(usersDir, fmt.Sprintf("%d.json", userID)))
	if err != nil {
		return JsonUserFile{}, err
	}

	var user JsonUserFile
	if err := json.Unmarshal(data, &user); err != nil {
		return user, err
	}
	return user, nil
}

func readExistingUserIDs() (map[int]struct{}, error) {
	entries, err := os.ReadDir(usersDir)
	if err != nil {
		return nil, err
	}

	ids := make(map[int]struct{})
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if strings.HasSuffix(entry.Name(), ".json") {
			idStr := entry.Name()[:len(entry.Name())-5]
			if id, err := strconv.Atoi(idStr); err == nil {
				ids[id] = struct{}{}
			}
		}
	}
	return ids, nil
}
