package model

import (
	"gorm.io/gorm"
)

const (
	ContactTypeUser  int8 = 0
	ContactTypeGroup int8 = 1
)

const (
	ContactStatusNormal         int8 = 0
	ContactStatusBlocked        int8 = 1
	ContactStatusBlockedByPeer  int8 = 2
	ContactStatusDeleted        int8 = 3
	ContactStatusDeletedByPeer  int8 = 4
	ContactStatusMuted          int8 = 5
	ContactStatusLeftGroup      int8 = 6
	ContactStatusRemovedFromGrp int8 = 7
)

type Contact struct {
	gorm.Model
	Id          int64  `gorm:"column:id;primaryKey;comment:自增id" json:"id"`
	UserId      string `gorm:"column:user_id;index;type:char(20);not null;comment:用户唯一id" json:"user_id"`
	ContactId   string `gorm:"column:contact_id;index;type:char(20);not null;comment:联系人id" json:"contact_id"`
	ContactType int8   `gorm:"column:contact_type;not null;comment:联系类型，0.用户，1.群聊" json:"contact_type"`
	Status      int8   `gorm:"column:status;not null;comment:联系状态，0.正常，1.拉黑，2.被拉黑，3.删除好友，4.被删除好友，5.被禁言，6.退出群聊，7.被踢出群聊" json:"status"`
}
