package division

import (
	"embed"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed china.yml
var chinaYML embed.FS

// Data 行政区划全量数据
// 第一层key为省/直辖市/自治区名称, 第二层key为市/区名称, value为区/县列表
type Data map[string]map[string][]string

// SearchResult 搜索结果
type SearchResult struct {
	Province string // 匹配到的省/直辖市/自治区
	City     string // 匹配到的市/区（省直辖时为空）
	County   string // 匹配到的区/县（市直辖时为空）
}

var divisions Data

func init() {
	content, err := chinaYML.ReadFile("china.yml")
	if err != nil {
		panic("failed to read embedded division data: " + err.Error())
	}
	if err := yaml.Unmarshal(content, &divisions); err != nil {
		panic("failed to parse division data: " + err.Error())
	}
}

// Provinces 返回所有省/直辖市/自治区名称列表
func Provinces() []string {
	names := make([]string, 0, len(divisions))
	for k := range divisions {
		names = append(names, k)
	}
	return names
}

// Cities 返回指定省/直辖市/自治区下的市/区名称列表
// 若省份不存在则返回空切片
func Cities(province string) []string {
	cities, ok := divisions[province]
	if !ok {
		return nil
	}
	names := make([]string, 0, len(cities))
	for k := range cities {
		names = append(names, k)
	}
	return names
}

// Counties 返回指定省/直辖市/自治区下指定市/区的区县列表
// 若省份或城市不存在则返回空切片
func Counties(province, city string) []string {
	cities, ok := divisions[province]
	if !ok {
		return nil
	}
	return cities[city]
}

// Search 按关键词搜索行政区划，匹配省/市/区县名称
// 返回所有匹配的结果，包括省名匹配、市名匹配、区县名匹配
func Search(keyword string) []SearchResult {
	var results []SearchResult
	for province, cities := range divisions {
		if match(province, keyword) {
			results = append(results, SearchResult{Province: province})
		}
		for city, counties := range cities {
			if match(city, keyword) {
				results = append(results, SearchResult{Province: province, City: city})
			}
			for _, county := range counties {
				if match(county, keyword) {
					results = append(results, SearchResult{
						Province: province,
						City:     city,
						County:   county,
					})
				}
			}
		}
	}
	return results
}

func match(s, keyword string) bool {
	return keyword != "" && strings.Contains(s, keyword)
}
