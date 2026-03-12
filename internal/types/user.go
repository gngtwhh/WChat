package types

import "wchat/internal/model"

type UserProfileResp struct {
    Uuid      string `json:"uuid"`
    Nickname  string `json:"nickname"`
    Telephone string `json:"telephone"`
    Email     string `json:"email"`
    Avatar    string `json:"avatar"`
    Gender    int8   `json:"gender"`
    Signature string `json:"signature"`
    Birthday  string `json:"birthday"`
}

func BuildUserProfileResp(user *model.User) UserProfileResp {
    return UserProfileResp{
        Uuid:      user.Uuid,
        Nickname:  user.Nickname,
        Telephone: user.Telephone,
        Email:     user.Email,
        Avatar:    user.Avatar,
        Gender:    user.Gender,
        Signature: user.Signature,
        Birthday:  user.Birthday,
    }
}

type UpdateProfileReq struct {
    Nickname  string `json:"nickname" binding:"omitempty,max=20"`
    Email     string `json:"email" binding:"omitempty,email"`
    Avatar    string `json:"avatar" binding:"omitempty,url"`
    Gender    *int8  `json:"gender" binding:"omitempty,oneof=0 1"`
    Signature string `json:"signature" binding:"omitempty,max=100"`
}

type GetUserListReq struct {
    Page    int    `form:"page" binding:"required,min=1"`
    Size    int    `form:"size" binding:"required,min=1,max=100"`
    Keyword string `form:"keyword" binding:"omitempty,max=20"`
}

type GetUserListResp struct {
    Total int64             `json:"total"`
    List  []UserProfileResp `json:"list"`
}

type SetUserStatusReq struct {
    Status *int8 `json:"status" binding:"required,oneof=0 1"` // 0.正常，1.禁用 (pointer to avoid default zero value)
}

type SetUserRoleReq struct {
    IsAdmin *int8 `json:"is_admin" binding:"required,oneof=0 1"`
}
