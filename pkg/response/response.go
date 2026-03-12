package response

import (
    "errors"
    "net/http"

    "github.com/gin-gonic/gin"

    "wchat/pkg/errcode"
)

type Response struct {
    Code int    `json:"code"`
    Msg  string `json:"msg"`
    Data any    `json:"data"`
}

func result(c *gin.Context, httpStatus int, code int, data any, msg string) {
    c.JSON(
        httpStatus, Response{
            Code: code,
            Msg:  msg,
            Data: data,
        },
    )
}

// Success 返回成功结果，可传入可选的 msg 覆盖默认消息
func Success(c *gin.Context, data any, msgs ...string) {
    msg := errcode.GetMsg(errcode.Success)
    if len(msgs) > 0 && msgs[0] != "" {
        msg = msgs[0]
    }
    result(c, http.StatusOK, errcode.Success, data, msg)
}

// Fail 返回失败结果，可传入可选的 msg 覆盖默认错误消息
func Fail(c *gin.Context, code int, msgs ...string) {
    msg := errcode.GetMsg(code)
    if len(msgs) > 0 && msgs[0] != "" {
        msg = msgs[0]
    }
    // 默认200 OK，业务逻辑错误代码自定义
    result(c, http.StatusOK, code, nil, msg)
}

// FailErr 解析底层传递上来的 error 并返回规范的 JSON
func FailErr(c *gin.Context, err error) {
    if err == nil {
        Success(c, nil)
        return
    }

    var bizErr *errcode.BizError
    if errors.As(err, &bizErr) {
        result(c, http.StatusOK, bizErr.Code, nil, bizErr.Msg)
    } else {
        // TODO: 生产环境应该用 zap 记录以便排查排错
        result(c, http.StatusOK, errcode.ServerError, nil, errcode.GetMsg(errcode.ServerError))
    }
}
