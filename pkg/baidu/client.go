package baidu

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	defaultBaseURL = "https://api.map.baidu.com"
	regionPath     = "/place/v3/region"
	defaultTimeout = 30 * time.Second
)

// Client 百度地图API客户端
type Client struct {
	ak               string
	baseURL          string
	httpClient       *http.Client
	disableCoordConv bool
}

// ClientOption 客户端配置选项
type ClientOption func(*Client)

// WithHTTPClient 设置自定义HTTP客户端
func WithHTTPClient(hc *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = hc
	}
}

// WithBaseURL 设置自定义基础URL（用于测试）
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

// WithoutCoordConvert 关闭自动坐标系转换（默认开启BD09/GCJ02→WGS84）
func WithoutCoordConvert() ClientOption {
	return func(c *Client) {
		c.disableCoordConv = true
	}
}

// NewClient 创建百度地图API客户端
func NewClient(ak string, opts ...ClientOption) *Client {
	c := &Client{
		ak:      ak,
		baseURL: defaultBaseURL,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// RegionSearch 执行行政区划区域检索
// 详见 https://api.map.baidu.com/place/v3/region
func (c *Client) RegionSearch(req *RegionRequest) (*RegionResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request must not be nil")
	}
	if req.Query == "" && req.Type == "" {
		return nil, fmt.Errorf("query or type is required")
	}
	if req.Region == "" {
		return nil, fmt.Errorf("region is required")
	}

	params := req.toValues()
	params["ak"] = c.ak

	queryURL, err := url.JoinPath(c.baseURL, regionPath)
	if err != nil {
		return nil, fmt.Errorf("failed to build URL: %w", err)
	}

	u, err := url.Parse(queryURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}
	u.RawQuery = buildQuery(params)

	httpReq, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var result RegionResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if result.Status != 0 {
		return &result, &APIError{
			Status:  result.Status,
			Message: result.Message,
		}
	}

	if !c.disableCoordConv {
		result.ConvertToWGS84(req.RetCoordType)
	}

	return &result, nil
}

func buildQuery(params map[string]string) string {
	values := url.Values{}
	for k, v := range params {
		values.Set(k, v)
	}
	return values.Encode()
}

// Ping 发送最简查询验证AK是否可用
func (c *Client) Ping() error {
	ps := 1
	_, err := c.RegionSearch(&RegionRequest{Query: "测试", Region: "北京", PageSize: &ps})
	return err
}
