package baidu

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestServer(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}

func TestNewClient(t *testing.T) {
	client := NewClient("test_ak")
	if client.ak != "test_ak" {
		t.Errorf("expected ak 'test_ak', got '%s'", client.ak)
	}
	if client.baseURL != defaultBaseURL {
		t.Errorf("expected baseURL '%s', got '%s'", defaultBaseURL, client.baseURL)
	}
	if client.httpClient == nil {
		t.Error("expected non-nil http client")
	}
	if client.httpClient.Timeout != defaultTimeout {
		t.Errorf("expected timeout %v, got %v", defaultTimeout, client.httpClient.Timeout)
	}
}

func TestNewClientWithOptions(t *testing.T) {
	customHTTP := &http.Client{Timeout: 10}
	client := NewClient("ak", WithBaseURL("https://custom.example.com"), WithHTTPClient(customHTTP))

	if client.baseURL != "https://custom.example.com" {
		t.Errorf("expected custom baseURL, got '%s'", client.baseURL)
	}
	if client.httpClient != customHTTP {
		t.Error("expected custom http client")
	}
}

func TestRegionSearchSuccess(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/place/v3/region" {
			t.Errorf("expected path '/place/v3/region', got '%s'", r.URL.Path)
		}

		query := r.URL.Query()
		if query.Get("query") != "ATM" {
			t.Errorf("expected query 'ATM', got '%s'", query.Get("query"))
		}
		if query.Get("region") != "北京" {
			t.Errorf("expected region '北京', got '%s'", query.Get("region"))
		}
		if query.Get("output") != "json" {
			t.Errorf("expected output 'json', got '%s'", query.Get("output"))
		}
		if query.Get("ak") != "test_ak" {
			t.Errorf("expected ak 'test_ak', got '%s'", query.Get("ak"))
		}

		resp := RegionResponse{
			Status:     0,
			Message:    "ok",
			ResultType: "poi_type",
			QueryType:  "general",
			Results: []POIResult{
				{
					UID:      "test_uid_001",
					Name:     "测试POI",
					Location: Location{Lat: 40.04383967179688, Lng: 116.31584460688308},
					Province: "北京市",
					City:     "北京市",
					Area:     "海淀区",
					Town:     "上地街道",
					TownCode: intPtr(110108022),
					Address:  "上地信息路15号",
					Detail:   intPtr(1),
					Telephone: "010-12345678",
					DetailInfo: &DetailInfo{
						OverallRating: "4.5",
						Price:         "100",
						ShopHours:     "09:00-18:00",
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	client := NewClient("test_ak", WithBaseURL(server.URL), WithoutCoordConvert())
	result, err := client.RegionSearch(&RegionRequest{
		Query:  "ATM",
		Region: "北京",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Status != 0 {
		t.Errorf("expected status 0, got %d", result.Status)
	}
	if result.Message != "ok" {
		t.Errorf("expected message 'ok', got '%s'", result.Message)
	}
	if len(result.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result.Results))
	}

	r := result.Results[0]
	if r.UID != "test_uid_001" {
		t.Errorf("expected uid 'test_uid_001', got '%s'", r.UID)
	}
	if r.Name != "测试POI" {
		t.Errorf("expected name '测试POI', got '%s'", r.Name)
	}
	if r.Location.Lat != 40.04383967179688 {
		t.Errorf("expected lat 40.04383967179688, got %f", r.Location.Lat)
	}
	if r.Location.Lng != 116.31584460688308 {
		t.Errorf("expected lng 116.31584460688308, got %f", r.Location.Lng)
	}
	if *r.TownCode != 110108022 {
		t.Errorf("expected town_code 110108022, got %d", *r.TownCode)
	}
	if r.DetailInfo.OverallRating != "4.5" {
		t.Errorf("expected overall_rating '4.5', got '%s'", r.DetailInfo.OverallRating)
	}
}

func TestRegionSearch_AutoConvertEnabled(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		resp := RegionResponse{
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
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	client := NewClient("test_ak", WithBaseURL(server.URL))
	result, err := client.RegionSearch(&RegionRequest{
		Query:  "ATM",
		Region: "北京",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r := result.Results[0]
	if r.Location.Lng == 116.31584460688308 && r.Location.Lat == 40.04383967179688 {
		t.Error("auto-convert should change coordinates for Chinese locations")
	}
}

func TestRegionSearch_AutoConvertDisabled(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		resp := RegionResponse{
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
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	client := NewClient("test_ak", WithBaseURL(server.URL), WithoutCoordConvert())
	result, err := client.RegionSearch(&RegionRequest{
		Query:  "ATM",
		Region: "北京",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r := result.Results[0]
	if r.Location.Lng != 116.31584460688308 || r.Location.Lat != 40.04383967179688 {
		t.Error("WithoutCoordConvert should preserve original coordinates")
	}
}

func TestRegionSearch_GCJ02SourceAutoConvert(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		resp := RegionResponse{
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
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	client := NewClient("test_ak", WithBaseURL(server.URL))
	result, err := client.RegionSearch(&RegionRequest{
		Query:        "ATM",
		Region:       "北京",
		RetCoordType: "gcj02ll",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r := result.Results[0]
	if r.Location.Lng == 116.391 && r.Location.Lat == 39.907 {
		t.Error("auto-convert should change GCJ02 coordinates to WGS84")
	}
}

func TestRegionSearchAPIError(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		resp := RegionResponse{
			Status:  201,
			Message: "invalid parameter",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	client := NewClient("test_ak", WithBaseURL(server.URL))
	result, err := client.RegionSearch(&RegionRequest{
		Query:  "",
		Region: "",
	})

	if err == nil {
		t.Error("expected an error for nil request, got nil")
	}
	_ = result
}

func TestRegionSearchNilRequest(t *testing.T) {
	client := NewClient("test_ak")
	_, err := client.RegionSearch(nil)
	if err == nil {
		t.Error("expected error for nil request")
	}
	if !strings.Contains(err.Error(), "must not be nil") {
		t.Errorf("expected 'must not be nil' in error, got '%v'", err)
	}
}

func TestRegionSearchEmptyQuery(t *testing.T) {
	client := NewClient("test_ak")
	_, err := client.RegionSearch(&RegionRequest{
		Region: "北京",
	})
	if err == nil {
		t.Error("expected error for empty query")
	}
	if !strings.Contains(err.Error(), "query is required") {
		t.Errorf("expected 'query is required' in error, got '%v'", err)
	}
}

func TestRegionSearchEmptyRegion(t *testing.T) {
	client := NewClient("test_ak")
	_, err := client.RegionSearch(&RegionRequest{
		Query: "ATM",
	})
	if err == nil {
		t.Error("expected error for empty region")
	}
	if !strings.Contains(err.Error(), "region is required") {
		t.Errorf("expected 'region is required' in error, got '%v'", err)
	}
}

func TestRegionSearchHTTPError(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	})
	defer server.Close()

	client := NewClient("test_ak", WithBaseURL(server.URL))
	_, err := client.RegionSearch(&RegionRequest{
		Query:  "ATM",
		Region: "北京",
	})
	if err == nil {
		t.Error("expected error for 500 status")
	}
	if !strings.Contains(err.Error(), "unexpected status code: 500") {
		t.Errorf("expected status code 500 error, got '%v'", err)
	}
}

func TestRegionSearchInvalidJSON(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("not valid json"))
	})
	defer server.Close()

	client := NewClient("test_ak", WithBaseURL(server.URL))
	_, err := client.RegionSearch(&RegionRequest{
		Query:  "ATM",
		Region: "北京",
	})
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "failed to unmarshal") {
		t.Errorf("expected unmarshal error, got '%v'", err)
	}
}

func TestRegionSearchFullPayload(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()

		if query.Get("page_num") != "1" {
			t.Errorf("expected page_num '1', got '%s'", query.Get("page_num"))
		}
		if query.Get("page_size") != "15" {
			t.Errorf("expected page_size '15', got '%s'", query.Get("page_size"))
		}
		if query.Get("scope") != "2" {
			t.Errorf("expected scope '2', got '%s'", query.Get("scope"))
		}
		if query.Get("coord_type") != "1" {
			t.Errorf("expected coord_type '1', got '%s'", query.Get("coord_type"))
		}
		if query.Get("region_limit") != "true" {
			t.Errorf("expected region_limit 'true', got '%s'", query.Get("region_limit"))
		}
		if query.Get("extensions_adcode") != "true" {
			t.Errorf("expected extensions_adcode 'true', got '%s'", query.Get("extensions_adcode"))
		}
		if query.Get("photo_show") != "false" {
			t.Errorf("expected photo_show 'false', got '%s'", query.Get("photo_show"))
		}
		if query.Get("ret_coordtype") != "gcj02ll" {
			t.Errorf("expected ret_coordtype 'gcj02ll', got '%s'", query.Get("ret_coordtype"))
		}

		total := 150
		resp := RegionResponse{
			Status:     0,
			Message:    "ok",
			Total:      &total,
			ResultType: "poi_type",
			QueryType:  "precise",
			Results:    []POIResult{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	scope := ScopeDetail
	coordType := CoordTypeWGS84
	pageNum := 1
	pageSize := 15
	regionLimit := true
	extensionsAdcode := true
	photoShow := false

	client := NewClient("test_ak", WithBaseURL(server.URL))
	result, err := client.RegionSearch(&RegionRequest{
		Query:            "银行",
		Region:           "北京市海淀区",
		RegionLimit:      &regionLimit,
		Scope:            &scope,
		CoordType:        &coordType,
		PageNum:          &pageNum,
		PageSize:         &pageSize,
		ExtensionsAdcode: &extensionsAdcode,
		PhotoShow:        &photoShow,
		RetCoordType:     "gcj02ll",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Status != 0 {
		t.Errorf("expected status 0, got %d", result.Status)
	}
	if result.Total == nil || *result.Total != 150 {
		t.Errorf("expected total 150, got %v", result.Total)
	}
	if result.QueryType != "precise" {
		t.Errorf("expected query_type 'precise', got '%s'", result.QueryType)
	}
}

func TestRegionSearchDetailInfoFields(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		children := []POIChild{
			{
				UID:      "child_001",
				ShowName: "子点简称",
				Name:     "子点名称",
				Location: Location{Lat: 40.0, Lng: 116.0},
				Address:  "子点地址",
			},
		}
		d7Hot := 12345
		viewHeat := 100
		arrivalHeat := 200
		interactHeat := 300
		resp := RegionResponse{
			Status:  0,
			Message: "ok",
			Results: []POIResult{
				{
					UID:   "full_uid",
					Name:  "完整POI",
					Location: Location{Lat: 39.9, Lng: 116.4},
					Province:   "上海市",
					City:       "上海市",
					Area:       "浦东新区",
					Town:       "陆家嘴街道",
					TownCode:   intPtr(310115005),
					Adcode:     intPtr(310115),
					Address:    "世纪大道1号",
					Status:     "正常营业",
					Telephone:  "021-12345678",
					StreetID:   "street_001",
					Detail:     intPtr(1),
					DetailInfo: &DetailInfo{
						ClassifiedPOITag: "美食:中餐厅",
						NewAlias:         "别名",
						Type:             "cater",
						DetailURL:        "https://map.baidu.com/detail/123",
						ShopHours:        "10:00-22:00",
						Price:            "80",
						Label:            "5A级景区",
						OverallRating:    "4.8",
						ImageNum:         "20",
						CommentNum:       "1500",
						NaviLocation:     &Location{Lat: 39.900, Lng: 116.401},
						Brand:            "知名品牌",
						IndoorFloor:      "F2",
						Ranking:          "第1名",
						ParentID:         "parent_001",
						Photos: []Photo{
							{ImageURL: "https://example.com/photo1.jpg"},
						},
						BestTime:     "春季",
						SugTime:      "2小时",
						Description:  "这是一个很好的地方",
						Children:     children,
						D7HotValue:   &d7Hot,
						ViewHeat:     &viewHeat,
						ArrivalHeat:  &arrivalHeat,
						InteractHeat: &interactHeat,
						HeatTrend:    "增长10%",
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	client := NewClient("test_ak", WithBaseURL(server.URL), WithoutCoordConvert())
	result, err := client.RegionSearch(&RegionRequest{
		Query:  "美食",
		Region: "上海",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r := result.Results[0]

	if r.Name != "完整POI" {
		t.Errorf("expected Name '完整POI', got '%s'", r.Name)
	}
	if r.Province != "上海市" {
		t.Errorf("expected Province '上海市', got '%s'", r.Province)
	}
	if r.Area != "浦东新区" {
		t.Errorf("expected Area '浦东新区', got '%s'", r.Area)
	}
	if *r.TownCode != 310115005 {
		t.Errorf("expected TownCode 310115005, got %d", *r.TownCode)
	}
	if *r.Adcode != 310115 {
		t.Errorf("expected Adcode 310115, got %d", *r.Adcode)
	}
	if r.Status != "正常营业" {
		t.Errorf("expected Status '正常营业', got '%s'", r.Status)
	}

	di := r.DetailInfo
	if di == nil {
		t.Fatal("expected non-nil DetailInfo")
	}
	if di.ClassifiedPOITag != "美食:中餐厅" {
		t.Errorf("expected ClassifiedPOITag '美食:中餐厅', got '%s'", di.ClassifiedPOITag)
	}
	if di.Type != "cater" {
		t.Errorf("expected Type 'cater', got '%s'", di.Type)
	}
	if di.Price != "80" {
		t.Errorf("expected Price '80', got '%s'", di.Price)
	}
	if di.OverallRating != "4.8" {
		t.Errorf("expected OverallRating '4.8', got '%s'", di.OverallRating)
	}
	if di.Brand != "知名品牌" {
		t.Errorf("expected Brand '知名品牌', got '%s'", di.Brand)
	}
	if di.NaviLocation == nil {
		t.Error("expected non-nil NaviLocation")
	} else {
		if di.NaviLocation.Lat != 39.900 {
			t.Errorf("expected NaviLocation.Lat 39.900, got %f", di.NaviLocation.Lat)
		}
	}

	if len(di.Photos) != 1 || di.Photos[0].ImageURL != "https://example.com/photo1.jpg" {
		t.Errorf("expected 1 photo with correct URL, got %+v", di.Photos)
	}
	if len(di.Children) != 1 {
		t.Errorf("expected 1 child, got %d", len(di.Children))
	} else {
		if di.Children[0].UID != "child_001" {
			t.Errorf("expected child uid 'child_001', got '%s'", di.Children[0].UID)
		}
		if di.Children[0].ShowName != "子点简称" {
			t.Errorf("expected child ShowName '子点简称', got '%s'", di.Children[0].ShowName)
		}
	}

	if *di.D7HotValue != 12345 {
		t.Errorf("expected D7HotValue 12345, got %d", *di.D7HotValue)
	}
	if *di.ViewHeat != 100 {
		t.Errorf("expected ViewHeat 100, got %d", *di.ViewHeat)
	}
	if di.HeatTrend != "增长10%" {
		t.Errorf("expected HeatTrend '增长10%%', got '%s'", di.HeatTrend)
	}
}

func TestAPIError(t *testing.T) {
	apiErr := &APIError{Status: 201, Message: "invalid parameter"}
	if apiErr.Error() != "invalid parameter" {
		t.Errorf("expected Error() to return 'invalid parameter', got '%s'", apiErr.Error())
	}
}

func TestRegionRequestToValuesBasic(t *testing.T) {
	req := &RegionRequest{
		Query:  "酒店",
		Region: "北京",
	}
	values := req.toValues()

	if values["query"] != "酒店" {
		t.Errorf("expected query '酒店', got '%s'", values["query"])
	}
	if values["region"] != "北京" {
		t.Errorf("expected region '北京', got '%s'", values["region"])
	}
	if values["output"] != "json" {
		t.Errorf("expected output 'json', got '%s'", values["output"])
	}
	if _, ok := values["ak"]; ok {
		t.Error("expected no ak in request values (added by client)")
	}
}

func TestRegionRequestToValuesOptionalFields(t *testing.T) {
	regionLimit := true
	isLight := true
	scope := ScopeBasic
	coordType := CoordTypeGCJ02
	pageNum := 0
	pageSize := 20
	extAdcode := false
	addressResult := true

	req := &RegionRequest{
		Query:            "餐厅",
		Region:           "深圳",
		Type:             "美食",
		RegionLimit:      &regionLimit,
		IsLightVersion:   &isLight,
		Scope:            &scope,
		CoordType:        &coordType,
		Center:           "22.5431,114.0579",
		Filter:           "industry_type:cater",
		ExtensionsAdcode: &extAdcode,
		AddressResult:    &addressResult,
		PageNum:          &pageNum,
		PageSize:         &pageSize,
		FromLanguage:     "auto",
		RetCoordType:     "gcj02ll",
	}
	values := req.toValues()

	if values["query"] != "餐厅" {
		t.Errorf("expected query '餐厅', got '%s'", values["query"])
	}
	if values["region"] != "深圳" {
		t.Errorf("expected region '深圳', got '%s'", values["region"])
	}
	if values["type"] != "美食" {
		t.Errorf("expected type '美食', got '%s'", values["type"])
	}
	if values["region_limit"] != "true" {
		t.Errorf("expected region_limit 'true', got '%s'", values["region_limit"])
	}
	if values["is_light_version"] != "true" {
		t.Errorf("expected is_light_version 'true', got '%s'", values["is_light_version"])
	}
	if values["scope"] != "1" {
		t.Errorf("expected scope '1', got '%s'", values["scope"])
	}
	if values["coord_type"] != "2" {
		t.Errorf("expected coord_type '2', got '%s'", values["coord_type"])
	}
	if values["center"] != "22.5431,114.0579" {
		t.Errorf("expected center '22.5431,114.0579', got '%s'", values["center"])
	}
	if values["filter"] != "industry_type:cater" {
		t.Errorf("expected filter 'industry_type:cater', got '%s'", values["filter"])
	}
	if values["extensions_adcode"] != "false" {
		t.Errorf("expected extensions_adcode 'false', got '%s'", values["extensions_adcode"])
	}
	if values["address_result"] != "true" {
		t.Errorf("expected address_result 'true', got '%s'", values["address_result"])
	}
	if values["page_num"] != "0" {
		t.Errorf("expected page_num '0', got '%s'", values["page_num"])
	}
	if values["page_size"] != "20" {
		t.Errorf("expected page_size '20', got '%s'", values["page_size"])
	}
	if values["from_language"] != "auto" {
		t.Errorf("expected from_language 'auto', got '%s'", values["from_language"])
	}
	if values["ret_coordtype"] != "gcj02ll" {
		t.Errorf("expected ret_coordtype 'gcj02ll', got '%s'", values["ret_coordtype"])
	}
}

func TestRegionRequestToValuesNilOptionals(t *testing.T) {
	req := &RegionRequest{
		Query:  "学校",
		Region: "广州",
	}
	values := req.toValues()

	optionalFields := []string{
		"type", "region_limit", "is_light_version", "scope", "coord_type",
		"center", "filter", "extensions_adcode", "address_result", "photo_show",
		"from_language", "language", "page_num", "page_size", "ret_coordtype",
	}
	for _, field := range optionalFields {
		if _, ok := values[field]; ok {
			t.Errorf("expected '%s' to be absent for nil optional value", field)
		}
	}
}

func TestCoordTypeConstants(t *testing.T) {
	if int(CoordTypeWGS84) != 1 {
		t.Errorf("expected CoordTypeWGS84=1, got %d", CoordTypeWGS84)
	}
	if int(CoordTypeGCJ02) != 2 {
		t.Errorf("expected CoordTypeGCJ02=2, got %d", CoordTypeGCJ02)
	}
	if int(CoordTypeBD09LL) != 3 {
		t.Errorf("expected CoordTypeBD09LL=3, got %d", CoordTypeBD09LL)
	}
	if int(CoordTypeBD09MC) != 4 {
		t.Errorf("expected CoordTypeBD09MC=4, got %d", CoordTypeBD09MC)
	}
}

func TestScopeConstants(t *testing.T) {
	if int(ScopeBasic) != 1 {
		t.Errorf("expected ScopeBasic=1, got %d", ScopeBasic)
	}
	if int(ScopeDetail) != 2 {
		t.Errorf("expected ScopeDetail=2, got %d", ScopeDetail)
	}
}

func intPtr(i int) *int {
	return &i
}
