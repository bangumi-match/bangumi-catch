package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Choose mode of operation: organize or find-missing")
	mode, _ := reader.ReadString('\n')
	mode = strings.TrimSpace(mode)

	switch mode {
	case "organize", "o":
		fmt.Println("Enter the number of files per directory:")
		intervalStr, _ := reader.ReadString('\n')
		interval, err := strconv.Atoi(strings.TrimSpace(intervalStr))
		if err != nil {
			log.Fatalf("Invalid interval: %v", err)
		}
		organizeFiles(interval)
	case "find-missing", "f":
		findMissingFiles()
	default:
		log.Fatalf("Unknown mode: %s", mode)
	}
}

func organizeFiles(interval int) {
	sourceDir := "data"
	destDir := "subject_data"

	files, err := ioutil.ReadDir(sourceDir)
	if err != nil {
		log.Fatalf("Failed to read directory %s: %v", sourceDir, err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		id, err := strconv.Atoi(file.Name()[:len(file.Name())-5])
		if err != nil {
			log.Printf("Skipping file %s: %v", file.Name(), err)
			continue
		}

		subDir := fmt.Sprintf("%d", id/interval)
		destPath := filepath.Join(destDir, subDir)

		if err := os.MkdirAll(destPath, os.ModePerm); err != nil {
			log.Printf("Failed to create directory %s: %v", destPath, err)
			continue
		}

		srcFile := filepath.Join(sourceDir, file.Name())
		destFile := filepath.Join(destPath, file.Name())
		if err := os.Rename(srcFile, destFile); err != nil {
			log.Printf("Failed to move file %s to %s: %v", srcFile, destFile, err)
		}
	}

	fmt.Println("Data has been organized into subject_data directory.")
}

func findMissingFiles() {
	destDir := "subject_data"
	var allFiles []string

	err := filepath.Walk(destDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".json") {
			allFiles = append(allFiles, info.Name())
		}
		return nil
	})
	if err != nil {
		log.Fatalf("Failed to walk through directory %s: %v", destDir, err)
	}

	idMap := make(map[int]bool)
	for _, file := range allFiles {
		id, err := strconv.Atoi(file[:len(file)-5])
		if err != nil {
			log.Printf("Skipping file %s: %v", file, err)
			continue
		}
		idMap[id] = true
	}

	var missingIDs []int
	for i := 1; i <= len(idMap); i++ {
		if !idMap[i] {
			missingIDs = append(missingIDs, i)
		}
	}

	csvFile, err := os.Create("missing.csv")
	if err != nil {
		log.Fatalf("Failed to create missing.csv: %v", err)
	}
	defer csvFile.Close()

	writer := csv.NewWriter(csvFile)
	defer writer.Flush()

	for _, id := range missingIDs {
		err := writer.Write([]string{strconv.Itoa(id)})
		if err != nil {
			log.Fatalf("Failed to write to missing.csv: %v", err)
		}
	}

	fmt.Println("Missing IDs have been written to missing.csv.")
}
