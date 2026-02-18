# T.LY Go Client

Go client for the [T.LY API](https://api.t.ly), aligned with the provided Postman collection (`collection.json`).

## Installation

```bash
go get github.com/timleland/t.ly-go-url-shortener-api
```

```go
import tly "github.com/timleland/t.ly-go-url-shortener-api"
```

## Quick Start

```go
client := tly.NewClient("YOUR_API_TOKEN")
```

## Method Index

### Short Links

- `CreateShortLink(reqData ShortLinkCreateRequest) (*ShortLink, error)`
- `GetShortLink(shortURL string) (*ShortLink, error)`
- `UpdateShortLink(reqData ShortLinkUpdateRequest) (*ShortLink, error)`
- `DeleteShortLink(shortURL string) error`
- `ExpandShortLink(reqData ExpandRequest) (*ExpandResponse, error)`
- `ListShortLinksDetailed(options ListShortLinksOptions) (*ShortLinkListResponse, error)`
- `ListShortLinks(queryParams map[string]string) (string, error)` raw JSON payload
- `BulkShortenLinks(reqData BulkShortenRequest) (string, error)` raw payload
- `BulkUpdateLinks(reqData BulkUpdateRequest) (string, error)` raw payload
- `GetStats(shortURL string) (*Stats, error)`
- `GetStatsWithRange(reqData StatsRequest) (*Stats, error)`

### OneLink

- `GetOneLinkStats(reqData OneLinkStatsRequest) (*OneLinkStats, error)`
- `DeleteOneLinkStats(shortURL string) error`
- `ListOneLinks(page int) (*OneLinkListResponse, error)`

### UTM Presets

- `CreateUTMPreset(reqData UTMPresetRequest) (*UTMPreset, error)`
- `ListUTMPresets() ([]UTMPreset, error)`
- `GetUTMPreset(id int) (*UTMPreset, error)`
- `UpdateUTMPreset(id int, reqData UTMPresetRequest) (*UTMPreset, error)`
- `DeleteUTMPreset(id int) error`

### QR Codes

- `GetQRCode(reqData QRCodeRequest) ([]byte, error)` raw bytes payload
- `UpdateQRCode(reqData QRCodeUpdateRequest) (*QRCode, error)`

### Pixels

- `CreatePixel(reqData PixelCreateRequest) (*Pixel, error)`
- `ListPixels() ([]Pixel, error)`
- `GetPixel(id int) (*Pixel, error)`
- `UpdatePixel(reqData PixelUpdateRequest) (*Pixel, error)`
- `DeletePixel(id int) error`

### Tags

- `ListTags() ([]Tag, error)`
- `CreateTag(tagValue string) (*Tag, error)`
- `GetTag(id int) (*Tag, error)`
- `UpdateTag(id int, tagValue string) (*Tag, error)`
- `DeleteTag(id int) error`

## Examples

### List Links with Filters

```go
links, err := client.ListShortLinksDetailed(tly.ListShortLinksOptions{
	Search:   "amazon",
	TagIDs:   []int{1, 2, 3},
	PixelIDs: []int{1, 2, 3},
	StartDate: "2035-01-17 15:00:00",
	EndDate:   "2037-01-17 15:00:00",
	Domains:   []int{1, 2, 3},
	Page:      1,
})
if err != nil {
	panic(err)
}
_ = links
```

### Get Stats with Date Range

```go
stats, err := client.GetStatsWithRange(tly.StatsRequest{
	ShortURL:  "https://t.ly/OYXL",
	StartDate: "2025-08-01T00:00:00Z",
	EndDate:   "2025-08-31T23:59:59Z",
})
if err != nil {
	panic(err)
}
_ = stats
```

### OneLink Stats

```go
stats, err := client.GetOneLinkStats(tly.OneLinkStatsRequest{
	ShortURL:  "https://t.ly/one",
	StartDate: "2024-06-01",
	EndDate:   "2024-06-08",
})
if err != nil {
	panic(err)
}
_ = stats
```

### UTM Preset CRUD

```go
preset, err := client.CreateUTMPreset(tly.UTMPresetRequest{
	Name:     "Newsletter Launch",
	Source:   "newsletter",
	Medium:   "email",
	Campaign: "fall_launch",
	Content:  "hero-cta",
	Term:     "running-shoes",
})
if err != nil {
	panic(err)
}

updated, err := client.UpdateUTMPreset(preset.ID, tly.UTMPresetRequest{
	Name:     "Newsletter Launch",
	Source:   "newsletter",
	Medium:   "email",
	Campaign: "fall_launch_v2",
	Content:  "hero-cta",
	Term:     "running-shoes",
})
if err != nil {
	panic(err)
}

_ = updated
```

### Fetch QR Code Bytes

```go
qrPayload, err := client.GetQRCode(tly.QRCodeRequest{
	ShortURL: "https://t.ly/c55j",
	Output:   "base64", // or omit for image bytes depending on API behavior
	Format:   "eps",
})
if err != nil {
	panic(err)
}

qrAsString := string(qrPayload)
_ = qrAsString
```

## Notes

- Non-2xx API responses return `*APIError` with status code and raw response body.
- Raw-response methods are `GetQRCode`, `ListShortLinks`, `BulkShortenLinks`, and `BulkUpdateLinks`.
- Raw-response methods return the API payload unchanged so callers can parse endpoint-specific formats.
- You can override base URL if needed:

```go
client.BaseURL = "https://api.t.ly"
```

## License

MIT
