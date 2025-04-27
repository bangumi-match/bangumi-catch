package user

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/schollz/progressbar/v3"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

func createMode(userIDs []int) {
	chunkSize := 100
	totalChunks := (len(userIDs) + chunkSize - 1) / chunkSize

	bar := progressbar.NewOptions(len(userIDs),
		progressbar.OptionSetDescription("总体进度"),
		progressbar.OptionShowCount(),
	)

	for chunkIdx := 0; chunkIdx < len(userIDs); chunkIdx += chunkSize {
		end := chunkIdx + chunkSize
		if end > len(userIDs) {
			end = len(userIDs)
		}
		batchIDs := userIDs[chunkIdx:end]

		processCreateBatch(batchIDs, (chunkIdx/chunkSize)+1, totalChunks, bar)
	}

	log.Printf("正在整理数据，分配project_id！")
	// 处理完成后不再重新生成映射，需要手动调用
	//generateUserMap()
	log.Printf("创建成功！总处理用户数: %d\n", len(userIDs))
}

func updateMode(userIDs []int) {
	// 读取现有用户ID集合
	existingIDs, err := readExistingUserIDs()
	if err != nil {
		log.Fatal("读取现有用户ID失败:", err)
	}

	// 过滤有效用户ID
	var validUserIDs []int
	for _, uid := range userIDs {
		if _, exists := existingIDs[uid]; exists {
			validUserIDs = append(validUserIDs, uid)
		}
	}

	totalUsers := len(validUserIDs)
	chunkSize := 200 // 根据内存情况调整批次大小
	totalChunks := (totalUsers + chunkSize - 1) / chunkSize

	for chunkIdx := 0; chunkIdx < totalUsers; chunkIdx += chunkSize {
		end := chunkIdx + chunkSize
		if end > totalUsers {
			end = totalUsers
		}
		currentChunk := validUserIDs[chunkIdx:end]

		processUpdateBatch(currentChunk, (chunkIdx/chunkSize)+1, totalChunks)
	}

	log.Printf("正在整理数据，分配project_id！")
	//generateUserMap() // 处理完成后不再重新生成映射，需要手动调用
	log.Printf("更新全部完成！总用户数: %d", len(existingIDs))
}

func processUpdateBatch(batchIDs []int, batchNumber int, totalChunks int) {
	var (
		wg           sync.WaitGroup
		successCount int
		failureCount int
		mu           sync.Mutex
	)

	results := make(chan JsonUserFile, len(batchIDs))
	sem := make(chan struct{}, runtime.NumCPU()*2)

	bar := progressbar.NewOptions(len(batchIDs),
		progressbar.OptionSetDescription(fmt.Sprintf("批次 %d 进度", batchNumber)),
		progressbar.OptionShowCount(),
	)

	startTime := time.Now()

	// 处理当前批次的每个用户
	for _, uid := range batchIDs {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			// 读取现有用户数据
			existingUser, err := readUserData(userID)
			if err != nil {
				mu.Lock()
				failureCount++
				mu.Unlock()
				log.Printf("读取用户 %d 数据失败: %v", userID, err)
				return
			}

			// 处理更新
			updatedUser, err := processUser(userID)
			if err != nil {
				mu.Lock()
				failureCount++
				mu.Unlock()
				log.Printf("用户 %d 更新失败: %v", userID, err)
				return
			}

			// 保留原有ProjectID
			updatedUser.ProjectID = existingUser.ProjectID
			results <- updatedUser
			bar.Add(1)
			mu.Lock()
			successCount++
			mu.Unlock()
		}(uid)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	// 保存本批次结果
	for u := range results {
		if err := saveUserData(u); err != nil {
			log.Printf("用户 %d 保存失败: %v", u.UserID, err)
		}
	}

	// 记录统计信息
	duration := time.Since(startTime).Round(time.Second)
	log.Printf("批次 %d/%d 完成 | 成功: %d | 失败: %d | 耗时: %v",
		batchNumber, totalChunks, successCount, failureCount, duration)
}

func splitUserFile(inputPath string) error {
	startTime := time.Now()
	log.Printf("开始拆分用户数据文件...")

	// 读取原始大文件
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("文件读取失败: %v", err)
	}

	// 解析JSON数据
	var users []JsonUserFile
	if err := json.Unmarshal(data, &users); err != nil {
		return fmt.Errorf("JSON解析失败: %v", err)
	}

	totalUsers := len(users)
	bar := progressbar.NewOptions(totalUsers,
		progressbar.OptionSetDescription("拆分进度"),
		progressbar.OptionShowCount(),
	)

	successCount := 0
	failureCount := 0
	sem := make(chan struct{}, runtime.NumCPU()*2) // 并发控制
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, user := range users {
		wg.Add(1)
		go func(u JsonUserFile) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			// 检查数据有效性
			if u.UserID == 0 {
				mu.Lock()
				failureCount++
				mu.Unlock()
				log.Printf("无效用户数据: %+v", u)
				return
			}

			// 保存为独立文件
			if err := saveUserData(u); err != nil {
				mu.Lock()
				failureCount++
				mu.Unlock()
				log.Printf("用户 %d 保存失败: %v", u.UserID, err)
				return
			}

			mu.Lock()
			successCount++
			mu.Unlock()
			bar.Add(1)
		}(user)
	}

	wg.Wait()

	// 处理完成后不再重新生成映射，需要手动调用
	//generateUserMap()

	log.Printf("拆分完成！总用户: %d | 成功: %d | 失败: %d | 耗时: %v",
		totalUsers,
		successCount,
		failureCount,
		time.Since(startTime).Round(time.Second))
	return nil
}

func generateUserMap() {
	log.Println("开始重新生成用户映射表...")
	startTime := time.Now()
	var deletedUsers []JsonUserFile // 记录被删除的空用户

	// 读取所有用户数据
	entries, err := os.ReadDir(usersDir)
	if err != nil {
		log.Fatal("读取用户目录失败:", err)
	}

	var users []JsonUserFile
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		// 解析用户ID
		idStr := entry.Name()[:len(entry.Name())-5]
		userID, err := strconv.Atoi(idStr)
		if err != nil {
			log.Printf("跳过无效文件名: %s", entry.Name())
			continue
		}

		// 读取用户数据
		user, err := readUserData(userID)
		if err != nil {
			log.Printf("读取用户 %d 数据失败: %v", userID, err)
			continue
		}

		// 更新动画条目的 project_id
		updateAnimeProjectIDs(&user)

		// 检查所有收藏是否为空
		if isEmptyUserData(user) {
			// 删除用户文件
			filePath := filepath.Join(usersDir, entry.Name())
			if err := os.Remove(filePath); err != nil {
				log.Printf("删除用户 %d 文件失败: %v", userID, err)
			} else {
				deletedUsers = append(deletedUsers, user)
				log.Printf("已删除空用户: ID=%-8d | 用户名=%s", user.UserID, user.UserName)
			}
			continue // 跳过有效用户列表
		}

		users = append(users, user)
	}

	// 输出清除结果
	if len(deletedUsers) > 0 {
		log.Println("\n=== 已删除空用户统计 ===")
		for _, u := range deletedUsers {
			log.Printf("ID: %-8d | 用户名: %s", u.UserID, u.UserName)
		}
		log.Printf("共删除 %d 个空用户\n", len(deletedUsers))
	} else {
		log.Println("未发现空用户")
	}

	// 按 UserID 排序
	sort.Slice(users, func(i, j int) bool {
		return users[i].UserID < users[j].UserID
	})

	// 重新分配用户 project_id
	for i := range users {
		users[i].ProjectID = i // 从0开始连续分配
		if err := saveUserData(users[i]); err != nil {
			log.Printf("保存用户 %d 数据失败: %v", users[i].UserID, err)
		}
	}

	// 生成CSV文件
	file, err := os.Create(userMapFile)
	if err != nil {
		log.Fatal("创建映射文件失败:", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 写入标题行
	if err := writer.Write([]string{"project_id", "user_id", "user_name"}); err != nil {
		log.Fatal("写入CSV标题失败:", err)
	}

	// 写入数据行
	for _, u := range users {
		record := []string{
			strconv.Itoa(u.ProjectID),
			strconv.Itoa(u.UserID),
			u.UserName,
		}
		if err := writer.Write(record); err != nil {
			log.Printf("跳过用户 %d: CSV写入失败: %v", u.UserID, err)
		}
	}

	log.Printf("映射表生成完成！有效用户数: %d | 耗时: %v",
		len(users),
		time.Since(startTime).Round(time.Second))
}

// 更新用户收藏动画的 project_id
func updateAnimeProjectIDs(user *JsonUserFile) {
	update := func(list *[]Subject) {
		newList := make([]Subject, 0, len(*list))
		for _, subject := range *list {
			if projectID, exists := animeIDMap[subject.SubjectID]; exists {
				subject.ProjectID = projectID
				newList = append(newList, subject)
			} else {
				log.Printf("用户 %d: 动画ID %d 无映射关系，已过滤", user.UserID, subject.SubjectID)
			}
		}
		*list = newList
	}

	update(&user.Wish)
	update(&user.Collect)
	update(&user.Doing)
	update(&user.OnHold)
	update(&user.Dropped)
}
