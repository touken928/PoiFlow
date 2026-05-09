package task

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/touken928/PoiFlow/internal/akpool"
	"github.com/touken928/PoiFlow/pkg/baidu"
	"github.com/touken928/PoiFlow/pkg/division"
)

const maxTotal = 150
const defaultPageSize = 20

type EventHandler func(event string, data interface{})

type Executor struct {
	pool  *akpool.Pool
	Logf  func(taskID, level, msg string)
}

func NewExecutor(pool *akpool.Pool) *Executor {
	return &Executor{pool: pool, Logf: func(_, _, _ string) {}}
}

func (e *Executor) SearchTarget(query, poiType, region string) ([]baidu.POIResult, error) {
	regionLimit := true
	scope := baidu.ScopeBasic
	pageSize := defaultPageSize
	poolSize := len(e.pool.Items())
	maxRetries := poolSize * 2
	if maxRetries < 3 { maxRetries = 3 }
	var retries int

	for pageNum := 0; ; pageNum++ {
		ak := e.pool.Next()
		e.Logf("", "info", "使用 AK: "+ak[:minInt(8, len(ak))]+"...")
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
				e.Logf("", "error", "AK "+ak[:minInt(8, len(ak))]+"... 失效: "+apiErr.Error())
				retries++
				if retries >= maxRetries {
					return nil, fmt.Errorf("所有AK均已失效")
				}
				pageNum--
				continue
			}
			if resp == nil || len(resp.Results) == 0 { return nil, err }
		}
		e.pool.MarkSuccess(ak)
		if pageNum == 0 {
			allResults := make([]baidu.POIResult, len(resp.Results))
			copy(allResults, resp.Results)
			maxPage := maxTotal / pageSize
			if resp.Total != nil && *resp.Total < maxTotal { maxPage = (*resp.Total + pageSize - 1) / pageSize }
			if maxPage <= 1 || len(resp.Results) < pageSize { return allResults, nil }
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
						retries++
						if retries >= maxRetries {
							return nil, fmt.Errorf("所有AK均已失效")
						}
						pn--
						continue
					}
					break
				}
				e.pool.MarkSuccess(ak)
				allResults = append(allResults, subResp.Results...)
				if len(subResp.Results) < pageSize { break }
			}
			return allResults, nil
		}
	}
}

func minInt(a, b int) int { if a < b { return a }; return b }

type Queue struct {
	mu       sync.Mutex
	tasks    []*Task
	running  bool
	executor *Executor
	onEvent  EventHandler
	records  map[string][]Record
	logs     map[string][]LogEntry
	cacheDir string
	statePath string
}

func NewQueue(executor *Executor, onEvent EventHandler, cacheDir, statePath string) *Queue {
	os.MkdirAll(cacheDir, 0755)
	executor.Logf = func(_, level, msg string) {
		// executor doesn't have taskID, logs are emitted at execute level instead
	}
	q := &Queue{
		executor: executor, onEvent: onEvent,
		records: make(map[string][]Record),
		logs:    make(map[string][]LogEntry),
		cacheDir: cacheDir, statePath: statePath,
	}
	q.loadState()
	return q
}

func (q *Queue) Add(name, exportPath string, areaGran, queryGran Granularity, targets []Target, queries []SearchTerm) *Task {
	q.mu.Lock()
	defer q.mu.Unlock()

	searchTargets := expandTargets(targets, areaGran, queryGran)
	t := &Task{
		ID: uuid.New().String(), Name: name, Queries: queries,
		ExportPath: exportPath, AreaGranularity: areaGran,
		QueryGranularity: queryGran, Targets: targets,
		CompletedTargets: 0, Status: StatusPending, Progress: "0/" + itoa(len(searchTargets)),
		Records: len(searchTargets), CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	q.tasks = append(q.tasks, t)
	q.addLogUnsafe(t.ID, "info", "任务已创建，目标: "+itoa(len(searchTargets))+"个 | 搜索词: "+itoa(len(queries))+"个")
	q.emit("task:added", t)
	q.saveStateUnsafe()
	if !q.running {
		q.running = true
		go q.processLoop()
	}
	return t
}

func (q *Queue) cachePath(id string) string { return filepath.Join(q.cacheDir, id+".csv") }
func (q *Queue) logPath(id string) string   { return filepath.Join(q.cacheDir, id+".log") }

func (q *Queue) Pause(id string) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	for _, t := range q.tasks {
		if t.ID == id && t.Status == StatusRunning {
			t.Status = StatusPaused; t.UpdatedAt = time.Now()
			q.emit("task:updated", t); q.saveStateUnsafe(); return true
		}
	}
	return false
}

func (q *Queue) Resume(id string) bool {
	q.mu.Lock()
	for _, t := range q.tasks {
		if t.ID == id && t.Status == StatusPaused {
			t.Status = StatusPending; t.UpdatedAt = time.Now()
			q.mu.Unlock()
			q.emit("task:updated", t)
			if !q.running { q.mu.Lock(); q.running = true; q.mu.Unlock(); go q.processLoop() }
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
			t.Status = StatusCancelled; t.UpdatedAt = time.Now()
			q.emit("task:updated", t); q.saveStateUnsafe(); return true
		}
	}
	return false
}

func (q *Queue) Delete(id string) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	for i, t := range q.tasks {
		if t.ID == id {
			q.tasks = append(q.tasks[:i], q.tasks[i+1:]...)
			delete(q.records, id); delete(q.logs, id)
			os.Remove(q.cachePath(id))
			os.Remove(q.logPath(id))
			q.emit("task:deleted", id); q.saveStateUnsafe(); return true
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
	memoryLogs := q.logs[taskID]
	q.mu.Unlock()

	var out []LogEntry
	if data, err := os.ReadFile(q.logPath(taskID)); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			if line == "" { continue }
			parts := strings.SplitN(line, " ", 3)
			level := "info"
			if len(parts) >= 2 { level = strings.Trim(parts[1], "[]") }
			msg := line
			if len(parts) >= 3 { msg = parts[2] }
			out = append(out, LogEntry{Time: parts[0], Message: msg, Level: level})
		}
	}
	if len(memoryLogs) > len(out) {
		for _, e := range memoryLogs[len(out):] {
			out = append(out, e)
		}
	}
	return out
}

func (q *Queue) processLoop() {
	for {
		q.mu.Lock()
		var current *Task
		for _, t := range q.tasks {
			if t.Status == StatusPending { current = t; break }
		}
		if current == nil { q.running = false; q.mu.Unlock(); return }
		q.mu.Unlock()
		q.execute(current)
	}
}

func (q *Queue) execute(t *Task) {
	q.mu.Lock()
	t.Status = StatusRunning; t.UpdatedAt = time.Now()
	q.mu.Unlock()
	q.addLog(t.ID, "info", "任务开始执行")
	q.emit("task:updated", t)
	q.saveState()

	searchTargets := expandTargets(t.Targets, t.AreaGranularity, t.QueryGranularity)
	knownUIDs := make(map[string]bool)
	for _, p := range []string{t.ExportPath, q.cachePath(t.ID)} {
		if p == "" { continue }
		if loaded, _ := LoadExistingUIDs(p); len(loaded) > 0 {
			for k := range loaded { knownUIDs[k] = true }
		}
	}
	var allRecords []Record
	total := len(searchTargets)
	startFrom := t.CompletedTargets
	if startFrom > 0 {
		q.addLog(t.ID, "info", fmt.Sprintf("从第 %d/%d 个目标继续", startFrom+1, total))
	}

	targetRegion := func(t Target) string {
		if t.City != "" { return t.City + t.Name }
		return t.Name
	}

	for i, target := range searchTargets {
		if i < startFrom { continue }
		func() {
			q.mu.Lock()
			paused := t.Status == StatusPaused
			q.mu.Unlock()
			if paused { q.addLog(t.ID, "warn", "任务已暂停"); return }
		}()
		if t.Status == StatusPaused { break }

		for qi, term := range t.Queries {
			q.addLog(t.ID, "info", fmt.Sprintf("搜索 [%d/%d] %s | 词: %s", i+1, total, target.Name, term.Query))

			results, err := q.executor.SearchTarget(term.Query, term.Type, targetRegion(target))
			if err != nil {
				q.mu.Lock()
				t.Error = err.Error()
				if err.Error() == "所有AK均已失效" {
					t.Status = StatusPaused
					q.addLogUnsafe(t.ID, "error", "所有AK均已失效，任务已暂停")
				} else {
					t.Status = StatusFailed
				}
				t.UpdatedAt = time.Now()
				q.mu.Unlock()
				q.addLog(t.ID, "error", "搜索失败: "+err.Error())
				q.emit("task:failed", map[string]interface{}{"task": t, "target": target, "error": err.Error()})
				q.saveState()
				return
			}

			for _, r := range results {
				rec := Record{
					Name: r.Name, Lng: r.Location.Lng, Lat: r.Location.Lat,
					Address: r.Address, Telephone: r.Telephone,
					Province: r.Province, City: r.City, Area: r.Area,
					UID: r.UID, Query: term.Query, Type: term.Type, TaskName: t.Name, Target: target.Name,
				}
				allRecords = append(allRecords, rec)
				if t.ExportPath != "" { _ = AppendRecord(t.ExportPath, rec, knownUIDs) }
				_ = AppendRecord(q.cachePath(t.ID), rec, knownUIDs)
				knownUIDs[r.UID] = true
			}

			q.addLog(t.ID, "info", fmt.Sprintf("完成 [%d/%d] %s | 词: %s，获取 %d 条", i+1, total, target.Name, term.Query, len(results)))

			func() {
				q.mu.Lock()
				paused := t.Status == StatusPaused
				q.mu.Unlock()
				if qi < len(t.Queries)-1 && paused { q.addLog(t.ID, "warn", "任务已暂停") }
			}()
			if t.Status == StatusPaused { break }
		}
		if t.Status == StatusPaused { break }

		q.mu.Lock()
		t.Records = len(allRecords)
		t.CompletedTargets = i + 1
		t.Progress = itoa(i+1) + "/" + itoa(total)
		t.UpdatedAt = time.Now()
		q.mu.Unlock()
		q.saveState()
		q.emit("task:progress", map[string]interface{}{
			"task": t, "target": target, "records": len(allRecords),
			"total": total, "current": i + 1, "allCount": len(allRecords),
		})
	}

	q.mu.Lock()
	if t.Status == StatusPaused {
		q.addLogUnsafe(t.ID, "warn", "任务已暂停，下次将继续")
		q.mu.Unlock()
		q.emit("task:updated", t)
		q.saveState()
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
	q.saveState()
}

func (q *Queue) addLog(taskID, level, msg string) {
	q.mu.Lock()
	entry := LogEntry{Time: time.Now().Format("15:04:05"), Message: msg, Level: level}
	q.logs[taskID] = append(q.logs[taskID], entry)
	line := entry.Time + " [" + level + "] " + msg + "\n"
	q.mu.Unlock()
	q.appendToLogFile(taskID, line)
	q.emit("task:log", map[string]interface{}{"taskID": taskID, "entry": entry})
}

func (q *Queue) addLogUnsafe(taskID, level, msg string) {
	entry := LogEntry{Time: time.Now().Format("15:04:05"), Message: msg, Level: level}
	q.logs[taskID] = append(q.logs[taskID], entry)
	q.appendToLogFile(taskID, entry.Time+" ["+level+"] "+msg+"\n")
	q.emit("task:log", map[string]interface{}{"taskID": taskID, "entry": entry})
}

func (q *Queue) appendToLogFile(taskID, line string) {
	f, err := os.OpenFile(q.logPath(taskID), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil { return }
	defer f.Close()
	f.WriteString(line)
}

func (q *Queue) emit(event string, data interface{}) {
	if q.onEvent != nil { q.onEvent(event, data) }
}

func (q *Queue) Records(taskID string) []Record {
	p := filepath.Join(q.cacheDir, taskID+".csv")
	f, err := os.Open(p)
	if err != nil { return nil }
	defer f.Close()
	r := csv.NewReader(f)
	rows, err := r.ReadAll()
	if err != nil { return nil }
	var out []Record
	for i, row := range rows {
		if i == 0 { continue }
		if len(row) < 13 { continue }
		out = append(out, Record{
			Name: row[0], Lng: parseFloat(row[1]), Lat: parseFloat(row[2]),
			Address: row[3], Telephone: row[4],
			Province: row[5], City: row[6], Area: row[7],
			UID: row[8], Query: row[9], Type: row[10], TaskName: row[11], Target: row[12],
		})
	}
	return out
}

func parseFloat(s string) float64 {
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}

func (q *Queue) saveState() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.saveStateUnsafe()
}

func (q *Queue) saveStateUnsafe() {
	if q.statePath == "" { return }
	data, err := json.Marshal(q.tasks)
	if err != nil { return }
	os.WriteFile(q.statePath, data, 0644)
}

func (q *Queue) loadState() {
	if q.statePath == "" { return }
	data, err := os.ReadFile(q.statePath)
	if err != nil { return }
	var loaded []Task
	if err := json.Unmarshal(data, &loaded); err != nil { return }
	for i := range loaded {
		if loaded[i].Status == StatusRunning { loaded[i].Status = StatusPaused }
		q.logs[loaded[i].ID] = append(q.logs[loaded[i].ID], LogEntry{
			Time: time.Now().Format("15:04:05"), Message: "任务已从上次会话恢复（已暂停）", Level: "warn",
		})
	}
	q.mu.Lock()
	for i := range loaded {
		t := loaded[i]
		q.tasks = append(q.tasks, &t)
	}
	q.mu.Unlock()
}

func expandTargets(targets []Target, areaGran, queryGran Granularity) []Target {
	if areaGran == queryGran {
		out := make([]Target, len(targets)); copy(out, targets); return out
	}
	var expanded []Target
	for _, t := range targets {
		switch {
		case areaGran == GranularityProvince && queryGran == GranularityCity:
			for _, c := range division.Cities(t.Name) { expanded = append(expanded, Target{Province: t.Name, Name: c}) }
		case areaGran == GranularityProvince && queryGran == GranularityCounty:
			for _, c := range division.Cities(t.Name) {
				for _, ct := range division.Counties(t.Name, c) { expanded = append(expanded, Target{Province: t.Name, City: c, Name: ct}) }
			}
		case areaGran == GranularityCity && queryGran == GranularityCounty:
			for _, ct := range division.Counties(t.Province, t.Name) { expanded = append(expanded, Target{Province: t.Province, City: t.Name, Name: ct}) }
		}
	}
	return expanded
}

func ExpandCount(targets []Target, areaGran, queryGran Granularity) int {
	return len(expandTargets(targets, areaGran, queryGran))
}

func itoa(n int) string {
	if n == 0 { return "0" }
	s := ""
	for n > 0 { s = string(rune('0'+n%10)) + s; n /= 10 }
	return s
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

func recordRow(r Record) []string {
	return []string{
		r.Name, formatFloat(r.Lng), formatFloat(r.Lat),
		r.Address, r.Telephone, r.Province, r.City, r.Area,
		r.UID, r.Query, r.Type, r.TaskName, r.Target,
	}
}

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
		if len(row) >= 10 { uids[row[8]] = true }
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

func init() {
	// verify uuid works
	_ = uuid.New()
}
