package division

import (
	"testing"
)

func TestProvinces(t *testing.T) {
	provinces := Provinces()
	if len(provinces) == 0 {
		t.Error("expected non-empty provinces list")
	}
	hasBeijing := false
	for _, p := range provinces {
		if p == "北京市" {
			hasBeijing = true
			break
		}
	}
	if !hasBeijing {
		t.Error("expected '北京市' in provinces list")
	}
}

func TestCities(t *testing.T) {
	tests := []struct {
		province  string
		expectLen int
		expectHas string
	}{
		{"上海市", 1, "上海市"},
		{"云南省", 16, "昆明市"},
		{"内蒙古自治区", 12, "呼和浩特市"},
		{"北京市", 1, "北京市"},
		{"不存在省", 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.province, func(t *testing.T) {
			cities := Cities(tt.province)
			if tt.expectLen > 0 && len(cities) == 0 {
				t.Errorf("expected non-empty cities for %s", tt.province)
			}
			if tt.expectLen == 0 && cities != nil {
				t.Errorf("expected nil or empty cities for %s, got %v", tt.province, cities)
			}
			if tt.expectHas != "" {
				found := false
				for _, c := range cities {
					if c == tt.expectHas {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected %q in cities of %s, got %v", tt.expectHas, tt.province, cities)
				}
			}
		})
	}
}

func TestCounties(t *testing.T) {
	tests := []struct {
		province  string
		city      string
		expectLen int
		expectHas string
	}{
		{"上海市", "上海市", 16, "浦东新区"},
		{"云南省", "昆明市", 14, "五华区"},
		{"云南省", "丽江市", 5, "古城区"},
		{"内蒙古自治区", "呼和浩特市", 9, "新城区"},
		{"不存在省", "市", 0, ""},
		{"云南省", "不存在市", 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.province+"-"+tt.city, func(t *testing.T) {
			counties := Counties(tt.province, tt.city)
			if tt.expectLen > 0 && len(counties) == 0 {
				t.Errorf("expected non-empty counties for %s %s", tt.province, tt.city)
			}
			if tt.expectLen == 0 && counties != nil && len(counties) > 0 {
				t.Errorf("expected nil or empty counties for %s %s, got %v", tt.province, tt.city, counties)
			}
			if tt.expectHas != "" {
				found := false
				for _, c := range counties {
					if c == tt.expectHas {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected %q in counties of %s %s, got %v", tt.expectHas, tt.province, tt.city, counties)
				}
			}
		})
	}
}

func TestSearch(t *testing.T) {
	results := Search("昆明")
	if len(results) == 0 {
		t.Error("expected search results for '昆明'")
	}
	hasKunming := false
	for _, r := range results {
		if r.Province == "云南省" && r.City == "昆明市" {
			hasKunming = true
		}
	}
	if !hasKunming {
		t.Error("expected '昆明市' city match in search results")
	}
}

func TestSearchCounty(t *testing.T) {
	results := Search("浦东新区")
	if len(results) == 0 {
		t.Fatal("expected search results for '浦东新区'")
	}
	r := results[0]
	if r.Province != "上海市" {
		t.Errorf("expected province '上海市', got %q", r.Province)
	}
	if r.City != "上海市" {
		t.Errorf("expected city '上海市', got %q", r.City)
	}
	if r.County != "浦东新区" {
		t.Errorf("expected county '浦东新区', got %q", r.County)
	}
}

func TestSearchEmptyKeyword(t *testing.T) {
	results := Search("")
	if len(results) != 0 {
		t.Errorf("expected empty results for empty keyword, got %d results", len(results))
	}
}

func TestSearchNoMatch(t *testing.T) {
	results := Search("asdfghjklqwertyuio")
	if len(results) != 0 {
		t.Errorf("expected no results for nonsense keyword, got %d results", len(results))
	}
}

func TestSearchProvinceOnly(t *testing.T) {
	results := Search("北京")
	if len(results) == 0 {
		t.Fatal("expected search results for '北京'")
	}
	found := false
	for _, r := range results {
		if r.Province == "北京市" && r.City == "" && r.County == "" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected province-only match for '北京市'")
	}
}

func TestCitiesNonExistentProvince(t *testing.T) {
	cities := Cities("不存在的省")
	if cities != nil {
		t.Errorf("expected nil for non-existent province, got %v", cities)
	}
}

func TestCountiesNonExistentProvince(t *testing.T) {
	counties := Counties("不存在的省", "某市")
	if counties != nil {
		t.Errorf("expected nil for non-existent province, got %v", counties)
	}
}

func TestCountiesNonExistentCity(t *testing.T) {
	counties := Counties("云南省", "不存在的市")
	if counties != nil {
		t.Errorf("expected nil for non-existent city, got %v", counties)
	}
}
