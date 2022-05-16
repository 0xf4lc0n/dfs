package models

type User struct {
	Id            uint   `json:"id"`
	Name          string `json:"name"`
	Email         string `json:"email" gorm:"unique"`
	Password      []byte `json:"-"`
	Verified      bool   `json:"-"`
	HomeDirectory string `json:"directory"`
	CryptKey      string `json:"cryptKey"`
	OwnedFiles    []File `json:"ownedFiles" gorm:"foreignKey:OwnerId;references:Id;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}
