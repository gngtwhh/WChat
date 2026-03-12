package render

import (
    "html/template"
    "wchat/pkg/errcode"
    "wchat/pkg/response"

    "github.com/gin-gonic/gin"
)

type Renderer struct {
    tmpls map[string]*template.Template
    entry string
}

var renderer *Renderer

func Init(t map[string]*template.Template, entry string) {
    renderer = &Renderer{
        tmpls: t,
        entry: entry,
    }
}

// Execute executes a template with the given data and writes the result to gin.Context.Writer.
func Execute(c *gin.Context, name string, data interface{}) {
    if renderer == nil {
        response.Fail(c, errcode.ServerError, "template not initialized")
        return
    }

    tmpl, ok := renderer.tmpls[name]
    if !ok {
        response.Fail(c, errcode.ServerError, "template not found")
        return
    }

    c.Header("Content-Type", "text/html; charset=utf-8")

    err := tmpl.ExecuteTemplate(c.Writer, renderer.entry, data)
    if err != nil {
        response.Fail(c, errcode.ServerError)
    }
}
