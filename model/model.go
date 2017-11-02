// Package model - contains all model info
package model

// ImageInfo - image info structure for (un)marshalling json
type ImageInfo struct {
	ID     string                 `json:"id"`
	URL    string                 `json:"arcurl"`
	Click  string                 `json:"click"`
	PicNum int                    `json:"picnum"`
	Stow   string                 `json:"stow"`
	X      map[string]interface{} `json:"-"`
}

// ImageCollection - image info list
type ImageCollection struct {
	Status int          `json:"statu"`
	List   []*ImageInfo `json:"list"`
	Total  string       `json:"total"`
	Pages  int          `json:"pages"`
}
