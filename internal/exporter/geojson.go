package exporter

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/touken928/PoiFlow/internal/task"
)

type geoFeature struct {
	Type       string                 `json:"type"`
	Geometry   geoPoint               `json:"geometry"`
	Properties map[string]interface{} `json:"properties"`
}

type geoPoint struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
}

type geoCollection struct {
	Type     string       `json:"type"`
	Features []geoFeature `json:"features"`
}

var geoPropGetters = map[string]func(task.Record) interface{}{
	"name":      func(r task.Record) interface{} { return r.Name },
	"address":   func(r task.Record) interface{} { return r.Address },
	"telephone": func(r task.Record) interface{} { return r.Telephone },
	"province":  func(r task.Record) interface{} { return r.Province },
	"city":      func(r task.Record) interface{} { return r.City },
	"area":      func(r task.Record) interface{} { return r.Area },
	"uid":       func(r task.Record) interface{} { return r.UID },
	"query":     func(r task.Record) interface{} { return r.Query },
	"type":      func(r task.Record) interface{} { return r.Type },
	"taskName":  func(r task.Record) interface{} { return r.TaskName },
	"target":    func(r task.Record) interface{} { return r.Target },
}

func ToGeoJSON(records []task.Record, filePath string) error {
	fields := []string{"name", "address", "telephone", "province", "city", "area", "uid", "query", "type", "taskName", "target"}
	return ToGeoJSONFiltered(records, filePath, fields)
}

func ToGeoJSONFiltered(records []task.Record, filePath string, fields []string) error {
	fc := geoCollection{Type: "FeatureCollection"}
	for _, r := range records {
		props := make(map[string]interface{})
		for _, f := range fields {
			if fn, ok := geoPropGetters[f]; ok {
				props[f] = fn(r)
			}
		}
		fc.Features = append(fc.Features, geoFeature{
			Type: "Feature",
			Geometry: geoPoint{
				Type:        "Point",
				Coordinates: []float64{r.Lng, r.Lat},
			},
			Properties: props,
		})
	}
	data, err := json.MarshalIndent(fc, "", "  ")
	if err != nil { return fmt.Errorf("failed to marshal geojson: %w", err) }
	return os.WriteFile(filePath, data, 0644)
}
