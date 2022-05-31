package models

type ShareSpace struct {
	Id            uint   `json:"id"`
	Name          string `json:"name"`
	Owner         uint   `json:"owner"`
	HomeDirectory string `json:"directory"`
	CryptKey      string `json:"cryptKey"`
}
