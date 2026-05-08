package baidu

import "strconv"

// CoordType 表示传入的坐标类型
type CoordType int

const (
	CoordTypeWGS84  CoordType = 1 // GPS经纬度
	CoordTypeGCJ02  CoordType = 2 // 国测局经纬度坐标
	CoordTypeBD09LL CoordType = 3 // 百度经纬度坐标（默认）
	CoordTypeBD09MC CoordType = 4 // 百度米制坐标
)

// Scope 检索结果详细程度
type Scope int

const (
	ScopeBasic   Scope = 1 // 返回基本信息
	ScopeDetail  Scope = 2 // 返回检索POI详细信息
)

// RegionRequest 行政区划区域检索请求参数
type RegionRequest struct {
	Query            string     `json:"query"`                       // 检索关键字（必选）
	Region           string     `json:"region"`                      // 检索行政区划区域（必选）
	RegionLimit      *bool      `json:"region_limit,omitempty"`      // 是否严格限制召回数据在区域内
	IsLightVersion   *bool      `json:"is_light_version,omitempty"`  // 轻量化检索
	Type             string     `json:"type,omitempty"`              // 对query召回结果进行二次筛选
	Center           string     `json:"center,omitempty"`            // 中心点坐标 "lat,lng"
	Scope            *Scope     `json:"scope,omitempty"`             // 检索结果详细程度
	CoordType        *CoordType `json:"coord_type,omitempty"`        // 传入的坐标类型
	Filter           string     `json:"filter,omitempty"`            // 检索排序条件
	ExtensionsAdcode *bool      `json:"extensions_adcode,omitempty"` // 是否召回国标行政区划编码
	AddressResult    *bool      `json:"address_result,omitempty"`    // 是否返回门址数据
	PhotoShow        *bool      `json:"photo_show,omitempty"`        // 是否输出图片信息
	FromLanguage     string     `json:"from_language,omitempty"`     // query的语言类型
	Language         string     `json:"language,omitempty"`          // 多语言检索
	PageNum          *int       `json:"page_num,omitempty"`          // 分页页码，从0开始
	PageSize         *int       `json:"page_size,omitempty"`         // 单次召回POI数量（10-20）
	RetCoordType     string     `json:"ret_coordtype,omitempty"`     // 返回的坐标类型，如 gcj02ll
}

// RegionResponse 行政区划区域检索返回结果
type RegionResponse struct {
	Status     int         `json:"status"`                // API访问状态，0表示成功
	Message    string      `json:"message"`               // 对API访问状态值的英文说明
	Total      *int        `json:"total,omitempty"`       // 召回poi数量（设置了page_num才会出现）
	ResultType string      `json:"result_type,omitempty"` // 召回结果类型
	QueryType  string      `json:"query_type,omitempty"`  // 搜索类型：精搜precise/泛搜general
	Results    []POIResult `json:"results"`               // 检索结果
}

// POIResult 表示单个POI检索结果
type POIResult struct {
	UID        string      `json:"uid"`                  // poi的唯一标示ID
	Name       string      `json:"name"`                 // poi名称
	Location   Location    `json:"location"`             // poi经纬度坐标
	Province   string      `json:"province,omitempty"`   // poi所属省份
	City       string      `json:"city,omitempty"`       // poi所属城市
	Area       string      `json:"area,omitempty"`       // poi所属区县
	Town       string      `json:"town,omitempty"`       // poi所属乡镇街道
	TownCode   *int        `json:"town_code,omitempty"`  // poi所属乡镇街道编码
	Adcode     *int        `json:"adcode,omitempty"`     // poi所属区域代码
	Address    string      `json:"address,omitempty"`    // poi所在地址
	Status     string      `json:"status,omitempty"`     // poi营业状态
	Telephone  string      `json:"telephone,omitempty"`  // poi的电话
	StreetID   string      `json:"street_id,omitempty"`  // poi所在街景图id
	Detail     *int        `json:"detail,omitempty"`     // 是否有详情页：1有，0没有
	DetailInfo *DetailInfo `json:"detail_info,omitempty"` // 详细信息
}

// Location 经纬度坐标
type Location struct {
	Lat float64 `json:"lat"` // 纬度值
	Lng float64 `json:"lng"` // 经度值
}

// DetailInfo POI详细信息
type DetailInfo struct {
	ClassifiedPOITag string       `json:"classified_poi_tag,omitempty"` // POI展示分类
	NewAlias         string       `json:"new_alias,omitempty"`          // poi别名
	Type             string       `json:"type,omitempty"`               // 类型（hotel、cater、life）
	DetailURL        string       `json:"detail_url,omitempty"`         // poi的详情页
	ShopHours        string       `json:"shop_hours,omitempty"`         // poi的营业时间
	Price            string       `json:"price,omitempty"`              // poi商户的价格
	Label            string       `json:"label,omitempty"`              // poi权威标签
	OverallRating    string       `json:"overall_rating,omitempty"`     // poi的综合评分
	ImageNum         string       `json:"image_num,omitempty"`          // poi图片数
	CommentNum       string       `json:"comment_num,omitempty"`        // poi的评论数
	NaviLocation     *Location    `json:"navi_location,omitempty"`      // poi对应的导航引导点坐标
	Brand            string       `json:"brand,omitempty"`              // poi对应的品牌
	IndoorFloor      string       `json:"indoor_floor,omitempty"`       // 室内poi所在楼层
	Ranking          string       `json:"ranking,omitempty"`            // poi的相关榜单排名
	ParentID         string       `json:"parent_id,omitempty"`          // poi父点id
	Photos           []Photo      `json:"photos,omitempty"`             // poi图片的下载链接
	BestTime         string       `json:"best_time,omitempty"`          // 最佳游玩时间
	SugTime          string       `json:"sug_time,omitempty"`           // 建议时长
	Description      string       `json:"description,omitempty"`        // 描述
	Children         []POIChild   `json:"children,omitempty"`           // poi子点
	D7HotValue       *int         `json:"d7_hot_value,omitempty"`       // poi近7天的热度总值
	ViewHeat         *int         `json:"view_heat,omitempty"`          // poi的浏览热度（高级权限）
	ArrivalHeat      *int         `json:"arrival_heat,omitempty"`       // poi的到达热度（高级权限）
	InteractHeat     *int         `json:"interact_heat,omitempty"`      // poi的互动热度（高级权限）
	HeatTrend        string       `json:"heat_trend,omitempty"`         // 近7天热度增长趋势
}

// Photo POI图片信息
type Photo struct {
	ImageURL string `json:"image_url,omitempty"` // 图片下载链接
}

// POIChild POI子点信息
type POIChild struct {
	UID               string   `json:"uid"`                           // poi子点ID
	ShowName          string   `json:"show_name,omitempty"`           // poi子点简称
	Name              string   `json:"name,omitempty"`                // poi子点名称
	ClassifiedPOITag  string   `json:"classified_poi_tag,omitempty"`  // poi子点详细分类标签
	Location          Location `json:"location,omitempty"`            // poi子点坐标
	Address           string   `json:"address,omitempty"`             // poi子点地址
}

// APIError 表示百度地图API返回的错误
type APIError struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

func (e *APIError) Error() string {
	return e.Message
}

// toValues 将RegionRequest转换为url.Values用于GET请求
func (r *RegionRequest) toValues() map[string]string {
	params := make(map[string]string)

	setIf := func(key, value string) {
		if value != "" {
			params[key] = value
		}
	}
	setBoolIf := func(key string, v *bool) {
		if v != nil {
			if *v {
				params[key] = "true"
			} else {
				params[key] = "false"
			}
		}
	}
	setIntIf := func(key string, v *int) {
		if v != nil {
			params[key] = strconv.Itoa(*v)
		}
	}
	setCoordTypeIf := func(key string, v *CoordType) {
		if v != nil {
			params[key] = strconv.Itoa(int(*v))
		}
	}
	setScopeIf := func(key string, v *Scope) {
		if v != nil {
			params[key] = strconv.Itoa(int(*v))
		}
	}

	setIf("query", r.Query)
	setIf("region", r.Region)
	setBoolIf("region_limit", r.RegionLimit)
	setBoolIf("is_light_version", r.IsLightVersion)
	setIf("type", r.Type)
	setIf("center", r.Center)
	setScopeIf("scope", r.Scope)
	setCoordTypeIf("coord_type", r.CoordType)
	setIf("filter", r.Filter)
	setBoolIf("extensions_adcode", r.ExtensionsAdcode)
	setBoolIf("address_result", r.AddressResult)
	setBoolIf("photo_show", r.PhotoShow)
	setIf("from_language", r.FromLanguage)
	setIf("language", r.Language)
	setIntIf("page_num", r.PageNum)
	setIntIf("page_size", r.PageSize)
	setIf("ret_coordtype", r.RetCoordType)
	params["output"] = "json"

	return params
}
