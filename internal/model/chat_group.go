package model

import (
    "encoding/json"

    "gorm.io/gorm"
)

type Group struct {
    gorm.Model
    Uuid      string          `gorm:"column:uuid;uniqueIndex;type:char(20);not null;comment:群组唯一id"`
    Name      string          `gorm:"column:name;type:varchar(20);not null;comment:群名称"`
    Notice    string          `gorm:"column:notice;type:varchar(500);comment:群公告"`
    Members   json.RawMessage `gorm:"column:members;type:json;comment:群组成员"`
    MemberCnt int             `gorm:"column:member_cnt;default:1;comment:群人数"` // 默认群主1人
    OwnerId   string          `gorm:"column:owner_id;type:char(20);not null;comment:群主uuid"`
    AddMode   int8            `gorm:"column:add_mode;default:0;comment:加群方式，0.直接，1.审核"`
    Avatar    string          `gorm:"column:avatar;type:char(255);default:https://api.dicebear.com/9.x/pixel-art/svg?seed=12345;not null;comment:头像"`
    Status    int8            `gorm:"column:status;default:0;comment:状态，0.正常，1.禁用，2.解散"`
}
