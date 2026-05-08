package task

import (
	"fmt"
	"sync"
	"time"

	"PoiFlow/internal/akpool"
	"PoiFlow/pkg/baidu"
	"PoiFlow/pkg/division"
)

const maxTotal = 150
const defaultPageSize = 20

type EventHandler func(event string, data interface{})

type Executor struct {
	pool *akpool.Pool
}

func NewExecutor(pool *akpool.Pool) *Executor {
	return &Executor{pool: pool}
}

func (e *Executor) SearchTarget(query, poiType, region string) ([]baidu.POIResult, error) {
	regionLimit := true
	scope := baidu.ScopeBasic
	pageSize := defaultPageSize

	for pageNum := 0; ; pageNum++ {
		ak := e.pool.Next()
		e.pool.Throttle(ak)
		client := baidu.NewClient(ak)
		resp, err := client.RegionSearch(&baidu.RegionRequest{
			Query: query, Region: region, Type: poiType,
			RegionLimit: &regionLimit, Scope: &scope,
			PageNum: &pageNum, PageSize: &pageSize,
		})

		if err != nil {
			if apiErr, ok := err.(*baidu.APIError); ok && akpool.NeedsRotate(apiErr.Status) {
				e.pool.MarkFailed(ak, apiErr.Error())
				pageNum--
				continue
			}
			if resp == nil || len(resp.Results) == 0 {
				return nil, err
			}
		}

		e.pool.MarkSuccess(ak)

		if pageNum == 0 {
			allResults := make([]baidu.POIResult, len(resp.Results))
			copy(allResults, resp.Results)
			maxPage := maxTotal / pageSize
			if resp.Total != nil && *resp.Total < maxTotal {
				maxPage = (*resp.Total + pageSize - 1) / pageSize
			}
			if maxPage <= 1 || len(resp.Results) < pageSize {
				return allResults, nil
			}
			for pn := 1; pn < maxPage; pn++ {
				ak := e.pool.Next()
				e.pool.Throttle(ak)
				client := baidu.NewClient(ak)
				subResp, subErr := client.RegionSearch(&baidu.RegionRequest{
					Query: query, Region: region, Type: poiType,
					RegionLimit: &regionLimit, Scope: &scope,
					PageNum: &pn, PageSize: &pageSize,
				})
				if subErr != nil {
					if apiErr, ok := subErr.(*baidu.APIError); ok && akpool.NeedsRotate(apiErr.Status) {
						e.pool.MarkFailed(ak, apiErr.Error())
						pn--
						continue
					}
					break
				}
				e.pool.MarkSuccess(ak)
				allResults = append(allResults, subResp.Results...)
				if len(subResp.Results) < pageSize {
					break
				}
			}
			return allResults, nil
		}
	}
}

type Queue struct {
	mu       sync.Mutex
	tasks    []*Task
	running  bool
	executor *Executor
	onEvent  EventHandler
	records  map[string][]Record
	logs     map[string][]LogEntry
}

func NewQueue(executor *Executor, onEvent EventHandler) *Queue {
	return &Queue{
		executor: executor, onEvent: onEvent,
		records: make(map[string][]Record),
		logs:    make(map[string][]LogEntry),
	}
}

func (q *Queue) Add(name, query, poiType, exportPath string, areaGran, queryGran Granularity, targets []Target) *Task {
	q.mu.Lock()
	defer q.mu.Unlock()

	searchTargets := expandTargets(targets, areaGran, queryGran)
	t := &Task{
		ID: newID(), Name: name, Query: query, Type: poiType,
		ExportPath: exportPath, AreaGranularity: areaGran,
		QueryGranularity: queryGran, Targets: targets,
		Status: StatusPending, Progress: "0/" + itoa(len(searchTargets)),
		Records: len(searchTargets), CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	q.tasks = append(q.tasks, t)
	q.addLogUnsafe(t.ID, "info", "任务已创建，目标数: "+itoa(len(searchTargets)))
	if exportPath != "" {
		q.addLogUnsafe(t.ID, "info", "导出路径: "+exportPath)
	}
	q.emit("task:added", t)
	if !q.running {
		q.running = true
		go q.processLoop()
	}
	return t
}

func (q *Queue) Pause(id string) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	for _, t := range q.tasks {
		if t.ID == id && t.Status == StatusRunning {
			t.Status = StatusPaused
			t.UpdatedAt = time.Now()
			q.emit("task:updated", t)
			return true
		}
	}
	return false
}

func (q *Queue) Resume(id string) bool {
	q.mu.Lock()
	for _, t := range q.tasks {
		if t.ID == id && t.Status == StatusPaused {
			t.Status = StatusPending
			t.UpdatedAt = time.Now()
			q.mu.Unlock()
			q.emit("task:updated", t)
			if !q.running {
				q.mu.Lock()
				q.running = true
				q.mu.Unlock()
				go q.processLoop()
			}
			return true
		}
	}
	q.mu.Unlock()
	return false
}

func (q *Queue) Cancel(id string) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	for _, t := range q.tasks {
		if t.ID == id && t.Status == StatusPending {
			t.Status = StatusCancelled
			t.UpdatedAt = time.Now()
			q.emit("task:updated", t)
			return true
		}
	}
	return false
}

func (q *Queue) Delete(id string) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	for i, t := range q.tasks {
		if t.ID == id {
			if t.Status == StatusRunning || t.Status == StatusPaused {
				return false
			}
			q.tasks = append(q.tasks[:i], q.tasks[i+1:]...)
			delete(q.records, id)
			delete(q.logs, id)
			q.emit("task:deleted", id)
			return true
		}
	}
	return false
}

func (q *Queue) List() []*Task {
	q.mu.Lock()
	defer q.mu.Unlock()
	out := make([]*Task, len(q.tasks))
	copy(out, q.tasks)
	return out
}

func (q *Queue) GetLogs(taskID string) []LogEntry {
	q.mu.Lock()
	defer q.mu.Unlock()
	l := q.logs[taskID]
	out := make([]LogEntry, len(l))
	copy(out, l)
	return out
}

func (q *Queue) processLoop() {
	for {
		q.mu.Lock()
		var current *Task
		for _, t := range q.tasks {
			if t.Status == StatusPending {
				current = t
				break
			}
		}
		q.mu.Unlock()
		if current == nil {
			q.mu.Lock()
			q.running = false
			q.mu.Unlock()
			return
		}
		q.execute(current)
	}
}

func (q *Queue) execute(t *Task) {
	q.mu.Lock()
	t.Status = StatusRunning
	t.UpdatedAt = time.Now()
	q.mu.Unlock()
	q.addLog(t.ID, "info", "任务开始执行")
	q.emit("task:updated", t)

	searchTargets := expandTargets(t.Targets, t.AreaGranularity, t.QueryGranularity)
	knownUIDs := make(map[string]bool)
	if t.ExportPath != "" {
		loaded, _ := LoadExistingUIDs(t.ExportPath)
		knownUIDs = loaded
		if len(knownUIDs) > 0 {
			q.addLog(t.ID, "info", fmt.Sprintf("从CSV加载 %d 个已知UID，将跳过重复", len(knownUIDs)))
		}
	}
	var allRecords []Record
	total := len(searchTargets)

	for i, target := range searchTargets {
		func() {
			q.mu.Lock()
			paused := t.Status == StatusPaused
			q.mu.Unlock()
			if paused {
				q.addLog(t.ID, "warn", "任务已暂停")
				return
			}
		}()
		if t.Status == StatusPaused {
			break
		}

		q.addLog(t.ID, "info", fmt.Sprintf("搜索 [%d/%d]: %s", i+1, total, target.Name))

		results, err := q.executor.SearchTarget(t.Query, t.Type, target.Name)
		if err != nil {
			q.mu.Lock()
			t.Error = err.Error()
			t.Status = StatusFailed
			t.UpdatedAt = time.Now()
			q.mu.Unlock()
			q.addLog(t.ID, "error", "搜索失败: "+err.Error())
			q.emit("task:failed", map[string]interface{}{"task": t, "target": target, "error": err.Error()})
			return
		}

		added := 0
		for _, r := range results {
			rec := Record{
				Name: r.Name, Lng: r.Location.Lng, Lat: r.Location.Lat,
				Address: r.Address, Telephone: r.Telephone,
				Province: r.Province, City: r.City, Area: r.Area,
				UID: r.UID, Query: t.Query, TaskName: t.Name, Target: target.Name,
			}
			allRecords = append(allRecords, rec)
			if t.ExportPath != "" {
				if !knownUIDs[r.UID] {
					_ = AppendRecord(t.ExportPath, rec, knownUIDs)
					added++
				}
			}
		}
		_ = added

		q.mu.Lock()
		t.Records = len(allRecords)
		t.Progress = itoa(i+1) + "/" + itoa(total)
		t.UpdatedAt = time.Now()
		q.mu.Unlock()

		hint := ""
		if len(results) > 0 {
			hint = fmt.Sprintf("，获取 %d 条", len(results))
		}
		q.addLog(t.ID, "info", fmt.Sprintf("完成 [%d/%d]: %s%s", i+1, total, target.Name, hint))
		q.emit("task:progress", map[string]interface{}{
			"task": t, "target": target, "records": len(results),
			"total": total, "current": i + 1, "allCount": len(allRecords),
		})
	}

	q.mu.Lock()
	if t.Status == StatusPaused {
		q.addLogUnsafe(t.ID, "warn", "任务已暂停，下次将继续")
		q.mu.Unlock()
		q.emit("task:updated", t)
		return
	}
	t.Status = StatusCompleted
	t.Records = len(allRecords)
	t.Progress = itoa(total) + "/" + itoa(total)
	t.UpdatedAt = time.Now()
	q.records[t.ID] = allRecords
	q.mu.Unlock()
	q.addLog(t.ID, "info", fmt.Sprintf("任务完成，共 %d 条记录", len(allRecords)))
	q.emit("task:completed", map[string]interface{}{"task": t, "records": allRecords})
}

func (q *Queue) addLog(taskID, level, msg string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.addLogUnsafe(taskID, level, msg)
}

func (q *Queue) addLogUnsafe(taskID, level, msg string) {
	q.logs[taskID] = append(q.logs[taskID], LogEntry{
		Time: time.Now().Format("15:04:05"), Message: msg, Level: level,
	})
}

func (q *Queue) emit(event string, data interface{}) {
	if q.onEvent != nil {
		q.onEvent(event, data)
	}
}

func (q *Queue) Records(taskID string) []Record {
	q.mu.Lock()
	defer q.mu.Unlock()
	r, ok := q.records[taskID]
	if !ok {
		return nil
	}
	out := make([]Record, len(r))
	copy(out, r)
	return out
}

func expandTargets(targets []Target, areaGran, queryGran Granularity) []Target {
	if areaGran == queryGran {
		out := make([]Target, len(targets))
		copy(out, targets)
		return out
	}
	var expanded []Target
	for _, t := range targets {
		switch {
		case areaGran == GranularityProvince && queryGran == GranularityCity:
			for _, c := range division.Cities(t.Name) {
				expanded = append(expanded, Target{Province: t.Name, Name: c})
			}
		case areaGran == GranularityProvince && queryGran == GranularityCounty:
			for _, c := range division.Cities(t.Name) {
				for _, ct := range division.Counties(t.Name, c) {
					expanded = append(expanded, Target{Province: t.Name, City: c, Name: ct})
				}
			}
		case areaGran == GranularityCity && queryGran == GranularityCounty:
			for _, ct := range division.Counties(t.Province, t.Name) {
				expanded = append(expanded, Target{Province: t.Province, City: t.Name, Name: ct})
			}
		}
	}
	return expanded
}

func ExpandCount(targets []Target, areaGran, queryGran Granularity) int {
	return len(expandTargets(targets, areaGran, queryGran))
}

var (
	idMu      sync.Mutex
	idCounter int64
)

func newID() string {
	idMu.Lock()
	idCounter++
	ts := time.Now().Format("150405.000") + itoa(int(idCounter%1000))
	idMu.Unlock()
	return ts
}

func itoa(n int) string {
	if n == 0 { return "0" }
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}
