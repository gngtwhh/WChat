package types

type GroupVO struct {
    Uuid      string `json:"uuid"`
    Name      string `json:"name"`
    Notice    string `json:"notice"`
    Avatar    string `json:"avatar"`
    MemberCnt int    `json:"member_cnt"`
    OwnerId   string `json:"owner_id"`
    AddMode   int8   `json:"add_mode"`
    Status    int8   `json:"status"`
}

type GroupMemberVO struct {
    Uuid      string `json:"uuid"`
    Nickname  string `json:"nickname"`
    Avatar    string `json:"avatar"`
    Signature string `json:"signature"`
}

type CreateGroupReq struct {
    Name    string   `json:"name" binding:"required,max=20"`
    Notice  string   `json:"notice" binding:"omitempty,max=500"`
    Avatar  string   `json:"avatar" binding:"omitempty,url"`
    AddMode *int8    `json:"add_mode" binding:"omitempty,oneof=0 1"`
    Members []string `json:"members" binding:"omitempty,dive,required"`
}

type UpdateGroupReq struct {
    Name    string `json:"name" binding:"omitempty,max=20"`
    Notice  string `json:"notice" binding:"omitempty,max=500"`
    Avatar  string `json:"avatar" binding:"omitempty,url"`
    AddMode *int8  `json:"add_mode" binding:"omitempty,oneof=0 1"`
}

type InviteMemberReq struct {
    UserIds []string `json:"user_ids" binding:"required,min=1"`
}

type CreateGroupResp struct {
    Uuid string `json:"uuid"`
}

type GetJoinedGroupsResp struct {
    Total int       `json:"total"`
    List  []GroupVO `json:"list"`
}

type GetGroupMembersResp struct {
    Total int             `json:"total"`
    List  []GroupMemberVO `json:"list"`
}
