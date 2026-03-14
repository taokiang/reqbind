// Package bind 是一个框架无关的 HTTP 请求结构体绑定库。
//
// 通过结构体标签从多个来源（path、header、query、body）同时绑定请求数据：
//
//	type CreateUserReq struct {
//	    OrgID string `path:"org_id"`
//	    Token string `header:"Authorization"`
//	    Page  int    `query:"page" default:"1"`
//	    Body  User   `body:"json"`
//	}
//
//	req, err := bind.Decode[CreateUserReq](ctx)
package bind

import "io"

// RequestContext 是框架适配器需要实现的抽象接口。
// 每个 HTTP 框架（net/http、gin、echo 等）提供一个实现此接口的适配器。
type RequestContext interface {
	// PathParam 返回 URL 路径参数，如路由 /orgs/:org_id 中的 "org_id"。
	PathParam(key string) string

	// Header 返回指定名称的 HTTP 请求头。
	Header(key string) string

	// Query 返回 URL 查询参数，如 ?page=2 中的 "page"。
	Query(key string) string

	// Body 返回请求体。无请求体时返回 nil。
	Body() io.ReadCloser

	// ContentType 返回 Content-Type 头的值。
	ContentType() string
}

// Decode 将 HTTP 请求解码到类型 T 的结构体中。
//
// 支持的结构体标签：
//
//	path:"key"      从 URL 路径参数读取
//	header:"key"    从 HTTP 请求头读取
//	query:"key"     从 URL 查询参数读取
//	body:"json"     将请求体作为 JSON 解码到该字段
//	body:"xml"      将请求体作为 XML 解码到该字段
//	body:"form"     将请求体作为 URL 编码表单解码到该字段
//	body:"text"     将请求体作为纯文本读取到 string 字段
//	default:"value" 当来源为空时使用的默认值
//
// path/header/query 支持的字段类型：
// string、bool、int/int8/int16/int32/int64、uint/uint8/uint16/uint32/uint64、
// float32/float64、time.Duration 以及以上类型的指针（*T）。
//
// 每个结构体最多只能有一个 body 字段。
func Decode[T any](ctx RequestContext) (T, error) {
	return decodeInto[T](ctx)
}
