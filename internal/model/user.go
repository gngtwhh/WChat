package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Uuid              string     `gorm:"column:uuid;uniqueIndex;type:char(20);comment:用户唯一id" json:"uuid"`
	Nickname          string     `gorm:"column:nickname;type:varchar(20);not null;comment:昵称" json:"nickname"`
	Telephone         string     `gorm:"column:telephone;uniqueIndex;not null;type:varchar(20);comment:电话" json:"telephone"`
	Email             string     `gorm:"column:email;type:varchar(50);comment:邮箱" json:"email"`
	Avatar            string     `gorm:"column:avatar;type:varchar(255);default:https://api.dicebear.com/9.x/pixel-art/svg?seed=12345;not null;comment:头像" json:"avatar"`
	Gender            int8       `gorm:"column:gender;comment:性别，0.男，1.女" json:"gender"`
	Signature         string     `gorm:"column:signature;type:varchar(100);comment:个性签名" json:"signature"`
	Password          string     `gorm:"column:password;type:varchar(100);not null;comment:密码" json:"-"`
	Birthday          string     `gorm:"column:birthday;type:varchar(20);comment:生日" json:"birthday"`
	LastOnlineAt      *time.Time `gorm:"column:last_online_at;type:datetime;comment:上次登录时间" json:"last_online_at"`
	LastOfflineAt     *time.Time `gorm:"column:last_offline_at;type:datetime;comment:最近离线时间" json:"last_offline_at"`
	DeleteRequestedAt *time.Time `gorm:"column:delete_requested_at;type:datetime;comment:发起注销时间" json:"delete_requested_at"`
	DeleteAfter       *time.Time `gorm:"column:delete_after;type:datetime;comment:注销生效时间" json:"delete_after"`
	IsAdmin           int8       `gorm:"column:is_admin;not null;comment:是否是管理员，0.不是，1.是" json:"is_admin"`
	Status            int8       `gorm:"column:status;index;not null;comment:状态，0.正常，1.禁用" json:"status"`
}
