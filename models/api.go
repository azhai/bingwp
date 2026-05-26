package models

// WallpaperRaw represents the raw data from the API
type WallpaperRaw struct {
	GUID        string `json:"guid"`
	Date        string `json:"date"`
	Title       string `json:"title"`
	Copyright   string `json:"copyright"`
	Headline    string `json:"headline"`
	Filepath    string `json:"filepath"`
	Description string `json:"description"`
}

// WiliiListResponse represents the response from the list API
type WiliiListResponse struct {
	Status   int              `json:"status"`
	Success  bool             `json:"success"`
	Msg      string           `json:"msg"`
	Response ListResponseData `json:"response"`
}

// ListResponseData represents the data section of the list API response
type ListResponseData struct {
	Page      int            `json:"page"`
	PageCount int            `json:"pageCount"`
	DataCount int            `json:"dataCount"`
	PageSize  int            `json:"pageSize"`
	Data      []WallpaperRaw `json:"data"`
}

// WiliiDetailResponse represents the response from the detail API
type WiliiDetailResponse struct {
	Status   int        `json:"status"`
	Success  bool       `json:"success"`
	Msg      string     `json:"msg"`
	Response DetailData `json:"response"`
}

// DetailData represents the data section of the detail API response
type DetailData struct {
	GUID        string `json:"guid"`
	Date        string `json:"date"`
	Title       string `json:"title"`
	Copyright   string `json:"copyright"`
	Headline    string `json:"headline"`
	Description string `json:"description"`
	Filepath    string `json:"filepath"`
}
