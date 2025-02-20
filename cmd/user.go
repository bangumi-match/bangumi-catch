package main

import (
	"encoding/json"
	"fmt"
	"github.com/gocolly/colly/v2"
	"log"
	"os"
)

type Tag struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type Subject struct {
	Name         string `json:"name"`
	NameCn       string `json:"name_cn"`
	ShortSummary string `json:"short_summary"`
	Tags         []Tag  `json:"tags"`
	Images       struct {
		Small  string `json:"small"`
		Grid   string `json:"grid"`
		Large  string `json:"large"`
		Medium string `json:"medium"`
		Common string `json:"common"`
	} `json:"images"`
	Score           float64 `json:"score"`
	Id              int     `json:"id"`
	Eps             int     `json:"eps"`
	Volumes         int     `json:"volumes"`
	CollectionTotal int     `json:"collection_total"`
	Rank            int     `json:"rank"`
}

type Collection struct {
	UpdatedAt   string   `json:"updated_at"`
	Comment     string   `json:"comment"`
	Tags        []string `json:"tags"`
	Subject     Subject  `json:"subject"`
	SubjectId   int      `json:"subject_id"`
	VolStatus   int      `json:"vol_status"`
	EpStatus    int      `json:"ep_status"`
	SubjectType int      `json:"subject_type"`
	Type        int      `json:"type"`
	Rate        int      `json:"rate"`
	Private     bool     `json:"private"`
}

type ApiResponse struct {
	Data   []Collection `json:"data"`
	Total  int          `json:"total"`
	Limit  int          `json:"limit"`
	Offset int          `json:"offset"`
}

func fetchData(userId string, collectionType int) ([]Collection, error) {
	var result []Collection
	offset := 0
	limit := 50

	// Create the collector
	c := colly.NewCollector()

	// Construct the API URL
	url := fmt.Sprintf("https://api.bgm.tv/v0/users/%s/collections?subject_type=2&type=%d&limit=%d&offset=%d", userId, collectionType, limit, offset)

	// API response handling
	c.OnResponse(func(r *colly.Response) {
		var response ApiResponse
		err := json.Unmarshal(r.Body, &response)
		if err != nil {
			log.Println("Error unmarshaling JSON:", err)
			return
		}

		// Append fetched data
		result = append(result, response.Data...)

		// If there are more pages, recursively fetch the next page
		if len(response.Data) > 0 && response.Offset+limit < response.Total {
			offset = response.Offset + limit
			nextPageUrl := fmt.Sprintf("https://api.bgm.tv/v0/users/%s/collections?subject_type=2&type=%d&limit=%d&offset=%d", userId, collectionType, limit, offset)
			err = c.Visit(nextPageUrl)
			if err != nil {
				log.Println("Error fetching next page:", err)
			}
		}
	})

	// Start scraping
	err := c.Visit(url)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func saveToFile(userId string, collectionType int, data []Collection) error {
	// Create directory for user
	dir := fmt.Sprintf("user_data/%s", userId)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}

	// Map collection type to Chinese name
	typeNames := map[int]string{
		1: "想看",
		2: "看过",
		3: "在看",
		4: "搁置",
		5: "抛弃",
	}

	// Create file
	fileName := fmt.Sprintf("%s/%s.json", dir, typeNames[collectionType])
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	// Convert to JSON and write to file
	jsonData, err := json.MarshalIndent(map[string][]Collection{"data": data}, "", "  ")
	if err != nil {
		return err
	}

	_, err = file.Write(jsonData)
	return err
}

func main() {
	var userId string
	fmt.Print("请输入用户ID: ")
	fmt.Scan(&userId)

	// Fetching "想看" (type=1)
	collections, err := fetchData(userId, 1)
	if err != nil {
		log.Fatal("Error fetching data:", err)
	}

	// Save to file
	err = saveToFile(userId, 1, collections)
	if err != nil {
		log.Fatal("Error saving data:", err)
	}

	// Repeat for other collection types (2-5)
	for i := 2; i <= 5; i++ {
		collections, err = fetchData(userId, i)
		if err != nil {
			log.Fatal("Error fetching data:", err)
		}

		// Save to file
		err = saveToFile(userId, i, collections)
		if err != nil {
			log.Fatal("Error saving data:", err)
		}
	}

	fmt.Println("数据已保存！")
}
