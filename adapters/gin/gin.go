// Package ginbind 为 Gin 框架提供 reqbind 适配器。
//
// 用法：
//
//	r := gin.Default()
//	r.GET("/orgs/:org_id/users", func(c *gin.Context) {
//	    req, err := ginbind.Decode[CreateUserReq](c)
//	    ...
//	})
package ginbind

import (
	"io"

	"github.com/gin-gonic/gin"
	"github.com/taokiang/reqbind"
)

// Adapter 将 *gin.Context 包装为 bind.RequestContext。
type Adapter struct {
	c *gin.Context
}

// New 创建一个包装了 c 的 Adapter。
func New(c *gin.Context) *Adapter {
	return &Adapter{c: c}
}

// PathParam 返回 Gin 路由参数（如 /users/:id 中的 "id"）。
func (a *Adapter) PathParam(key string) string {
	return a.c.Param(key)
}

// Header 返回指定的请求头。
func (a *Adapter) Header(key string) string {
	return a.c.GetHeader(key)
}

// Query 返回 URL 查询参数。
func (a *Adapter) Query(key string) string {
	return a.c.Query(key)
}

// Body 返回请求体。
func (a *Adapter) Body() io.ReadCloser {
	return a.c.Request.Body
}

// ContentType 返回 Content-Type 头。
func (a *Adapter) ContentType() string {
	return a.c.ContentType()
}

// Decode 将 *gin.Context 解码到类型 T。
// 这是最常用的入口函数，等价于 bind.Decode[T](ginbind.New(c))。
func Decode[T any](c *gin.Context) (T, error) {
	return bind.Decode[T](New(c))
}
