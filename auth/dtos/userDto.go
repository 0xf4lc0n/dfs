package dtos

type User struct {
	Id            uint   `json:"id"`
	Name          string `json:"name"`
	Email         string `json:"email" gorm:"unique"`
	Verified      bool   `json:"verified"`
	HomeDirectory string `json:"directory"`
	CryptKey      string `json:"cryptKey"`
}
