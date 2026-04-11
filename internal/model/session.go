package model

import (
	"time"

	"gorm.io/gorm"
)

type Session struct {
	gorm.Model
	Uuid        string `gorm:"uniqueIndex;type:char(20);not null;comment:会话唯一标识" json:"uuid"`
	UserId      string `gorm:"uniqueIndex:uk_user_target;type:char(20);not null;comment:该会话归属的用户id" json:"user_id"`
	TargetId    string `gorm:"uniqueIndex:uk_user_target;type:char(20);not null;comment:聊天对象id(单聊为好友id, 群聊为群id)" json:"target_id"`
	SessionType int8   `gorm:"uniqueIndex:uk_user_target;type:tinyint;not null;default:0;comment:会话类型, 0:单聊, 1:群聊" json:"session_type"`

	UnreadCount   int        `gorm:"type:int;not null;default:0;comment:未读消息数" json:"unread_count"`
	LastMessage   string     `gorm:"type:varchar(500);comment:最新一条消息的摘要(例如: '[图片]', '你在干嘛？')" json:"last_message"`
	LastMessageAt *time.Time `gorm:"index;comment:最新消息的接收/发送时间(用于列表排序)" json:"last_message_at"`

	IsTop int8 `gorm:"type:tinyint;not null;default:0;comment:是否置顶, 0:否, 1:是" json:"is_top"`
}
