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
	if err := ioutil.WriteFile("data/anime_lite.json", output, 0644); err != nil {
		log.Fatalf("文件写入失败: %v", err)
	}
	fmt.Printf("创建成功！共处理 %d 个条目\n", len(subjects))
}

func updateMode(ids []int, token string) {
	fileData, err := ioutil.ReadFile("data/anime_lite.json")
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
	if err := ioutil.WriteFile("data/anime_lite.json", output, 0644); err != nil {
		log.Fatalf("文件写入失败: %v", err)
	}
	fmt.Printf("更新成功！现有条目数: %d\n", len(existingList))
}

func createSubjectPerson(ids []int, token string) {
	// Read existing data
	existingList, err := readExistingData()
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
	if err := ioutil.WriteFile("data/subject_persons.json", output, 0644); err != nil {
		log.Fatalf("Failed to write file: %v", err)
	}
	fmt.Printf("Creation successful! Processed %d entries\n", len(subjectPersons))
}
func updateSubjectPerson(ids []int, token string) {
	// Read existing data
	fileData, err := ioutil.ReadFile("data/subject_persons.json")
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
	if err := ioutil.WriteFile("data/subject_persons.json", output, 0644); err != nil {
		log.Fatalf("Failed to write file: %v", err)
	}
	fmt.Printf("Update successful! Total entries: %d\n", len(existingList))
}
func fixProjectIDs() {
	// Read existing data
	existingList, err := readExistingData()
	if err != nil {
		log.Fatalf("Failed to read existing data: %v", err)
	}

	// Sort by original_id
	sort.Slice(existingList, func(i, j int) bool {
		return existingList[i].OriginalID < existingList[j].OriginalID
	})

	// Reassign project_id based on sorted order
	for i := range existingList {
		existingList[i].ProjectID = i + 1
	}

	// Write updated data back to JSON file
	output, err := json.MarshalIndent(existingList, "", "  ")
	if err != nil {
		log.Fatalf("Failed to generate JSON: %v", err)
	}
	if err := ioutil.WriteFile("data/anime_lite.json", output, 0644); err != nil {
		log.Fatalf("Failed to write file: %v", err)
	}
	fmt.Printf("Project IDs fixed successfully! Total entries: %d\n", len(existingList))

	// Update the remap CSV file
	updateRemap(existingList)
}
