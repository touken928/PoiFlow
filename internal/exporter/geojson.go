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

func ToGeoJSON(records []task.Record, filePath string) error {
	fc := geoCollection{Type: "FeatureCollection"}
	for _, r := range records {
		fc.Features = append(fc.Features, geoFeature{
			Type: "Feature",
			Geometry: geoPoint{
				Type:        "Point",
				Coordinates: []float64{r.Lng, r.Lat},
			},
			Properties: map[string]interface{}{
				"name":      r.Name,
				"address":   r.Address,
				"telephone": r.Telephone,
				"province":  r.Province,
				"city":      r.City,
				"area":      r.Area,
				"uid":       r.UID,
				"query":     r.Query,
				"type":      r.Type,
				"taskName":  r.TaskName,
				"target":    r.Target,
			},
		})
	}
	data, err := json.MarshalIndent(fc, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal geojson: %w", err)
	}
	return os.WriteFile(filePath, data, 0644)
}
