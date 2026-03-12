package types

type SessionVO struct {
    Uuid          string `json:"uuid"`
    TargetId      string `json:"target_id"`
    TargetName    string `json:"target_name"`
    TargetAvatar  string `json:"target_avatar"`
    SessionType   int8   `json:"session_type"`
    UnreadCount   int    `json:"unread_count"`
    LastMessage   string `json:"last_message"`
    LastMessageAt string `json:"last_message_at"`
    IsTop         int8   `json:"is_top"`
}

type GetSessionListReq struct {
    Page int `form:"page" binding:"required,min=1"`
    Size int `form:"size" binding:"required,min=1,max=100"`
}

type GetSessionListResp struct {
    Total int64       `json:"total"`
    List  []SessionVO `json:"list"`
}

type CreateSessionReq struct {
    TargetId    string `json:"target_id" binding:"required"`
    SessionType int8   `json:"session_type" binding:"oneof=0 1"`
}

type CreateSessionResp struct {
    Uuid string `json:"uuid"`
}

type PinSessionReq struct {
    IsTop *int8 `json:"is_top" binding:"required,oneof=0 1"`
}
