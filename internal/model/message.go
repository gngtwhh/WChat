package model

import (
    "time"

    "gorm.io/gorm"
)

type Message struct {
    gorm.Model
    Uuid      string `gorm:"uniqueIndex;type:char(20);not null;comment:消息唯一id"`
    SessionId string `gorm:"index;type:char(20);not null;comment:会话id"`
    Type      int8   `gorm:"type:tinyint;not null;comment:消息类型，0:文本, 1:语音, 2:文件, 3:WebRTC通话"`
    Content   string `gorm:"type:text;comment:文本消息内容"`
    Url       string `gorm:"type:varchar(255);comment:媒体文件的地址"`
    SendId    string `gorm:"index;type:char(20);not null;comment:发送者id"`
    ReceiveId string `gorm:"index;type:char(20);not null;comment:接收者id(群聊时可为空或存群组id)"`
    // 文件元数据
    FileType string `gorm:"type:varchar(20);comment:文件扩展名或MIME类型(如 image/png)"`
    FileName string `gorm:"type:varchar(100);comment:文件原始名称"`
    FileSize int64  `gorm:"comment:文件大小(单位:字节)"`
    // WebRTC 信令数据 (SDP, ICE 等)
    AVdata string     `gorm:"type:text;comment:WebRTC通话传递的信令数据"`
    Status int8       `gorm:"type:tinyint;not null;default:0;comment:状态，0:未发送/发送中, 1:已发送, 2:已撤回"`
    SendAt *time.Time `gorm:"comment:发送时间"`
}
