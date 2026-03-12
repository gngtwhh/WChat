package types

type ApplicationVO struct {
    Uuid        string `json:"uuid"`
    UserId      string `json:"user_id"`       // 申请人的 UUID
    Nickname    string `json:"nickname"`      // 申请人的昵称 (需要连表查 User)
    Avatar      string `json:"avatar"`        // 申请人的头像
    ContactId   string `json:"contact_id"`    // 目标ID (用户ID 或 群ID)
    ContactType int8   `json:"contact_type"`  // 0.用户，1.群聊
    Status      int8   `json:"status"`        // 0.待处理, 1.已通过, 2.已拒绝, 3.拉黑
    Message     string `json:"message"`       // 留言信息
    LastApplyAt string `json:"last_apply_at"` // 最后申请时间
}

type GetApplicationListResp struct {
    Total int             `json:"total"`
    List  []ApplicationVO `json:"list"`
}

type SubmitApplicationReq struct {
    ContactId   string `json:"contact_id" binding:"required"`
    ContactType int8   `json:"contact_type" binding:"oneof=0 1"`
    Message     string `json:"message" binding:"omitempty,max=100"`
}

type HandleApplicationReq struct {
    Status int8 `json:"status" binding:"required,oneof=1 2 3"` // 1:同意 2:拒绝 3:拉黑
}
