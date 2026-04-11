package errcode

const (
	// ==========================================
	// 全局系统错误 (10000 - 19999)
	// ==========================================
	Success     = 0
	ServerError = 10001
	ParamError  = 10002
	NotFound    = 10003
	TooManyReq  = 10004

	// ==========================================
	// Auth 鉴权与安全模块 (20000 - 20999)
	// ==========================================
	AuthFailed   = 20001 // 账号或密码错误
	TokenInvalid = 20002 // Token 无效或解析失败
	TokenExpired = 20003 // Token 已过期
	TokenMissing = 20004 // 缺少 Token
	Unauthorized = 20005 // 权限不足（如非管理员尝试越权）

	// ==========================================
	// User 用户模块 (21000 - 21999)
	// ==========================================
	UserExists             = 21001
	UserNotFound           = 21002
	AccountDisabled        = 21003 // 账号被禁用
	InvalidPassword        = 21004 // 密码格式不符合要求或旧密码错误
	AccountPendingDeletion = 21005 // 账号已申请注销，处于冷静期

	// ==========================================
	// Contact & Application 好友与申请模块 (30000 - 30999)
	// ==========================================
	ContactNotFound     = 30001 // 好友关系不存在
	AlreadyFriends      = 30002 // 已经是好友，无需重复添加
	CannotAddSelf       = 30003 // 不能添加自己为好友
	TargetBlocked       = 30004 // 对方已被拉黑或你被对方拉黑
	ApplyNotFound       = 30005 // 申请记录不存在
	ApplyAlreadyHandled = 30006 // 该申请已被处理（已同意或拒绝）

	// ==========================================
	// Group 群组模块 (40000 - 40999)
	// ==========================================
	GroupNotFound     = 40001
	NotInGroup        = 40002 // 你不在该群聊中
	AlreadyInGroup    = 40003 // 已经在群聊中
	NoGroupPermission = 40004 // 无群组操作权限（非群主或管理员）
	GroupDismissed    = 40005 // 该群聊已解散
	GroupFull         = 40006 // 群聊人数已满

	// ==========================================
	// Session & Message 会话与消息模块 (50000 - 50999)
	// ==========================================
	SessionNotFound      = 50001
	MessageNotFound      = 50002
	MessageRecallTimeout = 50003 // 消息撤回超时（通常设定为 2 分钟）
	UnsupportedMsgType   = 50004 // 不支持的消息类型
)

// TODO: International sufficiency (后续可引入 i18n 多语言包)
var msgFlags = map[int]string{
	// 全局
	Success:     "ok",
	ServerError: "系统内部错误，请稍后再试",
	ParamError:  "请求参数错误或格式不合法",
	NotFound:    "请求的资源不存在",
	TooManyReq:  "请求过于频繁，请稍后再试",

	// 鉴权
	AuthFailed:   "用户名或密码错误",
	TokenInvalid: "登录凭证无效，请重新登录",
	TokenExpired: "登录已过期，请重新登录",
	TokenMissing: "未授权的请求，缺少凭证",
	Unauthorized: "权限不足，拒绝访问",

	// 用户
	UserExists:             "该手机号/账号已被注册",
	UserNotFound:           "用户不存在",
	AccountDisabled:        "当前账号已被冻结或禁用，请联系管理员",
	InvalidPassword:        "密码错误或不符合规范",
	AccountPendingDeletion: "当前账号已申请注销，处于冷静期，请先取消注销",

	// 联系人与申请
	ContactNotFound:     "你们还不是好友",
	AlreadyFriends:      "你们已经是好友了，请勿重复添加",
	CannotAddSelf:       "不能添加自己为好友",
	TargetBlocked:       "无法发送好友申请，由于对方的隐私设置或黑名单状态",
	ApplyNotFound:       "该申请记录不存在或已被撤销",
	ApplyAlreadyHandled: "该申请已被处理，请勿重复操作",

	// 群组
	GroupNotFound:     "该群聊不存在",
	NotInGroup:        "你不在该群聊中，无法查看或发送消息",
	AlreadyInGroup:    "你已经在此群聊中",
	NoGroupPermission: "只有群主或管理员才能执行此操作",
	GroupDismissed:    "该群聊已被解散",
	GroupFull:         "该群聊人数已满",

	// 消息与会话
	SessionNotFound:      "聊天会话不存在",
	MessageNotFound:      "该消息不存在或已被删除",
	MessageRecallTimeout: "消息发送已超过时限，无法撤回",
	UnsupportedMsgType:   "当前版本暂不支持查看该类型消息",
}

func GetMsg(code int) string {
	msg, ok := msgFlags[code]
	if ok {
		return msg
	}
	// 兜底返回未知错误
	return msgFlags[ServerError]
}

type BizError struct {
	Code int
	Msg  string
}

// Error implements error interface
func (e *BizError) Error() string {
	return e.Msg
}

func New(code int) error {
	return &BizError{
		Code: code,
		Msg:  GetMsg(code),
	}
}

func NewWithMsg(code int, msg string) error {
	return &BizError{
		Code: code,
		Msg:  msg,
	}
}
