package task

import (
	"encoding/csv"
	"os"
)

func LoadExistingUIDs(path string) (map[string]bool, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) { return make(map[string]bool), nil }
		return nil, err
	}
	defer f.Close()
	r := csv.NewReader(f)
	rows, err := r.ReadAll()
	if err != nil { return nil, err }
	uids := make(map[string]bool)
	for i, row := range rows {
		if i == 0 { continue }
		if len(row) >= 10 { uids[row[9]] = true }
	}
	return uids, nil
}

func AppendRecord(path string, rec Record, knownUIDs map[string]bool) error {
	if knownUIDs[rec.UID] { return nil }
	flag := os.O_APPEND | os.O_CREATE | os.O_WRONLY
	f, err := os.OpenFile(path, flag, 0644)
	if err != nil { return err }
	defer f.Close()
	info, _ := f.Stat()
	if info.Size() == 0 {
		if _, err := f.WriteString("\xEF\xBB\xBF"); err != nil { return err }
		w := csv.NewWriter(f)
		_ = w.Write([]string{"名称", "经度", "纬度", "地址", "电话", "省份", "城市", "区县", "UID", "搜索词", "分类", "任务名", "搜索目标"})
		w.Flush()
	}
	w := csv.NewWriter(f)
	defer w.Flush()
	return w.Write(recordRow(rec))
}

func recordRow(r Record) []string {
	return []string{
		r.Name, formatFloat(r.Lng), formatFloat(r.Lat),
		r.Address, r.Telephone, r.Province, r.City, r.Area,
		r.UID, r.Query, r.Type, r.TaskName, r.Target,
	}
}

func formatFloat(v float64) string {
	s := itoa(int(v))
	frac := int((v - float64(int(v))) * 1000000)
	if frac < 0 { frac = -frac }
	return s + "." + padZeros(frac, 6)
}

func padZeros(n, w int) string {
	s := itoa(n)
	for len(s) < w { s = "0" + s }
	return s
}
