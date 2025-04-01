package subject

import (
	"encoding/json"
	"fmt"
	"github.com/gocolly/colly/v2"
	"github.com/schollz/progressbar/v3"
	"log"
	"runtime"
	"sync"
	"time"
)

// 按日期范围抓取数据
func fetchByDateRange(dates []struct{ Year, Month int }, token string) []JsonSubjectFile {
	var (
		wg        sync.WaitGroup
		results   = make(chan JsonSubjectFile, 1000)
		dateSem   = make(chan struct{}, 3) // 限制并发日期请求数
		pageSem   = make(chan struct{}, 5) // 限制分页请求并发数
		bar       = progressbar.Default(int64(len(dates)))
		collected []JsonSubjectFile
	)

	for _, date := range dates {
		wg.Add(1)
		go func(year, month int) {
			defer wg.Done()
			dateSem <- struct{}{}
			defer func() { <-dateSem }()

			limit := 40
			offset := 0
			for {
				pageSem <- struct{}{}

				var (
					pageSubjects []JsonSubjectFile
					_            error
					is400        bool
				)

				url := fmt.Sprintf("https://api.bgm.tv/v0/subjects?type=2&sort=date&year=%d&month=%d&limit=%d&offset=%d",
					year, month, limit, offset)

				c := colly.NewCollector()
				c.SetRequestTimeout(60 * time.Second)

				if token != "" {
					c.OnRequest(func(r *colly.Request) {
						r.Headers.Set("Authorization", "Bearer "+token)
					})
				}

				// 响应处理
				c.OnResponse(func(r *colly.Response) {
					var responseData struct {
						Data []JsonSubjectFile `json:"data"`
					}
					if err := json.Unmarshal(r.Body, &responseData); err != nil {
						log.Printf("日期 %d-%02d offset %d 解析失败: %v", year, month, offset, err)
						return
					}
					pageSubjects = responseData.Data

					// 过滤并发送有效条目
					for _, subj := range responseData.Data {
						if subj.Type == 2 && subj.Rating.Rank != 0 {
							results <- subj
						}
					}
				})

				// 错误处理
				c.OnError(func(r *colly.Response, err error) {
					_ = err
					if r != nil && r.StatusCode == 400 {
						is400 = true
					}
				})

				// 执行请求并等待完成
				err := c.Visit(url)
				c.Wait()
				<-pageSem

				// 错误处理逻辑
				if is400 {
					break
				}
				if err != nil {
					log.Printf("请求失败 %d-%02d offset %d: %v", year, month, offset, err)
					break
				}

				// 数据不足说明最后一页
				if len(pageSubjects) < limit {
					break
				}

				offset += limit
				time.Sleep(500 * time.Millisecond)
			}
			bar.Add(1)
		}(date.Year, date.Month)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for subj := range results {
		collected = append(collected, subj)
	}
	return collected
}

// ------------------------- 数据抓取逻辑 -------------------------

func fetchByIdList(ids []int, token string) []JsonSubjectFile {

	numCPU := runtime.NumCPU()
	log.Printf("使用 %d 线程", numCPU)

	var (
		wg       sync.WaitGroup
		logMutex sync.Mutex
		sem      = make(chan struct{}, numCPU)
		results  = make(chan JsonSubjectFile, len(ids))
		bar      = progressbar.Default(int64(len(ids)))
	)

	for _, id := range ids {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			url := fmt.Sprintf("https://api.bgm.tv/v0/subjects/%d", id)
			c := colly.NewCollector()

			if token != "" {
				c.OnRequest(func(r *colly.Request) {
					r.Headers.Set("Authorization", "Bearer "+token)
				})
			}

			c.OnResponse(func(r *colly.Response) {
				var subject JsonSubjectFile
				if err := json.Unmarshal(r.Body, &subject); err != nil {
					logMutex.Lock()
					log.Printf("解析JSON失败（ID %d）: %v", id, err)
					logMutex.Unlock()
					return
				}

				if subject.OriginalID != id {
					logMutex.Lock()
					log.Printf("ID %d 不存在", id)
					logMutex.Unlock()
					return
				}

				if subject.Type != 2 || subject.Rating.Rank == 0 {
					logMutex.Lock()
					log.Printf("ID %d 不符合条件（类型：%d，排名：%d）", id, subject.Type, subject.Rating.Rank)
					logMutex.Unlock()
					return
				}

				results <- subject
				bar.Add(1)
			})

			c.OnError(func(r *colly.Response, err error) {
				logMutex.Lock()
				log.Printf("请求错误（ID %d）: %v", id, err)
				logMutex.Unlock()
			})

			c.Visit(url)
		}(id)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var subjects []JsonSubjectFile
	for subj := range results {
		subjects = append(subjects, subj)
	}
	return subjects
}
