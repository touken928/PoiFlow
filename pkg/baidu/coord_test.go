package baidu

import (
	"testing"
)

func TestIsOutOfChina(t *testing.T) {
	tests := []struct {
		name    string
		lng     float64
		lat     float64
		isOut   bool
	}{
		{"北京天安门", 116.397451, 39.909187, false},
		{"上海外滩", 121.490486, 31.235191, false},
		{"北京上地", 116.31584460688308, 40.04383967179688, false},
		{"乌鲁木齐", 87.616828, 43.825592, false},
		{"哈尔滨", 126.642105, 45.756491, false},
		{"三亚", 109.512174, 18.252838, false},
		{"北边界外", 116.0, 54.0, true},
		{"南边界外", 116.0, 3.0, true},
		{"西边界外", 70.0, 40.0, true},
		{"东边界外", 140.0, 40.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsOutOfChina(tt.lng, tt.lat)
			if got != tt.isOut {
				t.Errorf("IsOutOfChina(%f, %f) = %v, want %v", tt.lng, tt.lat, got, tt.isOut)
			}
		})
	}
}

func TestBD09ToGCJ02_Deterministic(t *testing.T) {
	lng, lat := 116.31584460688308, 40.04383967179688
	a1, b1 := BD09ToGCJ02(lng, lat)
	a2, b2 := BD09ToGCJ02(lng, lat)
	if a1 != a2 || b1 != b2 {
		t.Errorf("BD09ToGCJ02 not deterministic: (%f,%f) vs (%f,%f)", a1, b1, a2, b2)
	}
}

func TestBD09ToGCJ02_DiffersFromInput(t *testing.T) {
	lng, lat := 116.31584460688308, 40.04383967179688
	newLng, newLat := BD09ToGCJ02(lng, lat)
	if newLng == lng && newLat == lat {
		t.Error("BD09ToGCJ02 should produce different values for Chinese coordinates")
	}
}

func TestBD09ToGCJ02_OutOfChina(t *testing.T) {
	// Tokyo
	lng, lat := 139.6503, 35.6762
	newLng, newLat := BD09ToGCJ02(lng, lat)
	if newLng != lng || newLat != lat {
		t.Errorf("BD09ToGCJ02 for out-of-China coords should return same values: got (%f,%f), want (%f,%f)", newLng, newLat, lng, lat)
	}
}

func TestGCJ02ToWGS84_Deterministic(t *testing.T) {
	lng, lat := 116.391, 39.907
	a1, b1 := GCJ02ToWGS84(lng, lat)
	a2, b2 := GCJ02ToWGS84(lng, lat)
	if a1 != a2 || b1 != b2 {
		t.Errorf("GCJ02ToWGS84 not deterministic: (%f,%f) vs (%f,%f)", a1, b1, a2, b2)
	}
}

func TestGCJ02ToWGS84_DiffersFromInput(t *testing.T) {
	lng, lat := 116.391, 39.907
	newLng, newLat := GCJ02ToWGS84(lng, lat)
	if newLng == lng && newLat == lat {
		t.Error("GCJ02ToWGS84 should produce different values for Chinese coordinates")
	}
}

func TestGCJ02ToWGS84_OutOfChina(t *testing.T) {
	// London
	lng, lat := -0.118092, 51.509865
	newLng, newLat := GCJ02ToWGS84(lng, lat)
	if newLng != lng || newLat != lat {
		t.Errorf("GCJ02ToWGS84 for out-of-China coords should return same values: got (%f,%f), want (%f,%f)", newLng, newLat, lng, lat)
	}
}

func TestBD09ToWGS84_OutOfChina(t *testing.T) {
	// New York
	lng, lat := -74.006065, 40.712776
	newLng, newLat := BD09ToWGS84(lng, lat)
	if newLng != lng || newLat != lat {
		t.Errorf("BD09ToWGS84 for out-of-China coords should return same values: got (%f,%f), want (%f,%f)", newLng, newLat, lng, lat)
	}
}

func TestBD09ToWGS84_Deterministic(t *testing.T) {
	lng, lat := 116.31584460688308, 40.04383967179688
	a1, b1 := BD09ToWGS84(lng, lat)
	a2, b2 := BD09ToWGS84(lng, lat)
	if a1 != a2 || b1 != b2 {
		t.Errorf("BD09ToWGS84 not deterministic: (%f,%f) vs (%f,%f)", a1, b1, a2, b2)
	}
}

func TestBD09ToWGS84_DiffersFromBD09ToGCJ02(t *testing.T) {
	lng, lat := 116.31584460688308, 40.04383967179688
	gcjLng, gcjLat := BD09ToGCJ02(lng, lat)
	wgsLng, wgsLat := BD09ToWGS84(lng, lat)
	if wgsLng == gcjLng && wgsLat == gcjLat {
		t.Error("BD09ToWGS84 should differ from BD09ToGCJ02 for Chinese coordinates")
	}
}

func TestConvertToWGS84_BD09Source(t *testing.T) {
	resp := &RegionResponse{
		Status:  0,
		Message: "ok",
		Results: []POIResult{
			{
				UID:      "uid001",
				Name:     "测试",
				Location: Location{Lng: 116.31584460688308, Lat: 40.04383967179688},
			},
		},
	}

	origLng := resp.Results[0].Location.Lng
	origLat := resp.Results[0].Location.Lat

	resp.ConvertToWGS84("")

	if resp.Results[0].Location.Lng == origLng && resp.Results[0].Location.Lat == origLat {
		t.Error("ConvertToWGS84 with BD09 source should change coordinates")
	}
}

func TestConvertToWGS84_GCJ02Source(t *testing.T) {
	resp := &RegionResponse{
		Status:  0,
		Message: "ok",
		Results: []POIResult{
			{
				UID:      "uid001",
				Name:     "测试",
				Location: Location{Lng: 116.391, Lat: 39.907},
			},
		},
	}

	origLng := resp.Results[0].Location.Lng
	origLat := resp.Results[0].Location.Lat

	resp.ConvertToWGS84("gcj02ll")

	if resp.Results[0].Location.Lng == origLng && resp.Results[0].Location.Lat == origLat {
		t.Error("ConvertToWGS84 with GCJ02 source should change coordinates")
	}
}

func TestConvertToWGS84_OutOfChina_Skip(t *testing.T) {
	lng := -74.006065
	lat := 40.712776
	resp := &RegionResponse{
		Status:  0,
		Message: "ok",
		Results: []POIResult{
			{
				UID:      "uid001",
				Name:     "NYC",
				Location: Location{Lng: lng, Lat: lat},
			},
		},
	}

	resp.ConvertToWGS84("")

	if resp.Results[0].Location.Lng != lng || resp.Results[0].Location.Lat != lat {
		t.Errorf("ConvertToWGS84 should not change out-of-China coordinates: got (%f,%f), want (%f,%f)",
			resp.Results[0].Location.Lng, resp.Results[0].Location.Lat, lng, lat)
	}
}

func TestConvertToWGS84_NaviLocation(t *testing.T) {
	naviLng := 116.31584460688308
	naviLat := 40.04383967179688
	resp := &RegionResponse{
		Status:  0,
		Message: "ok",
		Results: []POIResult{
			{
				UID:      "uid001",
				Name:     "测试",
				Location: Location{Lng: 116.0, Lat: 40.0},
				DetailInfo: &DetailInfo{
					NaviLocation: &Location{Lng: naviLng, Lat: naviLat},
				},
			},
		},
	}

	resp.ConvertToWGS84("")

	navi := resp.Results[0].DetailInfo.NaviLocation
	if navi.Lng == naviLng && navi.Lat == naviLat {
		t.Error("ConvertToWGS84 should change NaviLocation coordinates")
	}
}

func TestConvertToWGS84_Children(t *testing.T) {
	childLng := 116.31584460688308
	childLat := 40.04383967179688
	resp := &RegionResponse{
		Status:  0,
		Message: "ok",
		Results: []POIResult{
			{
				UID:      "uid001",
				Name:     "父POI",
				Location: Location{Lng: 116.0, Lat: 40.0},
				DetailInfo: &DetailInfo{
					Children: []POIChild{
						{
							UID:      "child001",
							Name:     "子POI",
							Location: Location{Lng: childLng, Lat: childLat},
						},
					},
				},
			},
		},
	}

	resp.ConvertToWGS84("")

	child := resp.Results[0].DetailInfo.Children[0]
	if child.Location.Lng == childLng && child.Location.Lat == childLat {
		t.Error("ConvertToWGS84 should change Children Location coordinates")
	}
}

func TestConvertToWGS84_MultipleResults(t *testing.T) {
	lng1, lat1 := 116.31584460688308, 40.04383967179688
	lng2, lat2 := 121.490486, 31.235191
	resp := &RegionResponse{
		Status:  0,
		Message: "ok",
		Results: []POIResult{
			{UID: "uid001", Name: "北京", Location: Location{Lng: lng1, Lat: lat1}},
			{UID: "uid002", Name: "上海", Location: Location{Lng: lng2, Lat: lat2}},
		},
	}

	resp.ConvertToWGS84("")

	if resp.Results[0].Location.Lng == lng1 && resp.Results[0].Location.Lat == lat1 {
		t.Error("ConvertToWGS84 should change result 0 coordinates")
	}
	if resp.Results[1].Location.Lng == lng2 && resp.Results[1].Location.Lat == lat2 {
		t.Error("ConvertToWGS84 should change result 1 coordinates")
	}
}

func TestConvertToWGS84_NilDetailInfo(t *testing.T) {
	resp := &RegionResponse{
		Status:  0,
		Message: "ok",
		Results: []POIResult{
			{
				UID:      "uid001",
				Name:     "无详情",
				Location: Location{Lng: 116.31584460688308, Lat: 40.04383967179688},
				DetailInfo: nil,
			},
		},
	}

	// Should not panic
	resp.ConvertToWGS84("")
}
