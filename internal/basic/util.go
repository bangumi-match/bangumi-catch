package basic

import (
	"fmt"
	"strconv"
	"strings"
)

// 解析用户输入的ID列表（支持逗号和短横线）
func ParseIDList(input string) ([]int, error) {
	var ids []int
	parts := strings.Split(input, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.Contains(part, "-") {
			rangeParts := strings.Split(part, "-")
			if len(rangeParts) != 2 {
				return nil, fmt.Errorf("无效范围格式: %s", part)
			}
			start, err := strconv.Atoi(rangeParts[0])
			if err != nil {
				return nil, fmt.Errorf("无效起始ID: %s", rangeParts[0])
			}
			end, err := strconv.Atoi(rangeParts[1])
			if err != nil {
				return nil, fmt.Errorf("无效结束ID: %s", rangeParts[1])
			}
			for i := start; i <= end; i++ {
				ids = append(ids, i)
			}
		} else {
			id, err := strconv.Atoi(part)
			if err != nil {
				return nil, fmt.Errorf("无效ID: %s", part)
			}
			ids = append(ids, id)
		}
	}
	// 去重
	seen := make(map[int]bool)
	var unique []int
	for _, id := range ids {
		if !seen[id] {
			seen[id] = true
			unique = append(unique, id)
		}
	}
	return unique, nil
}
