package subject

import (
	"fmt"
	"log"
	"os"
	"time"
)

// ------------------------- 工具函数 -------------------------
func initLog() (*os.File, error) {
	logDir := "logs"
	if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("创建日志目录失败: %v", err)
	}
	logFile, err := os.Create(fmt.Sprintf("%s/log_subject_%s.txt", logDir, time.Now().Format("20060102_150405")))
	if err != nil {
		return nil, fmt.Errorf("创建日志文件失败: %v", err)
	}
	log.SetOutput(logFile)
	return logFile, nil
}

// ------------------------- 新增日期处理相关函数 -------------------------

// 解析年月范围输入
func parseDateRange(startInput, endInput string) ([]struct{ Year, Month int }, error) {
	startTime, err := time.Parse("2006-01", startInput)
	if err != nil {
		return nil, fmt.Errorf("无效的起始日期格式")
	}

	endTime, err := time.Parse("2006-01", endInput)
	if err != nil {
		return nil, fmt.Errorf("无效的结束日期格式")
	}

	if startTime.After(endTime) {
		return nil, fmt.Errorf("起始日期不能晚于结束日期")
	}

	var dates []struct{ Year, Month int }
	for current := startTime; !current.After(endTime); current = current.AddDate(0, 1, 0) {
		dates = append(dates, struct{ Year, Month int }{
			Year:  current.Year(),
			Month: int(current.Month()),
		})
	}

	return dates, nil
}
