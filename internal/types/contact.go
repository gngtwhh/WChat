package types

type ContactVO struct {
    ContactId string `json:"contact_id"`
    Nickname  string `json:"nickname"`
    Avatar    string `json:"avatar"`
    Signature string `json:"signature"`
    Status    int8   `json:"status"` // 0.正常, 1.拉黑 等等
}

type GetContactListResp struct {
    Total int         `json:"total"`
    List  []ContactVO `json:"list"`
}
