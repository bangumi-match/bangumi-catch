package user

import (
	"encoding/json"
	"fmt"
	"github.com/gocolly/colly/v2"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

func getUserName(userID string) string {
	resp, err := http.Get(fmt.Sprintf("https://bgm.tv/user/%s", userID))
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.Request.URL.Path != fmt.Sprintf("/user/%s", userID) {
		parts := strings.Split(resp.Request.URL.Path, "/")
		if len(parts) > 2 {
			return parts[2]
		}
	}
	return ""
}

func fetchUserData(userID string, userName string, collectionType int) ([]Collection, error) {
	fetchID := ""
	if userName != "" {
		fetchID = userName
	} else {
		fetchID = userID
	}

	var result []Collection
	offset := 0
	limit := 40

	c := colly.NewCollector()
	c.SetRequestTimeout(60 * time.Second)
	c.Async = true

	url := fmt.Sprintf("https://api.bgm.tv/v0/users/%s/collections?subject_type=2&type=%d", fetchID, collectionType)

	var mu sync.Mutex
	var lastErr error
	requestedOffsets := make(map[int]bool) // **记录已请求的 offset**

	for {
		if requestedOffsets[offset] {
			log.Printf("跳过重复请求 offset=%d", offset)
			break // **如果该 offset 已请求，则跳过**
		}
		requestedOffsets[offset] = true // **标记为已请求**

		currentUrl := fmt.Sprintf("%s&limit=%d&offset=%d", url, limit, offset)
		var response ApiResponse

		c.OnResponse(func(r *colly.Response) {
			mu.Lock()
			defer mu.Unlock()

			if err := json.Unmarshal(r.Body, &response); err != nil {
				lastErr = fmt.Errorf("解析失败 %s: %v", currentUrl, err)
				return
			}

			result = append(result, response.Data...)
		})

		c.OnError(func(r *colly.Response, err error) {
			mu.Lock()
			lastErr = fmt.Errorf("请求失败 %s: %v", currentUrl, err)
			mu.Unlock()
		})

		if err := c.Visit(currentUrl); err != nil {
			return nil, err
		}

		c.Wait() // **等待当前请求完成**

		// **检查是否达到最后一页**
		if lastErr != nil {
			return nil, lastErr
		}
		if len(response.Data) < limit {
			break // **如果返回的数据不足 limit，说明到最后一页，停止请求**
		}

		offset += limit // **正确更新 offset**
	}
	return result, nil
}
