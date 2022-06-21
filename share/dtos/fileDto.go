package dtos

type FileDto struct {
	Id         uint   `json:"id"`
	UniqueName string `json:"uniqueName" gorm:"unique"`
	Name       string `json:"name"`
	//CreationDate time.Time `json:"creationDate"`
	OwnerId uint `json:"ownerId"`
}
