package exporter

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	"PoiFlow/internal/task"
)

func ToCSV(records []task.Record, filePath string) error {
	f, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()
	if _, err := f.WriteString("\xEF\xBB\xBF"); err != nil {
		return fmt.Errorf("failed to write BOM: %w", err)
	}
	w := csv.NewWriter(f)
	defer w.Flush()
	header := []string{"名称", "经度", "纬度", "地址", "电话", "省份", "城市", "区县", "UID", "搜索词", "任务名", "搜索目标"}
	if err := w.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}
	for _, r := range records {
		row := []string{
			r.Name, fmt.Sprintf("%.6f", r.Lng), fmt.Sprintf("%.6f", r.Lat),
			r.Address, r.Telephone, r.Province, r.City, r.Area,
			r.UID, r.Query, r.TaskName, r.Target,
		}
		for i, v := range row {
			row[i] = strings.ReplaceAll(v, "\"", "\"\"")
		}
		if err := w.Write(row); err != nil {
			return fmt.Errorf("failed to write record: %w", err)
		}
	}
	return nil
}
