package main

import (
	"encoding/json"
	"fmt"
	"github.com/gocolly/colly/v2"
	"github.com/schollz/progressbar/v3"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
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
	limit := 20

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

func fetchDataWithRetry(userId string, collectionType int, wg *sync.WaitGroup, bar *progressbar.ProgressBar) ([]Collection, error) {
	defer wg.Done() // 减少计数器

	// 尝试最多重试一次
	for attempt := 0; attempt < 2; attempt++ {
		// 调用原 fetchData 函数
		data, err := fetchData(userId, collectionType)
		if err == nil {
			// 更新进度条
			bar.Add(1)
			return data, nil
		}
		// 如果是第一次失败
		if attempt == 0 {
		} else {
			// 如果是第二次失败，记录错误并跳过
			log.Printf("获取数据失败，用户ID: %s，获取类型: %d，已跳过此数据获取\n", userId, collectionType)
			return nil, err
		}
	}
	return nil, fmt.Errorf("获取数据失败，用户ID: %s，获取类型: %d", userId, collectionType)
}

func saveToFile(userId int, userName string, collectionType int, data []Collection) error {
	// Create directory for user
	if data == nil || len(data) == 0 {
		return nil
	}
	var dir string
	if userName == "" {
		dir = fmt.Sprintf("user_data/%d/%d", userId/100, userId)
	} else {
		dir = fmt.Sprintf("user_data/%d/%d_%s", userId/100, userId, userName)
	}
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}

	// Map collection type to Chinese name
	typeNames := map[int]string{
		1: "1_wish",
		2: "2_collect",
		3: "3_doing",
		4: "4_onhold",
		5: "5_dropped",
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

func getUserName(userId string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://bgm.tv/user/%s", userId))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.Request.URL.Path != fmt.Sprintf("/user/%s", userId) {
		parts := strings.Split(resp.Request.URL.Path, "/")
		if len(parts) > 2 {
			return parts[2], nil
		}
	}
	return "", nil
}
func initLogger() {
	// 确保 logs 目录存在
	os.MkdirAll("logs", os.ModePerm)

	// 生成日志文件名
	logFileName := fmt.Sprintf("logs/log_user_%s.txt", time.Now().Format("20060102_150405"))
	logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("无法创建日志文件: %v", err)
	}

	// 设置日志输出到文件和控制台
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWriter)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}
func main() {
	initLogger()
	log.Println("程序启动...")

	var inputType string
	fmt.Print("请输入操作类型（single 或 range）: ")
	fmt.Scan(&inputType)

	var userIds []string

	if inputType == "single" || inputType == "s" {
		var userId string
		fmt.Print("请输入用户ID (数字或名称): ")
		fmt.Scan(&userId)
		userIds = append(userIds, userId)
	} else if inputType == "range" || inputType == "r" {
		var startId, endId int
		fmt.Print("请输入起始ID: ")
		fmt.Scan(&startId)
		fmt.Print("请输入结束ID: ")
		fmt.Scan(&endId)

		for id := startId; id <= endId; id++ {
			userIds = append(userIds, strconv.Itoa(id))
		}
	} else {
		log.Println("无效的输入类型，请输入 'single' 或 'range'.")
		return
	}

	// 获取用户名并获取数据
	var wg sync.WaitGroup
	totalTasks := len(userIds) * 5 // 每个用户需要5次请求
	bar := progressbar.NewOptions(totalTasks,
		progressbar.OptionSetDescription("抓取数据中..."),
		progressbar.OptionSetWidth(50),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "#",
			SaucerHead:    "|",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))

	for _, userId := range userIds {
		var userName string
		var userIdNumer int
		var fetchId string
		if _, err := strconv.Atoi(userId); err == nil {
			userIdNumer, _ = strconv.Atoi(userId)
			var err error
			userName, err = getUserName(userId)
			if err != nil {
				log.Fatalf("获取用户名时出错: %v", err)
			}
			if userName == "" {
				fetchId = userId
			} else {
				fetchId = userName
			}
		} else {
			userIdNumer = 0
			fetchId = userId
			userName = userId
		}

		for i := 1; i <= 5; i++ {
			wg.Add(1)
			go func(fetchId string, collectionType int) {
				collections, err := fetchDataWithRetry(fetchId, collectionType, &wg, bar)
				if err != nil {
					// 如果重试两次都失败，则跳过
					return
				}

				err = saveToFile(userIdNumer, userName, collectionType, collections)
				if err != nil {
					log.Fatalf("保存数据时出错: %v", err)
				}
			}(fetchId, i)
		}
	}

	// 等待所有并发任务完成
	wg.Wait()
	log.Printf("用户%s-%s数据已保存！\n", userIds[0], userIds[len(userIds)-1])
}
