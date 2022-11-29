package dto

type ImageList struct {
	Total int64
	List  []*ImageInfo
}

type ImageInfo struct {
	Repo    string `json:"repository"`
	ImageId string `json:"image_id"`
	Created string `json:"created"`
	Size    string `json:"size"`
}
