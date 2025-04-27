package subject

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
)

func createMode(ids []int, token string) {
	subjects := fetchByIdList(ids, token)
	projectID := 1

	for i := range subjects {
		subjects[i].ProjectID = projectID
		projectID++
	}

	if err := os.MkdirAll("data", os.ModePerm); err != nil {
		log.Fatalf("创建data目录失败: %v", err)
	}
	output, err := json.MarshalIndent(subjects, "", "  ")
	if err != nil {
		log.Fatalf("JSON生成失败: %v", err)
	}
	if err := os.WriteFile("data/anime.json", output, 0644); err != nil {
		log.Fatalf("文件写入失败: %v", err)
	}
	fmt.Printf("创建成功！共处理 %d 个条目\n", len(subjects))
}

func updateMode(ids []int, token string) {
	fileData, err := ioutil.ReadFile("data/anime.json")
	if err != nil {
		log.Fatalf("读取现有文件失败: %v", err)
	}

	var existingList []JsonSubject
	if err := json.Unmarshal(fileData, &existingList); err != nil {
		log.Fatalf("JSON解析失败: %v", err)
	}

	maxProjectID := 0
	for _, item := range existingList {
		if item.ProjectID > maxProjectID {
			maxProjectID = item.ProjectID
		}
	}

	newSubjects := fetchByIdList(ids, token)

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
	fmt.Printf("更新成功！现有条目数: %d\n", len(existingList))
}

func createSubjectPerson(ids []int, token string) {
	// Read existing data
	existingList, err := readExistingSubjects()
	if err != nil {
		log.Fatalf("Failed to read existing data: %v", err)
	}

	// Create a map for quick lookup of existing IDs and their project IDs
	existingIDMap := make(map[int]int)
	for _, item := range existingList {
		existingIDMap[item.OriginalID] = item.ProjectID
	}

	// Fetch subject persons by ID list
	subjectPersons := fetchPersonsByIdList(ids, token)

	// Check if all IDs have corresponding project IDs
	for i := range subjectPersons {
		if projectID, exists := existingIDMap[subjectPersons[i].OriginalID]; exists {
			subjectPersons[i].ProjectID = projectID
		} else {
			log.Fatalf("ID %d does not have a corresponding project ID. Please download the base data first.", subjectPersons[i].OriginalID)
		}
	}

	// Create data directory if it doesn't exist
	if err := os.MkdirAll("data", os.ModePerm); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	// Save the subject persons data to a JSON file
	output, err := json.MarshalIndent(subjectPersons, "", "  ")
	if err != nil {
		log.Fatalf("Failed to generate JSON: %v", err)
	}
	if err := os.WriteFile("data/anime_staffs.json", output, 0644); err != nil {
		log.Fatalf("Failed to write file: %v", err)
	}
	fmt.Printf("Creation successful! Processed %d entries\n", len(subjectPersons))
}

func updateSubjectPerson(ids []int, token string) {
	// Read existing data
	fileData, err := ioutil.ReadFile("data/anime_staffs.json")
	if err != nil {
		log.Fatalf("Failed to read existing file: %v", err)
	}

	var existingList []JsonSubjectPersonCollection
	if err := json.Unmarshal(fileData, &existingList); err != nil {
		log.Fatalf("Failed to parse JSON: %v", err)
	}

	// Create a map for quick lookup of existing IDs
	existingIDMap := make(map[int]*JsonSubjectPersonCollection)
	for i := range existingList {
		existingIDMap[existingList[i].OriginalID] = &existingList[i]
	}

	// Fetch subject persons by ID list
	newSubjectPersons := fetchPersonsByIdList(ids, token)

	// Update existing entries or add new ones
	for _, newSP := range newSubjectPersons {
		if existingSP, exists := existingIDMap[newSP.OriginalID]; exists {
			*existingSP = newSP // Update existing entry
		} else {
			existingList = append(existingList, newSP) // Add new entry
		}
	}

	// Save the updated subject persons data to a JSON file
	output, err := json.MarshalIndent(existingList, "", "  ")
	if err != nil {
		log.Fatalf("Failed to generate JSON: %v", err)
	}
	if err := os.WriteFile("data/anime_staffs.json", output, 0644); err != nil {
		log.Fatalf("Failed to write file: %v", err)
	}
	fmt.Printf("Update successful! Total entries: %d\n", len(existingList))
}

func fixProjectIDs() {
	// 读取并修复anime.json
	existingSubjectsList, err := readExistingSubjects()
	if err != nil {
		log.Fatalf("读取基础数据失败: %v", err)
	}

	// 创建OriginalID到ProjectID的映射表
	idMap := make(map[int]int)
	sort.Slice(existingSubjectsList, func(i, j int) bool {
		return existingSubjectsList[i].OriginalID < existingSubjectsList[j].OriginalID
	})
	for i := range existingSubjectsList {
		existingSubjectsList[i].ProjectID = i + 1 // 从1开始递增
		idMap[existingSubjectsList[i].OriginalID] = existingSubjectsList[i].ProjectID
	}

	// 写回anime.json
	output, _ := json.MarshalIndent(existingSubjectsList, "", "  ")
	if err := os.WriteFile("data/anime.json", output, 0644); err != nil {
		log.Fatalf("更新基础数据失败: %v", err)
	}

	// 处理staff数据
	if _, err := os.Stat("data/anime_staffs.json"); err == nil {
		fileData, _ := ioutil.ReadFile("data/anime_staffs.json")
		var staffs []JsonSubjectPersonCollection
		json.Unmarshal(fileData, &staffs)

		for i := range staffs {
			if projectID, exists := idMap[staffs[i].OriginalID]; exists {
				staffs[i].ProjectID = projectID
			}
		}
		output, _ := json.MarshalIndent(staffs, "", "  ")
		os.WriteFile("data/anime_staffs.json", output, 0644)
	}

	// 处理relation数据
	if _, err := os.Stat("data/anime_relations.json"); err == nil {
		fileData, _ := ioutil.ReadFile("data/anime_relations.json")
		var relations []JsonSubjectRelationCollection
		json.Unmarshal(fileData, &relations)

		for i := range relations {
			if projectID, exists := idMap[relations[i].OriginalID]; exists {
				relations[i].ProjectID = projectID
			}
		}
		output, _ := json.MarshalIndent(relations, "", "  ")
		os.WriteFile("data/anime_relations.json", output, 0644)
	}

	fmt.Printf("重新映射完成！总条目数: %d\n", len(existingSubjectsList))
	updateRemap(existingSubjectsList)
}

// 新增创建关系数据函数
func createSubjectRelations(ids []int, token string) {
	existingList, err := readExistingSubjects()
	if err != nil {
		log.Fatalf("读取基础数据失败: %v", err)
	}

	existingIDMap := make(map[int]int)
	for _, item := range existingList {
		existingIDMap[item.OriginalID] = item.ProjectID
	}

	subjectRelations := fetchRelationsByIdList(ids, token)

	for i := range subjectRelations {
		if projectID, exists := existingIDMap[subjectRelations[i].OriginalID]; exists {
			subjectRelations[i].ProjectID = projectID
		} else {
			log.Fatalf("ID %d 没有对应的project ID，请先下载基础数据", subjectRelations[i].OriginalID)
		}
	}

	output, err := json.MarshalIndent(subjectRelations, "", "  ")
	if err != nil {
		log.Fatalf("JSON生成失败: %v", err)
	}
	if err := os.WriteFile("data/anime_relations.json", output, 0644); err != nil {
		log.Fatalf("文件写入失败: %v", err)
	}
	fmt.Printf("关系数据创建成功！共处理 %d 个条目\n", len(subjectRelations))
}

func updateSubjectRelations(ids []int, token string) {
	fileData, err := ioutil.ReadFile("data/anime_relations.json")
	if err != nil {
		log.Fatalf("读取关系数据失败: %v", err)
	}

	var existingList []JsonSubjectRelationCollection
	if err := json.Unmarshal(fileData, &existingList); err != nil {
		log.Fatalf("JSON解析失败: %v", err)
	}

	existingIDMap := make(map[int]*JsonSubjectRelationCollection)
	for i := range existingList {
		existingIDMap[existingList[i].OriginalID] = &existingList[i]
	}

	newRelations := fetchRelationsByIdList(ids, token)

	for _, newRel := range newRelations {
		if existingRel, exists := existingIDMap[newRel.OriginalID]; exists {
			*existingRel = newRel
		} else {
			existingList = append(existingList, newRel)
		}
	}

	output, err := json.MarshalIndent(existingList, "", "  ")
	if err != nil {
		log.Fatalf("JSON生成失败: %v", err)
	}
	if err := os.WriteFile("data/anime_relations.json", output, 0644); err != nil {
		log.Fatalf("文件写入失败: %v", err)
	}
	fmt.Printf("关系数据更新成功！现有条目数: %d\n", len(existingList))
}
