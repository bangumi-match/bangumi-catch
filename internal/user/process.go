package user

import (
	"log"
	"strconv"
)

func processUser(userID int) (JsonUserFile, error) {
	userStr := strconv.Itoa(userID)

	// 获取有效用户标识
	fetchID, userName, err := resolveUserID(userStr)
	if err != nil {
		return JsonUserFile{}, err
	}

	user := JsonUserFile{
		UserID:   userID,
		UserName: userName,
	}

	// 获取所有收藏类型的数据
	var collections [5][]Collection
	// processUser: 确保数据去重
	for ct := 1; ct <= 5; ct++ {
		data, err := fetchUserData(fetchID, userName, ct)
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
	existingUsers, err := readExistingData()
	if err != nil {
		return nil, err
	}
	var userIDs []int
	for _, user := range existingUsers {
		userIDs = append(userIDs, user.UserID)
	}
	return userIDs, nil
}
func getUsersWithEmptyData() ([]int, error) {
	existingUsers, err := readExistingData()
	if err != nil {
		return nil, err
	}
	var userIDs []int
	for _, user := range existingUsers {
		if len(user.Wish) == 0 && len(user.Collect) == 0 && len(user.Doing) == 0 && len(user.OnHold) == 0 && len(user.Dropped) == 0 {
			userIDs = append(userIDs, user.UserID)
		}
	}
	return userIDs, nil
}
