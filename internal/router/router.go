package router

import (
    "github.com/gin-contrib/cors"
    "github.com/gin-gonic/gin"

    "wchat/internal/config"
    "wchat/internal/handler"
    "wchat/internal/middleware"
)

// LoadRouters 加载所有路由
func LoadRouters(app *handler.App) *gin.Engine {
    cfg := config.GetConfig()
    r := gin.Default()

    // CORS middleware
    corsConfig := cors.DefaultConfig()
    corsConfig.AllowAllOrigins = true // 开发环境可设为 true，生产环境建议配置具体域名
    corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
    r.Use(cors.New(corsConfig))

    // ==========================================
    // static file server
    // ==========================================
    r.Static("/static", cfg.StaticSrcConfig.StaticFilePath)

    // ==========================================
    // WebSocket connection
    // 客户端连接示例: ws://localhost:8080/ws?token=xxx
    // ==========================================
    r.GET("/ws", app.WebSocket.WsHandler)

    api := r.Group("/api/v1")

    // ==========================================
    // [PUBLIC] No Authentication Required
    // ==========================================
    public := api.Group("")
    {
        public.POST("/auth/register", app.Auth.Register)
        public.POST("/auth/login", app.Auth.Login)
    }

    // ==========================================
    // [PROTECTED] Authentication Required
    // ==========================================
    protected := api.Group("")
    protected.Use(middleware.Auth())
    {
        // -----------------------------
        // Auth & Self (账号与个人中心)
        // -----------------------------
        protected.POST("/auth/logout", app.Auth.Logout)
        protected.GET("/users/me", app.User.GetMyProfile)
        protected.PUT("/users/me", app.User.UpdateMyProfile)

        // -----------------------------
        // Users (系统用户管理/发现)
        // -----------------------------
        users := protected.Group("/users")
        {
            users.GET("", app.User.GetUserList)       // 搜索用户/获取列表
            users.GET("/:uuid", app.User.GetUserInfo) // 查看指定用户信息

            // 管理员操作
            users.PUT("/:uuid/status", app.User.SetUserStatus) // 启用/禁用 (操作 User.Status)
            users.PUT("/:uuid/role", app.User.SetUserRole)     // 设置管理员 (操作 User.IsAdmin)
            users.DELETE("/:uuid", app.User.DeleteUser)        // 删除用户
        }

        // -----------------------------
        // Contacts (联系人管理) 操作 Contact 表
        // -----------------------------
        contacts := protected.Group("/contacts")
        {
            contacts.GET("", app.Contact.GetContactList)         // 获取通讯录列表 (单聊)
            contacts.DELETE("/:uuid", app.Contact.DeleteContact) // 删除好友

            // 黑名单机制
            contacts.POST("/:uuid/block", app.Contact.BlockContact)     // 拉黑
            contacts.DELETE("/:uuid/block", app.Contact.UnblockContact) // 移出黑名单
        }

        // -----------------------------
        // Groups (群组管理) 操作 Group 表
        // -----------------------------
        groups := protected.Group("/groups")
        {
            groups.POST("", app.Group.CreateGroup)           // 建群
            groups.GET("/joined", app.Group.GetJoinedGroups) // 获取我加入的群
            groups.GET("/:uuid", app.Group.GetGroupInfo)     // 获取群资料
            groups.PUT("/:uuid", app.Group.UpdateGroupInfo)  // 修改群资料/公告
            groups.DELETE("/:uuid", app.Group.DismissGroup)  // 解散群聊 (群主权限)

            // 群成员管理
            groups.GET("/:uuid/members", app.Group.GetGroupMembers)        // 获取群成员列表
            groups.POST("/:uuid/members", app.Group.InviteToGroup)         // 直接拉人进群
            groups.DELETE("/:uuid/members/me", app.Group.LeaveGroup)       // 退出群聊
            groups.DELETE("/:uuid/members/:user_id", app.Group.KickMember) // 踢出群聊 (群主/管理员权限)
        }

        // -----------------------------
        // Applications (申请大厅) 操作 ContactApply 表
        // -----------------------------
        applications := protected.Group("/applications")
        {
            applications.GET("", app.Application.GetApplicationList)      // 获取我收到的申请
            applications.POST("", app.Application.SubmitApplication)      // 发起加好友/加群申请
            applications.PUT("/:uuid", app.Application.HandleApplication) // 处理申请 (同意/拒绝)
        }

        // -----------------------------
        // Sessions (会话列表管理) 操作 Session 表
        // -----------------------------
        sessions := protected.Group("/sessions")
        {
            sessions.GET("", app.Session.GetSessionList)              // 获取首页聊天列表
            sessions.POST("", app.Session.CreateSession)              // 点击头像发起聊天时，初始化会话
            sessions.DELETE("/:uuid", app.Session.DeleteSession)      // 从列表中删除某个会话
            sessions.PUT("/:uuid/top", app.Session.PinSession)        // 置顶/取消置顶 (操作 IsTop)
            sessions.PUT("/:uuid/read", app.Session.ClearUnreadCount) // 标记已读 (UnreadCount 归零)
        }

        // -----------------------------
        // Messages (消息历史) 操作 Message 表
        // -----------------------------
        messages := protected.Group("/messages")
        {
            messages.GET("", app.Message.GetMessageList)
            messages.PUT("/:uuid/recall", app.Message.RecallMessage) // 操作 Status=2
        }
    }

    return r
}
