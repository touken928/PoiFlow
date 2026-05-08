package main

import (
	"context"
	"fmt"
	"os"

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
	if len(entries) == 0 {
		env := os.Getenv("BAIDU_AK")
		if env != "" {
			aks := splitCSV(env)
			a.akPool = akpool.New(aks, nil)
			_ = akstore.Save(akFileName, aks)
			return
		}
	}
	if a.akPool == nil {
		a.akPool = akpool.New(entries, nil)
	} else {
		a.akPool.Rebuild(entries, nil)
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

func (a *App) CreateTask(name, query, poiType, exportPath string, areaGran, queryGran int, targets []TaskTargetInput) *task.Task {
	ts := make([]task.Target, len(targets))
	for i, t := range targets {
		ts[i] = task.Target{Province: t.Province, City: t.City, Name: t.Name}
	}
	return a.taskQ.Add(name, query, poiType, exportPath, task.Granularity(areaGran), task.Granularity(queryGran), ts)
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
	AK     string `json:"ak"`
	Used   int    `json:"used"`
	Failed bool   `json:"failed"`
	FailMsg string `json:"failMsg"`
}

func (a *App) GetAKItems() []AKInfo {
	items := a.akPool.Items()
	out := make([]AKInfo, len(items))
	for i, it := range items {
		out[i] = AKInfo{AK: it.AK, Used: it.Used, Failed: it.Failed, FailMsg: it.FailMsg}
	}
	return out
}

func (a *App) ResetAKPool()   { a.akPool.ResetAll() }

func (a *App) AddAK(ak string) string {
	if ak == "" { return "AK不能为空" }
	aks, err := akstore.Load(akFileName)
	if err != nil { return "读取配置失败: " + err.Error() }
	for _, e := range aks {
		if e == ak { return "AK已存在" }
	}
	aks = append(aks, ak)
	if err := akstore.Save(akFileName, aks); err != nil { return "保存失败: " + err.Error() }
	a.reloadAKs()
	return ""
}

func (a *App) RemoveAK(ak string) string {
	aks, err := akstore.Load(akFileName)
	if err != nil { return "读取配置失败: " + err.Error() }
	filtered := make([]string, 0, len(aks))
	for _, e := range aks {
		if e != ak { filtered = append(filtered, e) }
	}
	if len(filtered) == len(aks) { return "AK不存在" }
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

func splitCSV(s string) []string {
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == ',' { if i > start { out = append(out, s[start:i]) }; start = i + 1 }
	}
	if start < len(s) { out = append(out, s[start:]) }
	return out
}
