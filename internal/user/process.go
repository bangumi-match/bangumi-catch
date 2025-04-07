package user

import (
	"encoding/json"
	"fmt"
	"github.com/schollz/progressbar/v3"
	"io/fs"
	"log"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

func processUser(userID int) (JsonUserFile, error) {

	// 获取有效用户标识
	fetchID, err := resolveUserID(userID)
	if err != nil {
		return JsonUserFile{}, err
	}

	user := JsonUserFile{
		UserID: userID,
	}
	if fetchID != strconv.Itoa(userID) {
		user.UserName = fetchID
	}

	// 获取所有收藏类型的数据
	var collections [5][]Collection
	// processUser: 确保数据去重
	for ct := 1; ct <= 5; ct++ {
		data, err := fetchUserData(fetchID, ct)
		if err != nil {
			log.Printf("用户 %s 类型 %d 数据获取失败: %v", fetchID, ct, err)
			continue
		}

		// **确保每个类型的数据是唯一的**
		uniqueData := make(map[int]Collection)
		for _, item := range data {
			uniqueData[item.SubjectID] = item
		}

		collections[ct-1] = make([]Collection, 0, len(uniqueData))
		for _, v := range uniqueData {
			collections[ct-1] = append(collections[ct-1], v)
		}
	}

	// 处理并过滤数据
	user.Wish = processCollections(collections[0])
	user.Collect = processCollections(collections[1])
	user.Doing = processCollections(collections[2])
	user.OnHold = processCollections(collections[3])
	user.Dropped = processCollections(collections[4])

	//// 有效性检查
	//totalEntries := len(user.Data.Wish) + len(user.Data.Collect) +
	//	len(user.Data.Doing) + len(user.Data.OnHold) + len(user.Data.Dropped)
	//
	//if totalEntries < 100 {
	//	return JsonUserFile{}, fmt.Errorf("用户 %d 有效条目不足", userID)
	//}

	return user, nil
}

func processCollections(collections []Collection) []Subject {
	var result []Subject
	existingSubjects := make(map[int]struct{})

	for _, c := range collections {
		if pid, exists := animeIDMap[c.SubjectID]; exists {
			if _, seen := existingSubjects[c.SubjectID]; !seen {
				result = append(result, Subject{
					SubjectID: c.SubjectID,
					ProjectID: pid,
					Tags:      c.Tags,
					Comment:   c.Comment,
					Rate:      c.Rate,
					UpdatedAt: c.UpdatedAt,
				})
				existingSubjects[c.SubjectID] = struct{}{}
			}
		}
	}
	return result
}

func getAllUserIDs() ([]int, error) {
	existingIDs, err := readExistingUserIDs()
	if err != nil {
		return nil, err
	}

	ids := make([]int, 0, len(existingIDs))
	for id := range existingIDs {
		ids = append(ids, id)
	}
	return ids, nil
}

func getUsersWithEmptyData() ([]int, error) {
	existingIDs, err := readExistingUserIDs()
	if err != nil {
		return nil, err
	}

	var emptyUsers []int
	chunkSize := 500 // 分批检查

	// 分批处理避免内存问题
	ids := make([]int, 0, len(existingIDs))
	for id := range existingIDs {
		ids = append(ids, id)
	}

	total := len(ids)
	for i := 0; i < total; i += chunkSize {
		end := i + chunkSize
		if end > total {
			end = total
		}

		var batchEmpty []int
		for _, uid := range ids[i:end] {
			user, err := readUserData(uid)
			if err != nil {
				continue
			}

			if isEmptyUserData(user) {
				batchEmpty = append(batchEmpty, uid)
			}
		}
		emptyUsers = append(emptyUsers, batchEmpty...)
	}

	return emptyUsers, nil
}

func isEmptyUserData(user JsonUserFile) bool {
	return len(user.Wish) == 0 &&
		len(user.Collect) == 0 &&
		len(user.Doing) == 0 &&
		len(user.OnHold) == 0 &&
		len(user.Dropped) == 0
}

// ------------------------- 合并功能 -------------------------
func mergeUserFiles(outputPath string) error {
	startTime := time.Now()
	log.Printf("开始合并用户数据...")

	entries, err := os.ReadDir(usersDir)
	if err != nil {
		return err
	}

	totalFiles := len(entries)
	bar := progressbar.NewOptions(totalFiles,
		progressbar.OptionSetDescription("合并进度"),
		progressbar.OptionShowCount(),
	)

	result := make([]JsonUserFile, 0, totalFiles)
	errCh := make(chan error, 1)
	doneCh := make(chan struct{})

	// 使用并行处理加速读取
	go func() {
		sem := make(chan struct{}, runtime.NumCPU()*2)
		var mu sync.Mutex

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			sem <- struct{}{}
			go func(e fs.DirEntry) {
				defer func() { <-sem }()

				if strings.HasSuffix(e.Name(), ".json") {
					idStr := e.Name()[:len(e.Name())-5]
					userID, err := strconv.Atoi(idStr)
					if err != nil {
						return
					}

					user, err := readUserData(userID)
					if err != nil {
						select {
						case errCh <- fmt.Errorf("读取用户 %d 失败: %v", userID, err):
						default:
						}
						return
					}

					mu.Lock()
					result = append(result, user)
					mu.Unlock()
					bar.Add(1)
				}
			}(entry)
		}

		// 等待所有goroutine完成
		for i := 0; i < cap(sem); i++ {
			sem <- struct{}{}
		}
		close(doneCh)
	}()

	select {
	case err := <-errCh:
		return err
	case <-doneCh:
	}

	// 排序
	sort.Slice(result, func(i, j int) bool {
		return result[i].ProjectID < result[j].ProjectID
	})

	// 写入文件
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return err
	}

	log.Printf("合并完成！总用户数: %d，耗时: %v",
		len(result),
		time.Since(startTime).Round(time.Second))
	return nil
}

func processCreateBatch(batchIDs []int, batchNumber int, totalChunks int, bar *progressbar.ProgressBar) {
	var (
		wg           sync.WaitGroup
		successCount int
		failureCount int
		mu           sync.Mutex
	)

	sem := make(chan struct{}, runtime.NumCPU()*2)
	startTime := time.Now()

	for _, uid := range batchIDs {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			user, err := processUser(userID)
			if err != nil {
				mu.Lock()
				failureCount++
				mu.Unlock()
				log.Printf("用户 %d 创建失败: %v", userID, err)
				return
			}

			if err := saveUserData(user); err != nil {
				mu.Lock()
				failureCount++
				mu.Unlock()
				log.Printf("用户 %d 保存失败: %v", userID, err)
				return
			}

			mu.Lock()
			successCount++
			mu.Unlock()
			bar.Add(1)
		}(uid)
	}

	wg.Wait()
	duration := time.Since(startTime).Round(time.Second)
	log.Printf("批次 %d/%d 完成 | 成功: %d | 失败: %d | 耗时: %v",
		batchNumber, totalChunks, successCount, failureCount, duration)
}
