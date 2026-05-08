package task

import "testing"

func TestGranularityString(t *testing.T) {
	tests := []struct {
		g    Granularity
		want string
	}{
		{GranularityProvince, "省级"},
		{GranularityCity, "市级"},
		{GranularityCounty, "区县级"},
	}
	for _, tt := range tests {
		if tt.g.String() != tt.want {
			t.Errorf("Granularity(%d).String() = '%s', want '%s'", tt.g, tt.g.String(), tt.want)
		}
	}
}

func TestStatusString(t *testing.T) {
	tests := []struct {
		s    Status
		want string
	}{
		{StatusPending, "等待中"},
		{StatusRunning, "执行中"},
		{StatusPaused, "已暂停"},
		{StatusCompleted, "已完成"},
		{StatusFailed, "失败"},
		{StatusCancelled, "已取消"},
	}
	for _, tt := range tests {
		if tt.s.String() != tt.want {
			t.Errorf("Status(%d).String() = '%s', want '%s'", tt.s, tt.s.String(), tt.want)
		}
	}
}

func TestTargetFields(t *testing.T) {
	t1 := Target{Province: "云南省", City: "", Name: "昆明市"}
	if t1.Name != "昆明市" {
		t.Errorf("expected Name '昆明市', got '%s'", t1.Name)
	}
	t2 := Target{Province: "北京市", City: "", Name: "北京市"}
	if t2.Name != "北京市" {
		t.Errorf("expected Name '北京市', got '%s'", t2.Name)
	}
	t3 := Target{Province: "上海市", City: "上海市", Name: "浦东新区"}
	if t3.Name != "浦东新区" {
		t.Errorf("expected Name '浦东新区', got '%s'", t3.Name)
	}
}

func TestNewID(t *testing.T) {
	id1 := newID()
	id2 := newID()
	if id1 == id2 {
		t.Error("expected different IDs")
	}
	if len(id1) < 6 {
		t.Errorf("expected ID length >= 6, got %d (%s)", len(id1), id1)
	}
}
