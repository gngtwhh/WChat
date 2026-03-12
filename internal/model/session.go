package model

import (
    "time"

    "gorm.io/gorm"
)

type Session struct {
    gorm.Model
    Uuid        string `gorm:"uniqueIndex;type:char(20);not null;comment:会话唯一标识"`
    UserId      string `gorm:"index:idx_user_session;type:char(20);not null;comment:该会话归属的用户id"`
    TargetId    string `gorm:"index:idx_user_session;type:char(20);not null;comment:聊天对象id(单聊为好友id, 群聊为群id)"`
    SessionType int8   `gorm:"type:tinyint;not null;default:0;comment:会话类型, 0:单聊, 1:群聊"`

    UnreadCount   int        `gorm:"type:int;not null;default:0;comment:未读消息数"`
    LastMessage   string     `gorm:"type:varchar(500);comment:最新一条消息的摘要(例如: '[图片]', '你在干嘛？')"`
    LastMessageAt *time.Time `gorm:"index;comment:最新消息的接收/发送时间(用于列表排序)"`

    IsTop int8 `gorm:"type:tinyint;not null;default:0;comment:是否置顶, 0:否, 1:是"`
}
