package types

type GetMessageListReq struct {
    SessionId string `form:"session_id" binding:"required"`
    Page      int    `form:"page" binding:"required,min=1"`
    Size      int    `form:"size" binding:"required,min=1,max=100"`
}

type MessageVO struct {
    Uuid      string `json:"uuid"`
    SessionId string `json:"session_id"`
    Type      int8   `json:"type"`
    Content   string `json:"content"`
    Url       string `json:"url"`
    SendId    string `json:"send_id"`
    ReceiveId string `json:"receive_id"`
    FileType  string `json:"file_type"`
    FileName  string `json:"file_name"`
    FileSize  int64  `json:"file_size"`
    AVdata    string `json:"av_data,omitempty"`
    Status    int8   `json:"status"`
    SendAt    string `json:"send_at"`
}

type GetMessageListResp struct {
    Total int64       `json:"total"`
    List  []MessageVO `json:"list"`
}
