package baidu

import "math"

const (
	xPi   = math.Pi * 3000.0 / 180.0
	axisA = 6378245.0
	ee    = 0.00669342162296594323
)

// IsOutOfChina 判断坐标是否在中国境外
func IsOutOfChina(lng, lat float64) bool {
	return lng < 73.66 || lng > 135.05 || lat < 3.86 || lat > 53.55
}

func transformLat(lng, lat float64) float64 {
	pi := math.Pi
	ret := -100.0 + 2.0*lng + 3.0*lat + 0.2*lat*lat + 0.1*lng*lat + 0.2*math.Sqrt(math.Abs(lng))
	ret += (20.0*math.Sin(6.0*lng*pi) + 20.0*math.Sin(2.0*lng*pi)) * 2.0 / 3.0
	ret += (20.0*math.Sin(lat*pi) + 40.0*math.Sin(lat/3.0*pi)) * 2.0 / 3.0
	ret += (160.0*math.Sin(lat/12.0*pi) + 320.0*math.Sin(lat*pi/30.0)) * 2.0 / 3.0
	return ret
}

func transformLng(lng, lat float64) float64 {
	pi := math.Pi
	ret := 300.0 + lng + 2.0*lat + 0.1*lng*lng + 0.1*lng*lat + 0.1*math.Sqrt(math.Abs(lng))
	ret += (20.0*math.Sin(6.0*lng*pi) + 20.0*math.Sin(2.0*lng*pi)) * 2.0 / 3.0
	ret += (20.0*math.Sin(lng*pi) + 40.0*math.Sin(lng/3.0*pi)) * 2.0 / 3.0
	ret += (150.0*math.Sin(lng/12.0*pi) + 300.0*math.Sin(lng/30.0*pi)) * 2.0 / 3.0
	return ret
}

// BD09ToGCJ02 百度坐标系(BD09)转换为国测局坐标系(GCJ02)
func BD09ToGCJ02(lng, lat float64) (float64, float64) {
	if IsOutOfChina(lng, lat) {
		return lng, lat
	}
	x := lng - 0.0065
	y := lat - 0.006
	z := math.Sqrt(x*x+y*y) - 0.00002*math.Sin(y*xPi)
	theta := math.Atan2(y, x) - 0.000003*math.Cos(x*xPi)
	return z * math.Cos(theta), z * math.Sin(theta)
}

// GCJ02ToWGS84 国测局坐标系(GCJ02)转换为WGS84坐标系
func GCJ02ToWGS84(lng, lat float64) (float64, float64) {
	if IsOutOfChina(lng, lat) {
		return lng, lat
	}
	pi := math.Pi
	dlat := transformLat(lng-105.0, lat-35.0)
	dlng := transformLng(lng-105.0, lat-35.0)
	radlat := lat / 180.0 * pi
	magic := math.Sin(radlat)
	magic = 1 - ee*magic*magic
	sqrtMagic := math.Sqrt(magic)
	dlat = (dlat * 180.0) / ((axisA * (1 - ee)) / (magic * sqrtMagic) * pi)
	dlng = (dlng * 180.0) / (axisA / sqrtMagic * math.Cos(radlat) * pi)
	mglat := lat + dlat
	mglng := lng + dlng
	return lng*2 - mglng, lat*2 - mglat
}

// BD09ToWGS84 百度坐标系(BD09)转换为WGS84坐标系
func BD09ToWGS84(lng, lat float64) (float64, float64) {
	lng, lat = BD09ToGCJ02(lng, lat)
	return GCJ02ToWGS84(lng, lat)
}

// ConvertToWGS84 将RegionResponse中所有坐标字段从指定源坐标系转换为WGS84
// sourceCoordType 可选值: "" 或 "bd09ll"(默认BD09), "gcj02ll"
func (r *RegionResponse) ConvertToWGS84(sourceCoordType string) {
	for i := range r.Results {
		r.Results[i].convertToWGS84(sourceCoordType)
	}
}

func (p *POIResult) convertToWGS84(sourceCoordType string) {
	fn := BD09ToWGS84
	if sourceCoordType == "gcj02ll" {
		fn = GCJ02ToWGS84
	}

	p.Location.Lng, p.Location.Lat = fn(p.Location.Lng, p.Location.Lat)

	if p.DetailInfo == nil {
		return
	}
	if p.DetailInfo.NaviLocation != nil {
		p.DetailInfo.NaviLocation.Lng, p.DetailInfo.NaviLocation.Lat =
			fn(p.DetailInfo.NaviLocation.Lng, p.DetailInfo.NaviLocation.Lat)
	}
	for j := range p.DetailInfo.Children {
		p.DetailInfo.Children[j].Location.Lng, p.DetailInfo.Children[j].Location.Lat =
			fn(p.DetailInfo.Children[j].Location.Lng, p.DetailInfo.Children[j].Location.Lat)
	}
}
