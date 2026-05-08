package exporter

import (
	"os"
	"strings"
	"testing"

	"github.com/touken928/PoiFlow/internal/task"
)

func TestToCSV(t *testing.T) {
	records := []task.Record{
		{
			Name:     "测试POI",
			Lng:      116.315845,
			Lat:      40.043840,
			Address:  "北京市海淀区",
			Telephone: "010-12345678",
			Province: "北京市",
			City:     "北京市",
			Area:     "海淀区",
			UID:      "test_uid_001",
			Query:    "ATM机",
			TaskName: "测试任务",
			Target:   "北京市",
		},
		{
			Name:     "另一个POI",
			Lng:      121.490486,
			Lat:      31.235191,
			Address:  "上海市浦东新区",
			Telephone: "021-87654321",
			Province: "上海市",
			City:     "上海市",
			Area:     "浦东新区",
			UID:      "test_uid_002",
			Query:    "银行",
			TaskName: "测试任务",
			Target:   "上海市",
		},
	}

	tmpFile := "test_export.csv"
	defer os.Remove(tmpFile)

	if err := ToCSV(records, tmpFile); err != nil {
		t.Fatalf("ToCSV failed: %v", err)
	}

	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	content := string(data)
	if !strings.HasPrefix(content, "\xEF\xBB\xBF") {
		t.Error("expected BOM at start of file")
	}

	if !strings.Contains(content, "测试POI") {
		t.Error("expected '测试POI' in CSV output")
	}
	if !strings.Contains(content, "116.315845") {
		t.Error("expected '116.315845' in CSV output")
	}
	if !strings.Contains(content, "ATM机") {
		t.Error("expected 'ATM机' in CSV output")
	}
}
