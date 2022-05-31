package models

type Role uint

const (
	Member    Role = iota
	Moderator Role = iota
	Owner     Role = iota
)

type ShareSpaceMember struct {
	ShareSpaceId uint `json:"shareSpaceId" gorm:"primaryKey;autoIncrement:false"`
	UserId       uint `json:"userId" gorm:"primaryKey;autoIncrement:false"`
	Role         Role `json:"role"`
}
