// Package stdbind 为标准库 net/http 提供 reqbind 适配器。
//
// 路径参数依赖 Go 1.22+ 的 http.Request.PathValue()。
//
// 用法：
//
//	mux := http.NewServeMux()
//	mux.HandleFunc("/orgs/{org_id}/users", func(w http.ResponseWriter, r *http.Request) {
//	    req, err := stdbind.Decode[CreateUserReq](r)
//	    ...
//	})
package stdbind

import (
	"io"
	"net/http"

	"github.com/taokiang/reqbind"
)

// Adapter 将 *http.Request 包装为 bind.RequestContext。
type Adapter struct {
	r *http.Request
}

// New 创建一个包装了 r 的 Adapter。
func New(r *http.Request) *Adapter {
	return &Adapter{r: r}
}

// PathParam 返回 URL 路径参数（需要 Go 1.22+ 的路由语法 {name}）。
func (a *Adapter) PathParam(key string) string {
	return a.r.PathValue(key)
}

// Header 返回指定的请求头。
func (a *Adapter) Header(key string) string {
	return a.r.Header.Get(key)
}

// Query 返回 URL 查询参数。
func (a *Adapter) Query(key string) string {
	return a.r.URL.Query().Get(key)
}

// Body 返回请求体。
func (a *Adapter) Body() io.ReadCloser {
	return a.r.Body
}

// ContentType 返回 Content-Type 头。
func (a *Adapter) ContentType() string {
	return a.r.Header.Get("Content-Type")
}

// Decode 将 *http.Request 解码到类型 T。
// 这是最常用的入口函数，等价于 bind.Decode[T](stdbind.New(r))。
func Decode[T any](r *http.Request) (T, error) {
	return bind.Decode[T](New(r))
}
