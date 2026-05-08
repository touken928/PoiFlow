package exporter

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	"github.com/touken928/PoiFlow/internal/task"
)

var CSVHeaders = map[string]string{
	"name": "名称", "lng": "经度", "lat": "纬度",
	"address": "地址", "telephone": "电话",
	"province": "省份", "city": "城市", "area": "区县",
	"uid": "UID", "query": "搜索词", "type": "分类",
	"taskName": "任务名", "target": "搜索目标",
}

var CSVGetters = map[string]func(task.Record) string{
	"name":      func(r task.Record) string { return r.Name },
	"lng":       func(r task.Record) string { return fmt.Sprintf("%.6f", r.Lng) },
	"lat":       func(r task.Record) string { return fmt.Sprintf("%.6f", r.Lat) },
	"address":   func(r task.Record) string { return r.Address },
	"telephone": func(r task.Record) string { return r.Telephone },
	"province":  func(r task.Record) string { return r.Province },
	"city":      func(r task.Record) string { return r.City },
	"area":      func(r task.Record) string { return r.Area },
	"uid":       func(r task.Record) string { return r.UID },
	"query":     func(r task.Record) string { return r.Query },
	"type":      func(r task.Record) string { return r.Type },
	"taskName":  func(r task.Record) string { return r.TaskName },
	"target":    func(r task.Record) string { return r.Target },
}

func ToCSV(records []task.Record, filePath string) error {
	// default: all fields
	fields := []string{"name", "lng", "lat", "address", "telephone", "province", "city", "area", "uid", "query", "type", "taskName", "target"}
	return ToCSVFiltered(records, filePath, fields)
}

func ToCSVFiltered(records []task.Record, filePath string, fields []string) error {
	f, err := os.Create(filePath)
	if err != nil { return fmt.Errorf("failed to create file: %w", err) }
	defer f.Close()
	if _, err := f.WriteString("\xEF\xBB\xBF"); err != nil { return fmt.Errorf("failed to write BOM: %w", err) }

	w := csv.NewWriter(f)
	defer w.Flush()

	// ensure lng/lat are always first
	header := []string{CSVHeaders["lng"], CSVHeaders["lat"]}
	cols := []string{"lng", "lat"}
	for _, f := range fields {
		if f == "lng" || f == "lat" { continue }
		header = append(header, CSVHeaders[f])
		cols = append(cols, f)
	}
	if err := w.Write(header); err != nil { return fmt.Errorf("failed to write header: %w", err) }

	for _, r := range records {
		row := make([]string, len(cols))
		for i, c := range cols {
			v := CSVGetters[c](r)
			row[i] = strings.ReplaceAll(v, "\"", "\"\"")
		}
		if err := w.Write(row); err != nil { return fmt.Errorf("failed to write record: %w", err) }
	}
	return nil
}
