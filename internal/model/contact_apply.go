package model

import (
	"time"

	"gorm.io/gorm"
)

type ContactApply struct {
	gorm.Model
	Uuid        string    `gorm:"column:uuid;uniqueIndex;type:char(20);comment:申请id" json:"uuid"`
	UserId      string    `gorm:"column:user_id;index;type:char(20);not null;comment:申请人id" json:"user_id"`
	ContactId   string    `gorm:"column:contact_id;index;type:char(20);not null;comment:被申请id" json:"contact_id"`
	ContactType int8      `gorm:"column:contact_type;not null;comment:被申请类型，0.用户，1.群聊" json:"contact_type"`
	Status      int8      `gorm:"column:status;not null;comment:申请状态，0.申请中，1.通过，2.拒绝，3.拉黑" json:"status"`
	Message     string    `gorm:"column:message;type:varchar(100);comment:申请信息" json:"message"`
	LastApplyAt time.Time `gorm:"column:last_apply_at;type:datetime;not null;comment:最后申请时间" json:"last_apply_at"`
}
