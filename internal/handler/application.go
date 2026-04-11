package handler

import (
	"github.com/gin-gonic/gin"

	"wchat/internal/middleware"
	"wchat/internal/service"
	"wchat/internal/types"
	"wchat/pkg/errcode"
	"wchat/pkg/response"
)

type ApplicationHandler struct {
	svc *service.ApplicationService
}

func NewApplicationHandler(svc *service.ApplicationService) *ApplicationHandler {
	return &ApplicationHandler{svc: svc}
}

// GetApplicationList 获取收到的申请列表
// @Summary      获取收到的申请列表
// @Description  获取当前用户收到的所有好友/群聊申请，包含申请人昵称、头像等信息
// @Tags         申请
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.Response{data=types.GetApplicationListResp}  "申请列表"
// @Failure      200  {object}  response.Response{data=nil}                            "Token 无效"
// @Router       /applications [get]
func (h *ApplicationHandler) GetApplicationList(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Fail(c, errcode.TokenInvalid)
		return
	}

	applies, err := h.svc.GetApplicationList(c.Request.Context(), userID)
	if err != nil {
		response.FailErr(c, err)
		return
	}

	voList := make([]types.ApplicationVO, 0, len(applies))
	for _, app := range applies {
		voList = append(
			voList, types.ApplicationVO{
				Uuid:        app.Uuid,
				UserId:      app.UserId,
				Nickname:    app.Nickname,
				Avatar:      app.Avatar,
				ContactId:   app.ContactId,
				ContactType: app.ContactType,
				Status:      app.Status,
				Message:     app.Message,
				LastApplyAt: app.LastApplyAt,
			},
		)
	}

	response.Success(
		c, types.GetApplicationListResp{
			Total: len(voList),
			List:  voList,
		},
	)
}

// SubmitApplication 发起好友/群聊申请
// @Summary      发起好友/群聊申请
// @Description  向目标用户发送好友申请(contact_type=0)或向群聊发送加群申请(contact_type=1)，可附加留言
// @Tags         申请
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        req  body      types.SubmitApplicationReq       true  "申请参数"
// @Success      200  {object}  response.Response{data=nil}      "申请已发送"
// @Failure      200  {object}  response.Response{data=nil}      "不能添加自己 / 已经是好友 / 申请处理中"
// @Router       /applications [post]
func (h *ApplicationHandler) SubmitApplication(c *gin.Context) {
	applicantID, ok := middleware.GetUserID(c)
	if !ok {
		response.Fail(c, errcode.TokenInvalid)
		return
	}

	var req types.SubmitApplicationReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errcode.ParamError, "参数格式错误")
		return
	}

	err := h.svc.SubmitApplication(c.Request.Context(), applicantID, req.ContactId, req.ContactType, req.Message)
	if err != nil {
		response.FailErr(c, err)
		return
	}

	response.Success(c, nil)
}

// HandleApplication 处理好友/群聊申请
// @Summary      处理申请
// @Description  对收到的申请进行处理：1=同意（自动建立双向好友关系）、2=拒绝、3=拉黑
// @Tags         申请
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        uuid  path     string                      true  "申请记录 UUID"
// @Param        req   body     types.HandleApplicationReq   true  "处理操作: 1=同意, 2=拒绝, 3=拉黑"
// @Success      200   {object}  response.Response{data=nil}  "处理成功"
// @Failure      200   {object}  response.Response{data=nil}  "申请不存在 / 权限不足 / 已处理过"
// @Router       /applications/{uuid} [put]
func (h *ApplicationHandler) HandleApplication(c *gin.Context) {
	operatorID, ok := middleware.GetUserID(c)
	if !ok {
		response.Fail(c, errcode.TokenInvalid)
		return
	}

	applyUUID := c.Param("uuid")
	if applyUUID == "" {
		response.Fail(c, errcode.ParamError, "申请ID不能为空")
		return
	}

	var req types.HandleApplicationReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errcode.ParamError, "状态参数不合法")
		return
	}

	err := h.svc.HandleApplication(c.Request.Context(), operatorID, applyUUID, req.Status)
	if err != nil {
		response.FailErr(c, err)
		return
	}

	response.Success(c, nil)
}
