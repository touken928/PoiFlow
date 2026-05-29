package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/touken928/PoiFlow/internal/akpool"
	"github.com/touken928/PoiFlow/internal/exporter"
	"github.com/touken928/PoiFlow/internal/store"
	"github.com/touken928/PoiFlow/internal/task"
	"github.com/touken928/PoiFlow/pkg/division"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const akFileName = "config.yaml"

var Version = "dev"

func (a *App) GetVersion() string { return Version }

func poiflowDir() string {
	if d := os.Getenv("POIFLOW_DIR"); d != "" { return d }
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".poiflow")
}
func cacheDir() string     { return filepath.Join(poiflowDir(), "cache") }
func statePath() string    { return filepath.Join(poiflowDir(), "tasks.json") }
func configPath() string   { return filepath.Join(poiflowDir(), akFileName) }

type App struct {
	ctx    context.Context
	akPool *akpool.Pool
	taskQ  *task.Queue
}

func NewApp() *App {
	godotenv.Load()
	a := &App{}
	a.reloadAKs()
	executor := task.NewExecutor(a.akPool)
	a.taskQ = task.NewQueue(executor, a.onTaskEvent, cacheDir(), statePath())
	return a
}

func (a *App) startup(ctx context.Context) { a.ctx = ctx }

func (a *App) onSecondInstanceLaunch(secondInstanceData options.SecondInstanceData) {
	println("检测到第二个实例启动，参数:", strings.Join(secondInstanceData.Args, ","))
	if a.ctx != nil {
		runtime.WindowUnminimise(a.ctx)
		runtime.Show(a.ctx)
	}
}

func (a *App) reloadAKs() {
	entries, err := store.LoadAKs(configPath())
	if err != nil { println("load aks failed:", err.Error()) }
	keys := make([]string, len(entries))
	names := make([]string, len(entries))
	for i, e := range entries { keys[i] = e.Key; names[i] = e.Name }
	if a.akPool == nil {
		a.akPool = akpool.NewWithNames(keys, names, nil)
	} else {
		a.akPool.RebuildWithNames(keys, names, nil)
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
func (a *App) RetryTask(id string) bool   { return a.taskQ.Retry(id) }
func (a *App) PauseTask(id string) bool   { return a.taskQ.Pause(id) }
func (a *App) ResumeTask(id string) bool  { return a.taskQ.Resume(id) }
func (a *App) GetTaskLogs(id string) []task.LogEntry { return a.taskQ.GetLogs(id) }

func (a *App) GetTaskRecords(id string) []task.Record {
	r := a.taskQ.Records(id)
	if r == nil {
		return []task.Record{}
	}
	return r
}

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
	entries, _ := store.LoadAKs(configPath())
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
	entries, err := store.LoadAKs(configPath())
	if err != nil { return "读取配置失败: " + err.Error() }
	for _, e := range entries {
		if e.Key == key { return "AK已存在" }
	}
	entries = append(entries, store.Entry{Name: name, Key: key})
	if err := store.SaveAKs(configPath(), entries); err != nil { return "保存失败: " + err.Error() }
	a.reloadAKs()
	return ""
}

func (a *App) RemoveAK(key string) string {
	entries, err := store.LoadAKs(configPath())
	if err != nil { return "读取配置失败: " + err.Error() }
	filtered := make([]store.Entry, 0, len(entries))
	for _, e := range entries {
		if e.Key != key { filtered = append(filtered, e) }
	}
	if len(filtered) == len(entries) { return "AK不存在" }
	if err := store.SaveAKs(configPath(), filtered); err != nil { return "保存失败: " + err.Error() }
	a.reloadAKs()
	return ""
}

func (a *App) GetExportConfig() store.ExportConfig {
	ec, err := store.LoadExportConfig(configPath())
	if err != nil { return store.DefaultExportConfig() }
	if len(ec.Fields) == 0 { return store.DefaultExportConfig() }
	return ec
}

func (a *App) SetExportConfig(fields []string) string {
	ec := store.ExportConfig{}
	for _, f := range fields {
		ec.Fields = append(ec.Fields, store.ExportField(f))
	}
	if err := store.SaveExportConfig(configPath(), ec); err != nil {
		return "保存失败: " + err.Error()
	}
	return ""
}

func (a *App) ExportTaskCSV(taskID, filePath string) string {
	records := a.taskQ.Records(taskID)
	if records == nil { return "任务未找到或无记录" }
	ec, _ := store.LoadExportConfig(configPath())
	fields := exportFieldsToStrings(ec)
	if err := exporter.ToCSVFiltered(records, filePath, fields); err != nil { return "导出失败: " + err.Error() }
	return "成功导出至 " + filePath
}

func (a *App) ExportTaskDialog(taskID string) string {
	records := a.taskQ.Records(taskID)
	if records == nil { return "任务未找到或无记录" }
	if a.ctx == nil { return "应用未初始化" }

	taskName := taskID
	for _, t := range a.taskQ.List() {
		if t.ID == taskID && t.Name != "" { taskName = t.Name; break }
	}

	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "导出CSV",
		DefaultFilename: sanitizeFilename(taskName) + ".csv",
		Filters: []runtime.FileFilter{
			{DisplayName: "CSV文件 (*.csv)", Pattern: "*.csv"},
		},
		CanCreateDirectories: true,
	})
	if err != nil { return "对话框失败: " + err.Error() }
	if path == "" { return "" }

	ec, _ := store.LoadExportConfig(configPath())
	fields := exportFieldsToStrings(ec)
	if err := exporter.ToCSVFiltered(records, path, fields); err != nil { return "导出失败: " + err.Error() }
	return "成功导出至 " + path
}

func (a *App) ExportTaskGeoJSON(taskID string) string {
	records := a.taskQ.Records(taskID)
	if records == nil { return "任务未找到或无记录" }
	if a.ctx == nil { return "应用未初始化" }

	taskName := taskID
	for _, t := range a.taskQ.List() {
		if t.ID == taskID && t.Name != "" { taskName = t.Name; break }
	}

	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "导出GeoJSON",
		DefaultFilename: sanitizeFilename(taskName) + ".geojson",
		Filters: []runtime.FileFilter{
			{DisplayName: "GeoJSON文件 (*.geojson)", Pattern: "*.geojson"},
			{DisplayName: "JSON文件 (*.json)", Pattern: "*.json"},
		},
		CanCreateDirectories: true,
	})
	if err != nil { return "对话框失败: " + err.Error() }
	if path == "" { return "" }

	ec, _ := store.LoadExportConfig(configPath())
	fields := exportFieldsToStrings(ec)
	if err := exporter.ToGeoJSONFiltered(records, path, fields); err != nil { return "导出失败: " + err.Error() }
	return "成功导出至 " + path
}

func exportFieldsToStrings(ec store.ExportConfig) []string {
	out := make([]string, len(ec.Fields))
	for i, f := range ec.Fields { out[i] = string(f) }
	return out
}

func sanitizeFilename(s string) string {
	b := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '/' || c == '\\' || c == ':' || c == '*' || c == '?' || c == '"' || c == '<' || c == '>' || c == '|' || c == '\x00' {
			b = append(b, '_')
		} else {
			b = append(b, c)
		}
	}
	return string(b)
}

func (a *App) ImportSearchTerms() string {
	if a.ctx == nil { return "" }
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "导入搜索词",
		Filters: []runtime.FileFilter{
			{DisplayName: "CSV或文本文件 (*.csv,*.txt)", Pattern: "*.csv;*.txt"},
			{DisplayName: "所有文件 (*.*)", Pattern: "*.*"},
		},
	})
	if err != nil { return "" }
	if path == "" { return "" }
	data, err := os.ReadFile(path)
	if err != nil { return "" }
	return string(data)
}
