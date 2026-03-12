package types

type RegisterReq struct {
    // 根据你的表结构，使用 Telephone 作为主账号
    Telephone       string `json:"telephone" binding:"required,len=11"`
    Password        string `json:"password" binding:"required,min=6,max=20"`
    ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=Password"`
    Nickname        string `json:"nickname" binding:"required,max=20"`
}

type LoginReq struct {
    Telephone string `json:"telephone" binding:"required,len=11"`
    Password  string `json:"password" binding:"required"`
}

type LoginResp struct {
    Token    string `json:"token"`
    UserInfo UserVO `json:"user_info"`
}

// UserVO 脱敏后的用户信息展示对象 (View Object)
type UserVO struct {
    Uuid      string `json:"uuid"`
    Nickname  string `json:"nickname"`
    Avatar    string `json:"avatar"`
    Signature string `json:"signature"`
}
