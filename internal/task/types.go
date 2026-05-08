package task

import "time"

type Granularity int

const (
	GranularityProvince Granularity = iota
	GranularityCity
	GranularityCounty
)

func (g Granularity) String() string {
	switch g {
	case GranularityProvince:
		return "省级"
	case GranularityCity:
		return "市级"
	case GranularityCounty:
		return "区县级"
	default:
		return "未知"
	}
}

type Status int

const (
	StatusPending Status = iota
	StatusRunning
	StatusPaused
	StatusCompleted
	StatusFailed
	StatusCancelled
)

func (s Status) String() string {
	switch s {
	case StatusPending:
		return "等待中"
	case StatusRunning:
		return "执行中"
	case StatusPaused:
		return "已暂停"
	case StatusCompleted:
		return "已完成"
	case StatusFailed:
		return "失败"
	case StatusCancelled:
		return "已取消"
	default:
		return "未知"
	}
}

type SearchTerm struct {
	Query string `json:"query"`
	Type  string `json:"type"`
}

type Target struct {
	Province string `json:"province"`
	City     string `json:"city"`
	Name     string `json:"name"`
}

type Task struct {
	ID               string       `json:"id"`
	Name             string       `json:"name"`
	Queries          []SearchTerm `json:"queries"`
	ExportPath       string       `json:"exportPath"`
	AreaGranularity  Granularity  `json:"areaGranularity"`
	QueryGranularity Granularity  `json:"queryGranularity"`
	Targets          []Target     `json:"targets"`
	Status           Status       `json:"status"`
	Progress         string       `json:"progress"`
	Records          int          `json:"records"`
	CompletedTargets int          `json:"completedTargets"`
	Error            string       `json:"error"`
	CreatedAt        time.Time    `json:"createdAt"`
	UpdatedAt        time.Time    `json:"updatedAt"`
}

type LogEntry struct {
	Time    string `json:"time"`
	Message string `json:"message"`
	Level   string `json:"level"`
}

type Record struct {
	Name      string  `json:"name"`
	Lng       float64 `json:"lng"`
	Lat       float64 `json:"lat"`
	Address   string  `json:"address"`
	Telephone string  `json:"telephone"`
	Province  string  `json:"province"`
	City      string  `json:"city"`
	Area      string  `json:"area"`
	UID       string  `json:"uid"`
	Query     string  `json:"query"`
	Type      string  `json:"type"`
	TaskName  string  `json:"taskName"`
	Target    string  `json:"target"`
}
