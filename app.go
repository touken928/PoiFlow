package main

import (
	"context"
	"fmt"

	"github.com/touken928/PoiFlow/internal/akpool"
	"github.com/touken928/PoiFlow/internal/akstore"
	"github.com/touken928/PoiFlow/internal/exporter"
	"github.com/touken928/PoiFlow/internal/task"
	"github.com/touken928/PoiFlow/pkg/baidu"
	"github.com/touken928/PoiFlow/pkg/division"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const akFileName = "config.yaml"

type App struct {
	ctx    context.Context
	akPool *akpool.Pool
	taskQ  *task.Queue
}

func NewApp() *App {
	a := &App{}
	a.reloadAKs()
	executor := task.NewExecutor(a.akPool)
	a.taskQ = task.NewQueue(executor, a.onTaskEvent)
	return a
}

func (a *App) startup(ctx context.Context) { a.ctx = ctx }

func (a *App) reloadAKs() {
	entries, err := akstore.Load(akFileName)
	if err != nil { println("load aks failed:", err.Error()) }
	keys := make([]string, len(entries))
	for i, e := range entries { keys[i] = e.Key }
	if a.akPool == nil {
		a.akPool = akpool.New(keys, nil)
	} else {
		a.akPool.Rebuild(keys, nil)
	}
}

func (a *App) onTaskEvent(event string, data interface{}) {
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, event, data)
	}
}

func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

func (a *App) GetProvinces() []string                     { return division.Provinces() }
func (a *App) GetCities(province string) []string          { return division.Cities(province) }
func (a *App) GetCounties(province, city string) []string  { return division.Counties(province, city) }

type TaskTargetInput struct {
	Province string `json:"province"`
	City     string `json:"city"`
	Name     string `json:"name"`
}

type SearchTermInput struct {
	Query string `json:"query"`
	Type  string `json:"type"`
}

func (a *App) CreateTask(name, exportPath string, areaGran, queryGran int, targets []TaskTargetInput, queries []SearchTermInput) *task.Task {
	ts := make([]task.Target, len(targets))
	for i, t := range targets {
		ts[i] = task.Target{Province: t.Province, City: t.City, Name: t.Name}
	}
	qs := make([]task.SearchTerm, len(queries))
	for i, q := range queries {
		qs[i] = task.SearchTerm{Query: q.Query, Type: q.Type}
	}
	return a.taskQ.Add(name, exportPath, task.Granularity(areaGran), task.Granularity(queryGran), ts, qs)
}

func (a *App) GetTasks() []*task.Task     { return a.taskQ.List() }
func (a *App) CancelTask(id string) bool  { return a.taskQ.Cancel(id) }
func (a *App) DeleteTask(id string) bool  { return a.taskQ.Delete(id) }
func (a *App) PauseTask(id string) bool   { return a.taskQ.Pause(id) }
func (a *App) ResumeTask(id string) bool  { return a.taskQ.Resume(id) }
func (a *App) GetTaskLogs(id string) []task.LogEntry { return a.taskQ.GetLogs(id) }

func (a *App) ExpandCount(areaGran, queryGran int, targets []TaskTargetInput) int {
	ts := make([]task.Target, len(targets))
	for i, t := range targets {
		ts[i] = task.Target{Province: t.Province, City: t.City, Name: t.Name}
	}
	return task.ExpandCount(ts, task.Granularity(areaGran), task.Granularity(queryGran))
}

type AKInfo struct {
	Name    string `json:"name"`
	AK      string `json:"ak"`
	Used    int    `json:"used"`
	Failed  bool   `json:"failed"`
	FailMsg string `json:"failMsg"`
}

func (a *App) GetAKItems() []AKInfo {
	items := a.akPool.Items()
	entries, _ := akstore.Load(akFileName)
	nameMap := make(map[string]string, len(entries))
	for _, e := range entries { nameMap[e.Key] = e.Name }

	out := make([]AKInfo, len(items))
	for i, it := range items {
		out[i] = AKInfo{Name: nameMap[it.AK], AK: it.AK, Used: it.Used, Failed: it.Failed, FailMsg: it.FailMsg}
	}
	return out
}

func (a *App) ResetAKPool()   { a.akPool.ResetAll() }

func (a *App) AddAK(name, key string) string {
	if key == "" { return "AK不能为空" }
	entries, err := akstore.Load(akFileName)
	if err != nil { return "读取配置失败: " + err.Error() }
	for _, e := range entries {
		if e.Key == key { return "AK已存在" }
	}
	entries = append(entries, akstore.Entry{Name: name, Key: key})
	if err := akstore.Save(akFileName, entries); err != nil { return "保存失败: " + err.Error() }
	a.reloadAKs()
	return ""
}

func (a *App) RemoveAK(key string) string {
	entries, err := akstore.Load(akFileName)
	if err != nil { return "读取配置失败: " + err.Error() }
	filtered := make([]akstore.Entry, 0, len(entries))
	for _, e := range entries {
		if e.Key != key { filtered = append(filtered, e) }
	}
	if len(filtered) == len(entries) { return "AK不存在" }
	if err := akstore.Save(akFileName, filtered); err != nil { return "保存失败: " + err.Error() }
	a.reloadAKs()
	return ""
}

func (a *App) VerifyAK(ak string) string {
	client := baidu.NewClient(ak)
	err := client.Ping()
	if err != nil {
		return "失效: " + err.Error()
	}
	return ""
}

func (a *App) ExportTaskCSV(taskID, filePath string) string {
	records := a.taskQ.Records(taskID)
	if records == nil { return "任务未找到或无记录" }
	if err := exporter.ToCSV(records, filePath); err != nil { return "导出失败: " + err.Error() }
	return "成功导出至 " + filePath
}

func (a *App) ExportTaskDialog(taskID string) string {
	records := a.taskQ.Records(taskID)
	if records == nil { return "任务未找到或无记录" }
	if a.ctx == nil { return "应用未初始化" }

	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "导出CSV",
		DefaultFilename: taskID + ".csv",
		Filters: []runtime.FileFilter{
			{DisplayName: "CSV文件 (*.csv)", Pattern: "*.csv"},
		},
		CanCreateDirectories: true,
	})
	if err != nil { return "对话框失败: " + err.Error() }
	if path == "" { return "" }

	if err := exporter.ToCSV(records, path); err != nil { return "导出失败: " + err.Error() }
	return "成功导出至 " + path
}
