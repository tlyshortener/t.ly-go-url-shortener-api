package tly

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// Client is the main API client for T.LY.
type Client struct {
	APIKey  string
	BaseURL string
	Client  *http.Client
}

// APIError is returned when the T.LY API responds with a non-2xx status.
type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error (status %d): %s", e.StatusCode, e.Body)
}

// NewClient creates a new T.LY API client.
func NewClient(apiKey string) *Client {
	return &Client{
		APIKey:  apiKey,
		BaseURL: "https://api.t.ly",
		Client:  &http.Client{},
	}
}

func (c *Client) doRequestRaw(method, path string, query url.Values, body interface{}) ([]byte, error) {
	requestURL := strings.TrimRight(c.BaseURL, "/") + path
	if query != nil && len(query) > 0 {
		requestURL += "?" + query.Encode()
	}

	var reqBody *bytes.Buffer
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(data)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req, err := http.NewRequest(method, requestURL, reqBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		return nil, readErr
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Body:       string(data),
		}
	}
	return data, nil
}

// doRequest is an internal helper for making API calls and decoding JSON responses.
func (c *Client) doRequest(method, path string, query url.Values, body interface{}, result interface{}) error {
	data, err := c.doRequestRaw(method, path, query, body)
	if err != nil {
		return err
	}

	if result == nil || len(bytes.TrimSpace(data)) == 0 {
		return nil
	}
	if byteTarget, ok := result.(*[]byte); ok {
		*byteTarget = append((*byteTarget)[:0], data...)
		return nil
	}

	return json.Unmarshal(data, result)
}

func queryFromMap(params map[string]string) url.Values {
	query := url.Values{}
	for k, v := range params {
		query.Set(k, v)
	}
	return query
}

func addIndexedIntSlice(query url.Values, key string, values []int) {
	for i, v := range values {
		query.Set(fmt.Sprintf("%s[%d]", key, i), strconv.Itoa(v))
	}
}

// =====================
// Pixel Management
// =====================

// Pixel represents a pixel object.
type Pixel struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	PixelID   string `json:"pixel_id"`
	PixelType string `json:"pixel_type"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// PixelCreateRequest is used to create a new pixel.
type PixelCreateRequest struct {
	Name      string `json:"name"`
	PixelID   string `json:"pixel_id"`
	PixelType string `json:"pixel_type"`
}

// PixelUpdateRequest is used to update a pixel.
type PixelUpdateRequest struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	PixelID   string `json:"pixel_id"`
	PixelType string `json:"pixel_type"`
}

// CreatePixel calls the API to create a new pixel.
func (c *Client) CreatePixel(reqData PixelCreateRequest) (*Pixel, error) {
	var pixel Pixel
	err := c.doRequest(http.MethodPost, "/api/v1/link/pixel", nil, reqData, &pixel)
	if err != nil {
		return nil, err
	}
	return &pixel, nil
}

// ListPixels retrieves a list of pixels.
func (c *Client) ListPixels() ([]Pixel, error) {
	var pixels []Pixel
	err := c.doRequest(http.MethodGet, "/api/v1/link/pixel", nil, nil, &pixels)
	if err != nil {
		return nil, err
	}
	return pixels, nil
}

// GetPixel retrieves a pixel by its ID.
func (c *Client) GetPixel(id int) (*Pixel, error) {
	path := fmt.Sprintf("/api/v1/link/pixel/%d", id)
	var pixel Pixel
	err := c.doRequest(http.MethodGet, path, nil, nil, &pixel)
	if err != nil {
		return nil, err
	}
	return &pixel, nil
}

// UpdatePixel updates an existing pixel.
func (c *Client) UpdatePixel(reqData PixelUpdateRequest) (*Pixel, error) {
	path := fmt.Sprintf("/api/v1/link/pixel/%d", reqData.ID)
	var pixel Pixel
	err := c.doRequest(http.MethodPut, path, nil, reqData, &pixel)
	if err != nil {
		return nil, err
	}
	return &pixel, nil
}

// DeletePixel deletes a pixel by its ID.
func (c *Client) DeletePixel(id int) error {
	path := fmt.Sprintf("/api/v1/link/pixel/%d", id)
	return c.doRequest(http.MethodDelete, path, nil, nil, nil)
}

// =====================
// Short Link Management
// =====================

// ShortLink represents a shortened URL.
type ShortLink struct {
	ShortURL         string        `json:"short_url"`
	Description      *string       `json:"description"`
	LongURL          string        `json:"long_url"`
	Domain           string        `json:"domain"`
	ShortID          string        `json:"short_id"`
	ExpireAtViews    interface{}   `json:"expire_at_views"`
	ExpireAtDatetime interface{}   `json:"expire_at_datetime"`
	PublicStats      bool          `json:"public_stats"`
	CreatedAt        string        `json:"created_at"`
	UpdatedAt        string        `json:"updated_at"`
	Meta             interface{}   `json:"meta"`
	QRCodeURL        string        `json:"qr_code_url,omitempty"`
	QRCodeBase64     string        `json:"qr_code_base64,omitempty"`
	Tags             []interface{} `json:"tags,omitempty"`
	Pixels           []interface{} `json:"pixels,omitempty"`
}

// ShortLinkCreateRequest is used to create a short link.
type ShortLinkCreateRequest struct {
	LongURL          string      `json:"long_url"`
	ShortID          *string     `json:"short_id,omitempty"`
	Domain           string      `json:"domain"`
	ExpireAtDatetime *string     `json:"expire_at_datetime,omitempty"`
	ExpireAtViews    *int        `json:"expire_at_views,omitempty"`
	Description      *string     `json:"description,omitempty"`
	PublicStats      *bool       `json:"public_stats,omitempty"`
	Password         *string     `json:"password,omitempty"`
	Tags             []int       `json:"tags,omitempty"`
	Pixels           []int       `json:"pixels,omitempty"`
	Meta             interface{} `json:"meta,omitempty"`
}

// ShortLinkUpdateRequest is used to update a short link.
type ShortLinkUpdateRequest struct {
	ShortURL         string      `json:"short_url"`
	ShortID          *string     `json:"short_id,omitempty"`
	LongURL          string      `json:"long_url"`
	ExpireAtDatetime *string     `json:"expire_at_datetime,omitempty"`
	ExpireAtViews    *int        `json:"expire_at_views,omitempty"`
	Description      *string     `json:"description,omitempty"`
	PublicStats      *bool       `json:"public_stats,omitempty"`
	Password         *string     `json:"password,omitempty"`
	Tags             []int       `json:"tags,omitempty"`
	Pixels           []int       `json:"pixels,omitempty"`
	Meta             interface{} `json:"meta,omitempty"`
}

// CreateShortLink creates a new short link.
func (c *Client) CreateShortLink(reqData ShortLinkCreateRequest) (*ShortLink, error) {
	var link ShortLink
	err := c.doRequest(http.MethodPost, "/api/v1/link/shorten", nil, reqData, &link)
	if err != nil {
		return nil, err
	}
	return &link, nil
}

// GetShortLink retrieves a short link using its URL.
func (c *Client) GetShortLink(shortURL string) (*ShortLink, error) {
	query := url.Values{}
	query.Set("short_url", shortURL)
	var link ShortLink
	err := c.doRequest(http.MethodGet, "/api/v1/link", query, nil, &link)
	if err != nil {
		return nil, err
	}
	return &link, nil
}

// UpdateShortLink updates an existing short link.
func (c *Client) UpdateShortLink(reqData ShortLinkUpdateRequest) (*ShortLink, error) {
	var link ShortLink
	err := c.doRequest(http.MethodPut, "/api/v1/link", nil, reqData, &link)
	if err != nil {
		return nil, err
	}
	return &link, nil
}

// DeleteShortLink deletes a short link.
func (c *Client) DeleteShortLink(shortURL string) error {
	reqBody := map[string]string{
		"short_url": shortURL,
	}
	return c.doRequest(http.MethodDelete, "/api/v1/link", nil, reqBody, nil)
}

// ExpandRequest is used to expand a short link.
type ExpandRequest struct {
	ShortURL string  `json:"short_url"`
	Password *string `json:"password,omitempty"`
}

// ExpandResponse represents the response when expanding a short link.
type ExpandResponse struct {
	LongURL string `json:"long_url"`
	Expired bool   `json:"expired"`
}

// ExpandShortLink expands a short URL to its original long URL.
func (c *Client) ExpandShortLink(reqData ExpandRequest) (*ExpandResponse, error) {
	var resp ExpandResponse
	err := c.doRequest(http.MethodPost, "/api/v1/link/expand", nil, reqData, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListShortLinksOptions includes optional filters for list endpoint.
type ListShortLinksOptions struct {
	Search    string
	TagIDs    []int
	PixelIDs  []int
	StartDate string
	EndDate   string
	Domains   []int
	Page      int
}

// ShortLinkListResponse is the paginated response for listing short links.
type ShortLinkListResponse struct {
	CurrentPage int         `json:"current_page"`
	Data        []ShortLink `json:"data"`
	LastPage    int         `json:"last_page,omitempty"`
	PerPage     int         `json:"per_page,omitempty"`
	Total       int         `json:"total,omitempty"`
}

// ListShortLinksDetailed retrieves short links with typed filter options.
func (c *Client) ListShortLinksDetailed(options ListShortLinksOptions) (*ShortLinkListResponse, error) {
	query := url.Values{}
	if options.Search != "" {
		query.Set("search", options.Search)
	}
	addIndexedIntSlice(query, "tag_ids", options.TagIDs)
	addIndexedIntSlice(query, "pixel_ids", options.PixelIDs)
	if options.StartDate != "" {
		query.Set("start_date", options.StartDate)
	}
	if options.EndDate != "" {
		query.Set("end_date", options.EndDate)
	}
	addIndexedIntSlice(query, "domains", options.Domains)
	if options.Page > 0 {
		query.Set("page", strconv.Itoa(options.Page))
	}

	var result ShortLinkListResponse
	err := c.doRequest(http.MethodGet, "/api/v1/link/list", query, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ListShortLinks retrieves a list of short links using optional query parameters.
// The returned string is the raw JSON payload.
func (c *Client) ListShortLinks(queryParams map[string]string) (string, error) {
	var raw []byte
	err := c.doRequest(http.MethodGet, "/api/v1/link/list", queryFromMap(queryParams), nil, &raw)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

// BulkShortenLink represents one entry in a bulk shorten request.
type BulkShortenLink struct {
	LongURL     string  `json:"long_url"`
	Backhalf    *string `json:"backhalf,omitempty"`
	Password    *string `json:"password,omitempty"`
	Description *string `json:"description,omitempty"`
}

// BulkShortenRequest is used for bulk shortening of links.
// Links can be []BulkShortenLink, []string, or the raw format accepted by the API.
type BulkShortenRequest struct {
	Domain string      `json:"domain"`
	Links  interface{} `json:"links"`
	Tags   []int       `json:"tags,omitempty"`
	Pixels []int       `json:"pixels,omitempty"`
}

// BulkShortenLinks sends a bulk shorten request and returns the raw API payload.
func (c *Client) BulkShortenLinks(reqData BulkShortenRequest) (string, error) {
	var raw []byte
	err := c.doRequest(http.MethodPost, "/api/v1/link/bulk", nil, reqData, &raw)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

// BulkUpdateLink represents one entry in a bulk update request.
type BulkUpdateLink struct {
	ShortURL    string  `json:"short_url"`
	LongURL     string  `json:"long_url,omitempty"`
	Backhalf    *string `json:"backhalf,omitempty"`
	Password    *string `json:"password,omitempty"`
	Description *string `json:"description,omitempty"`
}

// BulkUpdateRequest is used for bulk updating links.
// Links can be []BulkUpdateLink or the raw format accepted by the API.
type BulkUpdateRequest struct {
	Links  interface{} `json:"links"`
	Tags   []int       `json:"tags,omitempty"`
	Pixels []int       `json:"pixels,omitempty"`
}

// BulkUpdateLinks updates multiple short links and returns the raw API payload.
func (c *Client) BulkUpdateLinks(reqData BulkUpdateRequest) (string, error) {
	var raw []byte
	err := c.doRequest(http.MethodPost, "/api/v1/link/bulk/update", nil, reqData, &raw)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

// =====================
// Stats Management
// =====================

// Stats represents link stats.
type Stats struct {
	Clicks       int                      `json:"clicks"`
	UniqueClicks int                      `json:"unique_clicks"`
	TotalQRScans int                      `json:"total_qr_scans,omitempty"`
	Browsers     []map[string]interface{} `json:"browsers"`
	Countries    []map[string]interface{} `json:"countries"`
	Cities       []map[string]interface{} `json:"cities,omitempty"`
	Referrers    []map[string]interface{} `json:"referrers"`
	Platforms    []map[string]interface{} `json:"platforms"`
	DailyClicks  []map[string]interface{} `json:"daily_clicks"`
	LinkClicks   []map[string]interface{} `json:"link_clicks,omitempty"`
	Data         map[string]interface{}   `json:"data"`
}

// StatsRequest includes parameters for the stats endpoints.
type StatsRequest struct {
	ShortURL  string
	StartDate string
	EndDate   string
}

// GetStats retrieves statistics for a short link.
func (c *Client) GetStats(shortURL string) (*Stats, error) {
	return c.GetStatsWithRange(StatsRequest{
		ShortURL: shortURL,
	})
}

// GetStatsWithRange retrieves statistics for a short link with an optional date range.
func (c *Client) GetStatsWithRange(reqData StatsRequest) (*Stats, error) {
	query := url.Values{}
	query.Set("short_url", reqData.ShortURL)
	if reqData.StartDate != "" {
		query.Set("start_date", reqData.StartDate)
	}
	if reqData.EndDate != "" {
		query.Set("end_date", reqData.EndDate)
	}

	var stats Stats
	err := c.doRequest(http.MethodGet, "/api/v1/link/stats", query, nil, &stats)
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

// =====================
// OneLink Management
// =====================

// OneLinkStatsRequest includes parameters for OneLink stats.
type OneLinkStatsRequest struct {
	ShortURL  string
	StartDate string
	EndDate   string
}

// OneLinkStats represents OneLink statistics.
type OneLinkStats struct {
	Clicks       int                      `json:"clicks"`
	UniqueClicks int                      `json:"unique_clicks"`
	TotalQRScans int                      `json:"total_qr_scans"`
	Browsers     []map[string]interface{} `json:"browsers"`
	Countries    []map[string]interface{} `json:"countries"`
	Cities       []map[string]interface{} `json:"cities"`
	Referrers    []map[string]interface{} `json:"referrers"`
	Platforms    []map[string]interface{} `json:"platforms"`
	DailyClicks  []map[string]interface{} `json:"daily_clicks"`
	LinkClicks   []map[string]interface{} `json:"link_clicks"`
	Data         map[string]interface{}   `json:"data"`
}

// GetOneLinkStats retrieves OneLink stats with optional date range.
func (c *Client) GetOneLinkStats(reqData OneLinkStatsRequest) (*OneLinkStats, error) {
	query := url.Values{}
	query.Set("short_url", reqData.ShortURL)
	if reqData.StartDate != "" {
		query.Set("start_date", reqData.StartDate)
	}
	if reqData.EndDate != "" {
		query.Set("end_date", reqData.EndDate)
	}

	var stats OneLinkStats
	err := c.doRequest(http.MethodGet, "/api/v1/onelink/stats", query, nil, &stats)
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

// DeleteOneLinkStats deletes OneLink stats for a short URL.
func (c *Client) DeleteOneLinkStats(shortURL string) error {
	reqBody := map[string]string{
		"short_url": shortURL,
	}
	return c.doRequest(http.MethodDelete, "/api/v1/onelink/stat", nil, reqBody, nil)
}

// OneLink represents a OneLink item.
type OneLink struct {
	ID          int         `json:"id"`
	ShortID     string      `json:"short_id"`
	ShortURL    string      `json:"short_url"`
	Domain      string      `json:"domain"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	AvatarURL   string      `json:"avatar_url"`
	Meta        interface{} `json:"meta"`
	CreatedAt   string      `json:"created_at"`
	UpdatedAt   string      `json:"updated_at"`
	LastClicked string      `json:"last_clicked,omitempty"`
}

// OneLinkListResponse is a paginated OneLink response.
type OneLinkListResponse struct {
	CurrentPage int       `json:"current_page"`
	Data        []OneLink `json:"data"`
	LastPage    int       `json:"last_page,omitempty"`
	PerPage     int       `json:"per_page,omitempty"`
	Total       int       `json:"total,omitempty"`
}

// ListOneLinks retrieves paginated OneLink records.
func (c *Client) ListOneLinks(page int) (*OneLinkListResponse, error) {
	query := url.Values{}
	if page > 0 {
		query.Set("page", strconv.Itoa(page))
	}

	var result OneLinkListResponse
	err := c.doRequest(http.MethodGet, "/api/v1/onelink/list", query, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// =====================
// UTM Preset Management
// =====================

// UTMPreset represents a UTM preset object.
type UTMPreset struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Source    string `json:"source"`
	Medium    string `json:"medium"`
	Campaign  string `json:"campaign"`
	Content   string `json:"content"`
	Term      string `json:"term"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// UTMPresetRequest is used to create/update a UTM preset.
type UTMPresetRequest struct {
	Name     string `json:"name"`
	Source   string `json:"source"`
	Medium   string `json:"medium"`
	Campaign string `json:"campaign"`
	Content  string `json:"content"`
	Term     string `json:"term"`
}

func decodeUTMPreset(data []byte) (*UTMPreset, error) {
	var preset UTMPreset
	if err := json.Unmarshal(data, &preset); err == nil {
		return &preset, nil
	}

	var wrapped struct {
		Data UTMPreset `json:"data"`
	}
	if err := json.Unmarshal(data, &wrapped); err == nil {
		return &wrapped.Data, nil
	}
	return nil, fmt.Errorf("unable to decode UTM preset response")
}

// CreateUTMPreset creates a UTM preset.
func (c *Client) CreateUTMPreset(reqData UTMPresetRequest) (*UTMPreset, error) {
	data, err := c.doRequestRaw(http.MethodPost, "/api/v1/link/utm-preset", nil, reqData)
	if err != nil {
		return nil, err
	}
	return decodeUTMPreset(data)
}

// ListUTMPresets retrieves all UTM presets.
func (c *Client) ListUTMPresets() ([]UTMPreset, error) {
	data, err := c.doRequestRaw(http.MethodGet, "/api/v1/link/utm-preset", nil, nil)
	if err != nil {
		return nil, err
	}

	var presets []UTMPreset
	if err := json.Unmarshal(data, &presets); err == nil {
		return presets, nil
	}

	var wrapped struct {
		Data []UTMPreset `json:"data"`
	}
	if err := json.Unmarshal(data, &wrapped); err == nil {
		return wrapped.Data, nil
	}
	return nil, fmt.Errorf("unable to decode UTM preset list response")
}

// GetUTMPreset retrieves a UTM preset by ID.
func (c *Client) GetUTMPreset(id int) (*UTMPreset, error) {
	path := fmt.Sprintf("/api/v1/link/utm-preset/%d", id)
	data, err := c.doRequestRaw(http.MethodGet, path, nil, nil)
	if err != nil {
		return nil, err
	}
	return decodeUTMPreset(data)
}

// UpdateUTMPreset updates a UTM preset by ID.
func (c *Client) UpdateUTMPreset(id int, reqData UTMPresetRequest) (*UTMPreset, error) {
	path := fmt.Sprintf("/api/v1/link/utm-preset/%d", id)
	data, err := c.doRequestRaw(http.MethodPut, path, nil, reqData)
	if err != nil {
		return nil, err
	}
	return decodeUTMPreset(data)
}

// DeleteUTMPreset deletes a UTM preset by ID.
func (c *Client) DeleteUTMPreset(id int) error {
	path := fmt.Sprintf("/api/v1/link/utm-preset/%d", id)
	return c.doRequest(http.MethodDelete, path, nil, nil, nil)
}

// =====================
// QR Code Management
// =====================

// QRCodeRequest includes query options for retrieving a QR code.
type QRCodeRequest struct {
	ShortURL string
	Output   string
	Format   string
}

// GetQRCode retrieves QR code bytes (image or raw payload based on output parameter).
func (c *Client) GetQRCode(reqData QRCodeRequest) ([]byte, error) {
	query := url.Values{}
	query.Set("short_url", reqData.ShortURL)
	if reqData.Output != "" {
		query.Set("output", reqData.Output)
	}
	if reqData.Format != "" {
		query.Set("format", reqData.Format)
	}
	return c.doRequestRaw(http.MethodGet, "/api/v1/link/qr-code", query, nil)
}

// QRCodeUpdateRequest includes QR code customization options.
type QRCodeUpdateRequest struct {
	ShortURL        string  `json:"short_url"`
	Image           *string `json:"image,omitempty"`
	BackgroundColor *string `json:"background_color,omitempty"`
	CornerDotsColor *string `json:"corner_dots_color,omitempty"`
	DotsColor       *string `json:"dots_color,omitempty"`
	DotsStyle       *string `json:"dots_style,omitempty"`
	CornerStyle     *string `json:"corner_style,omitempty"`
}

// QRCode represents a QR code record.
type QRCode struct {
	ID            int                    `json:"id"`
	ShortURL      string                 `json:"short_url"`
	QRCodeOptions map[string]interface{} `json:"qr_code_options"`
	TeamID        int                    `json:"team_id"`
	UserID        int                    `json:"user_id"`
	UpdatedAt     string                 `json:"updated_at"`
}

// UpdateQRCode updates QR code options for a short link.
func (c *Client) UpdateQRCode(reqData QRCodeUpdateRequest) (*QRCode, error) {
	var qrCode QRCode
	err := c.doRequest(http.MethodPut, "/api/v1/link/qr-code", nil, reqData, &qrCode)
	if err != nil {
		return nil, err
	}
	return &qrCode, nil
}

// =====================
// Tag Management
// =====================

// Tag represents a tag.
type Tag struct {
	ID        int    `json:"id"`
	Tag       string `json:"tag"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// ListTags retrieves all tags.
func (c *Client) ListTags() ([]Tag, error) {
	var tags []Tag
	err := c.doRequest(http.MethodGet, "/api/v1/link/tag", nil, nil, &tags)
	if err != nil {
		return nil, err
	}
	return tags, nil
}

// CreateTag creates a new tag.
func (c *Client) CreateTag(tagValue string) (*Tag, error) {
	reqBody := map[string]string{
		"tag": tagValue,
	}
	var tag Tag
	err := c.doRequest(http.MethodPost, "/api/v1/link/tag", nil, reqBody, &tag)
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

// GetTag retrieves a tag by its ID.
func (c *Client) GetTag(id int) (*Tag, error) {
	path := fmt.Sprintf("/api/v1/link/tag/%d", id)
	var tag Tag
	err := c.doRequest(http.MethodGet, path, nil, nil, &tag)
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

// UpdateTag updates an existing tag.
func (c *Client) UpdateTag(id int, tagValue string) (*Tag, error) {
	path := fmt.Sprintf("/api/v1/link/tag/%d", id)
	reqBody := map[string]string{
		"tag": tagValue,
	}
	var tag Tag
	err := c.doRequest(http.MethodPut, path, nil, reqBody, &tag)
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

// DeleteTag deletes a tag by its ID.
func (c *Client) DeleteTag(id int) error {
	path := fmt.Sprintf("/api/v1/link/tag/%d", id)
	return c.doRequest(http.MethodDelete, path, nil, nil, nil)
}
