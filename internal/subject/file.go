package subject

import (
	"encoding/csv"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"strconv"
)

// 读取现有数据文件
func readExistingSubjects() ([]JsonSubject, error) {
	fileData, err := ioutil.ReadFile("data/anime.json")
	if err != nil {
		return nil, err
	}

	var existingList []JsonSubject
	if err := json.Unmarshal(fileData, &existingList); err != nil {
		return nil, err
	}
	return existingList, nil
}

func readExsitingStaffs() {

}

// ------------------------- 整理csv功能 -------------------------
func updateRemap(data []JsonSubject) {
	file, err := os.Create("data/anime_remap.csv")
	if err != nil {
		log.Fatalf("创建CSV文件失败: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 写入CSV头
	writer.Write([]string{"project_id", "original_id"})

	// 写入数据
	for _, item := range data {
		writer.Write([]string{strconv.Itoa(item.ProjectID), strconv.Itoa(item.OriginalID)})
	}
}
