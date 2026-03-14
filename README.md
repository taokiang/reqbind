# reqbind

框架无关的 Go HTTP 请求绑定库。通过结构体标签，一次调用即可从 URL 路径参数、请求头、查询参数、请求体等多个来源同步绑定数据。

## 特性

- **多来源绑定**：`path`、`header`、`query`、`body` 标签，一个结构体覆盖所有来源
- **框架无关**：核心包零外部依赖，通过适配器支持任意框架
- **泛型 API**：`Decode[T](ctx)` 返回强类型结果，无需手动断言
- **自动类型转换**：字符串自动转换为 int、bool、float、`time.Duration` 等类型
- **默认值**：`default:"value"` 标签，来源为空时自动填充
- **字段元数据缓存**：首次解析后缓存 struct 反射结果，热路径无额外开销

## 安装

```bash
# 核心库（含 net/http 适配器）
go get github.com/troytao/reqbind

# gin 适配器（独立模块，不使用 gin 则无需安装）
go get github.com/troytao/reqbind/adapters/gin
```

> **环境要求**：Go 1.22+（`net/http` 适配器的路径参数依赖 Go 1.22 引入的 `r.PathValue()`）

---

## 快速上手

### 定义请求结构体

```go
type CreateUserReq struct {
    OrgID string `path:"org_id"`              // 从 URL 路径参数读取
    Token string `header:"Authorization"`      // 从请求头读取
    Page  int    `query:"page" default:"1"`    // 从查询参数读取，默认值 1
    Size  int    `query:"size" default:"20"`   // 从查询参数读取，默认值 20
    Body  User   `body:"json"`                 // 将请求体解析为 JSON
}
```

### 标准库 `net/http`

```go
import stdbind "github.com/troytao/reqbind/std"

mux := http.NewServeMux()
mux.HandleFunc("POST /orgs/{org_id}/users", func(w http.ResponseWriter, r *http.Request) {
    req, err := stdbind.Decode[CreateUserReq](r)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    // req.OrgID  → 路径参数 org_id
    // req.Token  → 请求头 Authorization
    // req.Page   → 查询参数 page（未传时为 1）
    // req.Body   → JSON 请求体解析结果
})
```

### Gin

```go
import ginbind "github.com/troytao/reqbind/adapters/gin"

r := gin.Default()
r.POST("/orgs/:org_id/users", func(c *gin.Context) {
    req, err := ginbind.Decode[CreateUserReq](c)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    // 用法同上
})
```

---

## 标签参考

### `path:"key"` — URL 路径参数

从 URL 路径段读取。

```go
// 路由：/orgs/{org_id}/users/{user_id}
type Req struct {
    OrgID  string `path:"org_id"`
    UserID int64  `path:"user_id"`
}
```

- `net/http` 使用花括号语法：`/orgs/{org_id}`
- `gin` 使用冒号语法：`/orgs/:org_id`

### `header:"key"` — 请求头

按标准 HTTP 头名称读取（不区分大小写由框架保证）。

```go
type Req struct {
    Token     string `header:"Authorization"`
    TraceID   string `header:"X-Trace-Id"`
    UserAgent string `header:"User-Agent"`
}
```

### `query:"key"` — 查询参数

从 URL 查询字符串读取，`?key=value` 中的 `key`。

```go
// GET /users?page=2&size=10&active=true
type Req struct {
    Page   int  `query:"page"   default:"1"`
    Size   int  `query:"size"   default:"20"`
    Active bool `query:"active"`
}
```

### `body:"format"` — 请求体

将整个请求体解析为指定格式，填入该字段。**每个结构体最多只能有一个 body 字段。**

| 格式 | 说明 | 字段类型要求 |
|------|------|-------------|
| `body:"json"` | JSON 格式 | `struct` 或 `*struct` |
| `body:"xml"` | XML 格式 | `struct` 或 `*struct` |
| `body:"form"` | `application/x-www-form-urlencoded` | `struct` 或 `*struct` |
| `body:"text"` | 纯文本，原样读取 | `string` |

```go
type CreateReq struct {
    OrgID string `path:"org_id"`
    Body  User   `body:"json"`    // JSON 请求体 → User 结构体
}

type UploadReq struct {
    Name    string `path:"name"`
    Content string `body:"text"`  // 纯文本请求体 → string
}
```

### `default:"value"` — 默认值

当来源为空（未传参、头不存在等）时使用该值。支持与 `path`、`header`、`query` 搭配使用（`body` 字段不支持默认值）。

```go
type Req struct {
    Page    int    `query:"page"    default:"1"`
    Size    int    `query:"size"    default:"20"`
    Version string `header:"X-API-Version" default:"v1"`
}
```

---

## 支持的字段类型

### 标量类型（适用于 path / header / query）

| Go 类型 | 说明 |
|---------|------|
| `string` | 原样读取 |
| `bool` | 接受 `true`、`false`、`1`、`0` 等（`strconv.ParseBool`） |
| `int`、`int8`、`int16`、`int32`、`int64` | 十进制整数 |
| `uint`、`uint8`、`uint16`、`uint32`、`uint64` | 十进制无符号整数 |
| `float32`、`float64` | 浮点数 |
| `time.Duration` | Go duration 字符串，如 `5m30s`、`100ms` |
| `time.Time` | RFC3339 格式，如 `2026-03-14T10:00:00Z` |
| `*T`（上述类型的指针） | 未传参时字段为 `nil`，传参时自动分配 |

### body 字段类型

| 格式 | 字段类型 |
|------|---------|
| `json` / `xml` / `form` | `struct` 或 `*struct` |
| `text` | `string` |

---

## form body 的字段映射

使用 `body:"form"` 时，内部结构体字段通过 `form:"key"` 标签指定表单键名；若无该标签，则使用**字段名的全小写**。

```go
type Address struct {
    City    string `form:"city"`     // 映射表单字段 "city"
    ZipCode string `form:"zip_code"` // 映射表单字段 "zip_code"
    Country string                   // 无标签，映射表单字段 "country"（字段名小写）
}

type Req struct {
    Body Address `body:"form"`
}
```

---

## 指针字段的语义

对 `path`、`header`、`query` 使用指针类型，可以区分「未传参」和「传了空值」两种情况：

```go
type Req struct {
    Filter *string `query:"filter"` // 未传 → nil；传了 → 指向值的指针
    Page   *int    `query:"page"`
}

req, _ := bind.Decode[Req](ctx)
if req.Filter == nil {
    // 客户端未传 filter 参数
} else {
    // 使用 *req.Filter
}
```

---

## 自定义框架适配器

只需实现 `bind.RequestContext` 接口，即可接入任意框架：

```go
import "github.com/troytao/reqbind"

type EchoAdapter struct {
    c echo.Context
}

func (a *EchoAdapter) PathParam(key string) string  { return a.c.Param(key) }
func (a *EchoAdapter) Header(key string) string     { return a.c.Request().Header.Get(key) }
func (a *EchoAdapter) Query(key string) string      { return a.c.QueryParam(key) }
func (a *EchoAdapter) Body() io.ReadCloser          { return a.c.Request().Body }
func (a *EchoAdapter) ContentType() string          { return a.c.Request().Header.Get("Content-Type") }

// 使用
func echoHandler(c echo.Context) error {
    req, err := bind.Decode[MyReq](&EchoAdapter{c})
    ...
}
```

`RequestContext` 接口定义：

```go
type RequestContext interface {
    PathParam(key string) string
    Header(key string) string
    Query(key string) string
    Body() io.ReadCloser
    ContentType() string
}
```

---

## 错误处理

所有错误均包含来源上下文，便于定位问题：

```go
req, err := stdbind.Decode[MyReq](r)
if err != nil {
    // 错误示例：
    // bind: query["page"]: cannot parse "abc" as int: ...
    // bind: body["json"]: invalid character ...
    // bind: struct MyReq has more than one body field
    log.Println(err)
}
```

---

## 项目结构

```
reqbind/
├── go.mod                  # module github.com/troytao/reqbind（零外部依赖）
├── go.work                 # Go workspace，关联核心模块与适配器
├── bind.go                 # RequestContext 接口 + Decode[T] 公开 API
├── field.go                # struct tag 解析 + sync.Map 字段元数据缓存
├── convert.go              # 字符串到 Go 类型的转换逻辑
├── body.go                 # body 格式解码（json / xml / form / text）
├── decode.go               # 核心反射绑定实现
├── bind_test.go            # 单元测试 + net/http 集成测试
├── std/
│   └── std.go              # package stdbind：net/http 适配器
└── adapters/
    └── gin/
        ├── go.mod          # 独立模块，依赖 gin
        └── gin.go          # package ginbind：gin 适配器
```

核心模块与 gin 适配器**分开作为两个 Go module**，使用核心库无需引入 gin 依赖。

---

## 初始化

```bash
# 克隆后初始化（生成 go.sum）
cd reqbind
go mod tidy

# gin 适配器单独初始化
cd adapters/gin
go mod tidy
```

使用 Go workspace 开发时：

```bash
# 在根目录，两个模块会自动通过 replace 指令关联
go work sync
```
