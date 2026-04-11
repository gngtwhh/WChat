package model

import (
	"time"

	"gorm.io/gorm"
)

type Message struct {
	gorm.Model
	Uuid           string `gorm:"uniqueIndex;type:char(20);not null;comment:消息唯一id" json:"uuid"`
	SessionId      string `gorm:"index;type:char(20);not null;comment:发送方会话id" json:"session_id"`
	ConversationId string `gorm:"index;type:varchar(64);not null;comment:稳定会话主体id(私聊按双方uuid排序, 群聊按群uuid)" json:"-"`
	Type           int8   `gorm:"type:tinyint;not null;comment:消息类型，0:文本, 1:语音, 2:文件, 3:WebRTC通话" json:"type"`
	Content        string `gorm:"type:text;comment:文本消息内容" json:"content"`
	Url            string `gorm:"type:varchar(255);comment:媒体文件的地址" json:"url"`
	SendId         string `gorm:"index;type:char(20);not null;comment:发送者id" json:"send_id"`
	ReceiveId      string `gorm:"index;type:char(20);not null;comment:接收者id(群聊时可为空或存群组id)" json:"receive_id"`
	// 文件元数据
	FileType string `gorm:"type:varchar(20);comment:文件扩展名或MIME类型(如 image/png)" json:"file_type"`
	FileName string `gorm:"type:varchar(100);comment:文件原始名称" json:"file_name"`
	FileSize int64  `gorm:"comment:文件大小(单位:字节)" json:"file_size"`
	// WebRTC 信令数据 (SDP, ICE 等)
	AVdata string     `gorm:"type:text;comment:WebRTC通话传递的信令数据" json:"av_data,omitempty"`
	Status int8       `gorm:"type:tinyint;not null;default:0;comment:状态，0:未发送/发送中, 1:已发送, 2:已撤回" json:"status"`
	SendAt *time.Time `gorm:"comment:发送时间" json:"send_at"`
}
