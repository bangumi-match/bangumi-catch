package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/gocolly/colly/v2"
	"github.com/schollz/progressbar/v3"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"
)

type ResponseData map[string]interface{}

func fetchAndSave(id int, wg *sync.WaitGroup, logMutex *sync.Mutex, bar *progressbar.ProgressBar, token string) {
	defer wg.Done()

	url := fmt.Sprintf("https://api.bgm.tv/v0/subjects/%d", id)
	c := colly.NewCollector()

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Authorization", "Bearer "+token)
	})

	c.OnResponse(func(r *colly.Response) {
		var data ResponseData
		if err := json.Unmarshal(r.Body, &data); err != nil {
			logMutex.Lock()
			log.Printf("Failed to parse JSON for ID %d: %v", id, err)
			logMutex.Unlock()
			return
		}
		if title, ok := data["title"].(string); ok && title == "Not Found" {
			logMutex.Lock()
			log.Printf("ID %d not found", id)
			logMutex.Unlock()
			return
		}
		data["catch_time"] = time.Now().Format(time.RFC3339)
		formattedJSON, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			logMutex.Lock()
			log.Printf("Failed to format JSON for ID %d: %v", id, err)
			logMutex.Unlock()
			return
		}
		dir := fmt.Sprintf("subject_data/%d", id/100)
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			logMutex.Lock()
			log.Printf("Failed to create directory: %v", err)
			logMutex.Unlock()
			return
		}
		filename := fmt.Sprintf("%s/%d.json", dir, id)
		if err := ioutil.WriteFile(filename, formattedJSON, 0644); err != nil {
			logMutex.Lock()
			log.Printf("Failed to write file %s: %v", filename, err)
			logMutex.Unlock()
		}
		bar.Add(1)
	})

	c.OnError(func(r *colly.Response, err error) {
		logMutex.Lock()
		log.Printf("Error fetching ID %d: %v", id, err)
		logMutex.Unlock()
	})

	c.Visit(url)
}
func doWork(ids []int, token string) {
	dir := "logs"
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		log.Fatalf("Failed to create logs directory: %v", err)
	}
	logFile, err := os.Create(fmt.Sprintf("%s/log_subject_%s.txt", dir, time.Now().Format("20060102_150405")))
	if err != nil {
		log.Fatalf("Failed to create log file: %v", err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	numCPU := runtime.NumCPU()
	log.Printf("Using %d threads", numCPU)
	var wg sync.WaitGroup
	var logMutex sync.Mutex
	sem := make(chan struct{}, numCPU)

	totalTasks := len(ids)
	bar := progressbar.Default(int64(totalTasks))

	for _, id := range ids {
		sem <- struct{}{} // 控制并发
		wg.Add(1)
		go func(i int) {
			defer func() { <-sem }()
			fetchAndSave(i, &wg, &logMutex, bar, token)
		}(id)
	}
	wg.Wait()
}

func main() {
	var token string

	token = os.Getenv("TOKEN")
	if token == "" {
		log.Printf("set env TOKEN to get full access of subjects")
	}

	var ids []int
	if _, err := os.Stat("missing.csv"); err == nil {
		var useCSV string
		fmt.Print("missing.csv found. Do you want to use the IDs from the CSV file? (y/n): ")
		fmt.Scan(&useCSV)
		if useCSV == "y" {
			file, err := os.Open("missing.csv")
			if err != nil {
				log.Fatalf("Failed to open missing.csv: %v", err)
			}
			defer file.Close()

			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				id, err := strconv.Atoi(scanner.Text())
				if err != nil {
					log.Printf("Invalid ID in CSV: %v", err)
					continue
				}
				ids = append(ids, id)
			}
			if err := scanner.Err(); err != nil {
				log.Fatalf("Error reading missing.csv: %v", err)
			}
		}
	}

	if len(ids) == 0 {
		var startID, endID int
		fmt.Print("Enter start ID: ")
		fmt.Scan(&startID)
		fmt.Print("Enter end ID: ")
		fmt.Scan(&endID)

		for i := startID; i <= endID; i++ {
			ids = append(ids, i)
		}
	}

	doWork(ids, token)
}
