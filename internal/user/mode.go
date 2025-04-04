package user

import (
	"fmt"
	"github.com/schollz/progressbar/v3"
	"log"
	"runtime"
	"sort"
	"sync"
	"time"
)

func createMode(userIDs []int) {
	var wg sync.WaitGroup
	sem := make(chan struct{}, runtime.NumCPU()*2)
	results := make(chan JsonUserFile, len(userIDs))

	bar := progressbar.NewOptions(len(userIDs),
		progressbar.OptionSetDescription("创建用户数据..."),
		progressbar.OptionShowCount(),
	)

	for _, uid := range userIDs {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			if user, err := processUser(userID); err == nil {
				results <- user
			} else {
				log.Println(err)
			}
			bar.Add(1)
		}(uid)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var users []JsonUserFile
	for u := range results {
		users = append(users, u)
	}

	// 排序并分配ProjectID
	sort.Slice(users, func(i, j int) bool {
		return users[i].UserID < users[j].UserID
	})
	for i := range users {
		users[i].ProjectID = i + 1
	}

	saveUserData(users)
	generateUserMap()
	fmt.Printf("创建成功！有效用户数: %d\n", len(users))
}

func updateMode(userIDs []int) {
	existingUsers, err := readExistingData()
	if err != nil {
		log.Fatal("读取现有数据失败:", err)
	}

	userMap := make(map[int]JsonUserFile)
	for _, user := range existingUsers {
		userMap[user.UserID] = user
	}

	// 过滤不存在的用户
	var validUserIDs []int
	notExistNumber := 0
	for _, userID := range userIDs {
		if _, exists := userMap[userID]; exists {
			validUserIDs = append(validUserIDs, userID)
		} else {
			notExistNumber++
		}
	}
	log.Printf("过滤不存在用户数: %d", notExistNumber)

	totalUsers := len(validUserIDs)
	totalChunks := (totalUsers + chunkSize - 1) / chunkSize

	log.Printf("开始批量更新，总用户数: %d，分 %d 批处理", totalUsers, totalChunks)

	for chunkIdx := 0; chunkIdx < totalUsers; chunkIdx += chunkSize {
		end := chunkIdx + chunkSize
		if end > totalUsers {
			end = totalUsers
		}
		currentChunk := validUserIDs[chunkIdx:end]

		// 批次信息
		batchNumber := (chunkIdx / chunkSize) + 1
		log.Printf("开始处理批次 %d/%d（用户 %d-%d）",
			batchNumber, totalChunks, chunkIdx+1, end)

		// 处理当前批次
		var (
			wg           sync.WaitGroup
			successCount int
			failureCount int
			mu           sync.Mutex
		)

		results := make(chan JsonUserFile, len(currentChunk))
		sem := make(chan struct{}, runtime.NumCPU()*2)
		bar := progressbar.NewOptions(len(currentChunk),
			progressbar.OptionSetDescription(fmt.Sprintf("批次 %d 进度", batchNumber)),
			progressbar.OptionShowCount(),
		)

		startTime := time.Now()

		// 处理当前批次的每个用户
		for _, uid := range currentChunk {
			wg.Add(1)
			go func(userID int) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				updatedUser, err := processUser(userID)
				if err != nil {
					mu.Lock()
					failureCount++
					mu.Unlock()
					log.Printf("用户 %d 更新失败: %v", userID, err)
					return
				}

				// 保持原有ProjectID
				updatedUser.ProjectID = userMap[userID].ProjectID
				results <- updatedUser
				bar.Add(1)
				mu.Lock()
				successCount++
				mu.Unlock()
			}(uid)
		}

		// 等待当前批次完成
		go func() {
			wg.Wait()
			close(results)
		}()

		// 收集当前批次结果
		var chunkUsers []JsonUserFile
		for u := range results {
			chunkUsers = append(chunkUsers, u)
		}

		// 更新用户映射
		for _, user := range chunkUsers {
			userMap[user.UserID] = user
		}

		// 转换为切片并排序
		finalUsers := make([]JsonUserFile, 0, len(userMap))
		for _, user := range userMap {
			finalUsers = append(finalUsers, user)
		}
		sort.Slice(finalUsers, func(i, j int) bool {
			return finalUsers[i].ProjectID < finalUsers[j].ProjectID
		})

		// 保存当前批次数据
		if err := saveUserData(finalUsers); err != nil {
			log.Printf("批次 %d 数据保存失败: %v", batchNumber, err)
		}

		// 记录批次统计
		duration := time.Since(startTime).Round(time.Second)
		log.Printf("批次 %d/%d 完成 | 成功: %d | 失败: %d | 耗时: %v | 当前总用户: %d",
			batchNumber, totalChunks, successCount, failureCount, duration, len(finalUsers))
	}

	// 最终保存
	finalUsers := make([]JsonUserFile, 0, len(userMap))
	for _, user := range userMap {
		finalUsers = append(finalUsers, user)
	}
	sort.Slice(finalUsers, func(i, j int) bool {
		return finalUsers[i].ProjectID < finalUsers[j].ProjectID
	})
	saveUserData(finalUsers)
	log.Printf("更新全部完成！总用户数: %d", len(userMap))
}

func addMode(userIDs []int) {
	existingUsers, err := readExistingData()
	if err != nil {
		log.Fatal("读取现有数据失败:", err)
	}

	// 获取当前最大ProjectID
	maxProjectID := 0
	for _, user := range existingUsers {
		if user.ProjectID > maxProjectID {
			maxProjectID = user.ProjectID
		}
	}

	existingUserIDs := make(map[int]bool)
	for _, user := range existingUsers {
		existingUserIDs[user.UserID] = true
	}

	var newUsers []int
	for _, uid := range userIDs {
		if !existingUserIDs[uid] {
			newUsers = append(newUsers, uid)
		} else {
			log.Printf("用户 %d 已存在，跳过添加", uid)
		}
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, runtime.NumCPU()*2)
	results := make(chan JsonUserFile, len(newUsers))

	bar := progressbar.NewOptions(len(newUsers),
		progressbar.OptionSetDescription("添加用户数据..."),
		progressbar.OptionShowCount(),
	)

	for _, uid := range newUsers {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			user, err := processUser(userID)
			if err != nil {
				log.Printf("用户 %d 添加失败: %v", userID, err)
				return
			}
			results <- user
			bar.Add(1)
		}(uid)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var usersToAdd []JsonUserFile
	for u := range results {
		usersToAdd = append(usersToAdd, u)
	}

	// 分配新的ProjectID
	currentMax := maxProjectID
	for i := range usersToAdd {
		currentMax++
		usersToAdd[i].ProjectID = currentMax
	}

	allUsers := append(existingUsers, usersToAdd...)
	saveUserData(allUsers)
	generateUserMap()
	fmt.Printf("添加成功！新增用户数: %d\n", len(usersToAdd))
}
