package bind_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/taokiang/reqbind"
	stdbind "github.com/taokiang/reqbind/std"
)

// ── 测试辅助：mock RequestContext ─────────────────────────────────────────────

type mockCtx struct {
	path   map[string]string
	header map[string]string
	query  map[string]string
	body   io.ReadCloser
}

func (m *mockCtx) PathParam(k string) string { return m.path[k] }
func (m *mockCtx) Header(k string) string    { return m.header[k] }
func (m *mockCtx) Query(k string) string     { return m.query[k] }
func (m *mockCtx) Body() io.ReadCloser       { return m.body }
func (m *mockCtx) ContentType() string       { return m.header["Content-Type"] }

func jsonBody(v any) io.ReadCloser {
	data, _ := json.Marshal(v)
	return io.NopCloser(bytes.NewReader(data))
}

func textBody(s string) io.ReadCloser {
	return io.NopCloser(strings.NewReader(s))
}

// ── 测试用结构体 ──────────────────────────────────────────────────────────────

type SimpleReq struct {
	OrgID string `path:"org_id"`
	Token string `header:"Authorization"`
	Page  int    `query:"page" default:"1"`
	Size  int    `query:"size" default:"20"`
}

type BodyUser struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type CreateUserReq struct {
	OrgID string   `path:"org_id"`
	Token string   `header:"Authorization"`
	Page  int      `query:"page" default:"1"`
	Body  BodyUser `body:"json"`
}

type FormProfile struct {
	Name  string
	Email string
}

type FormReq struct {
	Username string      `path:"username"`
	Profile  FormProfile `body:"form"`
}

type TextReq struct {
	Raw string `body:"text"`
}

// ── 基础绑定测试 ───────────────────────────────────────────────────────────────

func TestDecode_Simple(t *testing.T) {
	ctx := &mockCtx{
		path:   map[string]string{"org_id": "acme"},
		header: map[string]string{"Authorization": "Bearer tok123"},
		query:  map[string]string{"page": "3"},
	}

	req, err := bind.Decode[SimpleReq](ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.OrgID != "acme" {
		t.Errorf("OrgID: want %q, got %q", "acme", req.OrgID)
	}
	if req.Token != "Bearer tok123" {
		t.Errorf("Token: want %q, got %q", "Bearer tok123", req.Token)
	}
	if req.Page != 3 {
		t.Errorf("Page: want 3, got %d", req.Page)
	}
	if req.Size != 20 {
		t.Errorf("Size (default): want 20, got %d", req.Size)
	}
}

func TestDecode_Default(t *testing.T) {
	ctx := &mockCtx{
		path:   map[string]string{},
		header: map[string]string{},
		query:  map[string]string{},
	}

	req, err := bind.Decode[SimpleReq](ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Page != 1 {
		t.Errorf("Page default: want 1, got %d", req.Page)
	}
	if req.Size != 20 {
		t.Errorf("Size default: want 20, got %d", req.Size)
	}
}

func TestDecode_JSONBody(t *testing.T) {
	ctx := &mockCtx{
		path:   map[string]string{"org_id": "acme"},
		header: map[string]string{"Authorization": "tok"},
		query:  map[string]string{},
		body:   jsonBody(BodyUser{Name: "Alice", Email: "alice@example.com"}),
	}

	req, err := bind.Decode[CreateUserReq](ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Body.Name != "Alice" {
		t.Errorf("Body.Name: want %q, got %q", "Alice", req.Body.Name)
	}
	if req.Body.Email != "alice@example.com" {
		t.Errorf("Body.Email: want %q, got %q", "alice@example.com", req.Body.Email)
	}
	if req.Page != 1 {
		t.Errorf("Page default: want 1, got %d", req.Page)
	}
}

func TestDecode_FormBody(t *testing.T) {
	ctx := &mockCtx{
		path:   map[string]string{"username": "bob"},
		header: map[string]string{},
		query:  map[string]string{},
		body:   textBody("name=Bob+Smith&email=bob%40example.com"),
	}

	req, err := bind.Decode[FormReq](ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Username != "bob" {
		t.Errorf("Username: want %q, got %q", "bob", req.Username)
	}
	if req.Profile.Name != "Bob Smith" {
		t.Errorf("Profile.Name: want %q, got %q", "Bob Smith", req.Profile.Name)
	}
}

func TestDecode_TextBody(t *testing.T) {
	ctx := &mockCtx{
		path:   map[string]string{},
		header: map[string]string{},
		query:  map[string]string{},
		body:   textBody("hello world"),
	}

	req, err := bind.Decode[TextReq](ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Raw != "hello world" {
		t.Errorf("Raw: want %q, got %q", "hello world", req.Raw)
	}
}

// ── 类型转换测试 ───────────────────────────────────────────────────────────────

type TypesReq struct {
	Count  uint64        `query:"count"`
	Score  float64       `query:"score"`
	Active bool          `query:"active"`
	TTL    time.Duration `query:"ttl"`
}

func TestDecode_TypeConversion(t *testing.T) {
	ctx := &mockCtx{
		path:   map[string]string{},
		header: map[string]string{},
		query: map[string]string{
			"count":  "42",
			"score":  "3.14",
			"active": "true",
			"ttl":    "5m30s",
		},
	}

	req, err := bind.Decode[TypesReq](ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Count != 42 {
		t.Errorf("Count: want 42, got %d", req.Count)
	}
	if req.Score != 3.14 {
		t.Errorf("Score: want 3.14, got %f", req.Score)
	}
	if !req.Active {
		t.Errorf("Active: want true, got false")
	}
	if req.TTL != 5*time.Minute+30*time.Second {
		t.Errorf("TTL: want 5m30s, got %v", req.TTL)
	}
}

// ── 指针字段测试 ───────────────────────────────────────────────────────────────

type PtrReq struct {
	OrgID *string `path:"org_id"`
	Page  *int    `query:"page"`
}

func TestDecode_PointerFields(t *testing.T) {
	ctx := &mockCtx{
		path:   map[string]string{"org_id": "acme"},
		header: map[string]string{},
		query:  map[string]string{"page": "5"},
	}

	req, err := bind.Decode[PtrReq](ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.OrgID == nil || *req.OrgID != "acme" {
		t.Errorf("OrgID: want ptr(%q), got %v", "acme", req.OrgID)
	}
	if req.Page == nil || *req.Page != 5 {
		t.Errorf("Page: want ptr(5), got %v", req.Page)
	}
}

func TestDecode_PointerFields_Missing(t *testing.T) {
	ctx := &mockCtx{
		path:   map[string]string{},
		header: map[string]string{},
		query:  map[string]string{},
	}

	req, err := bind.Decode[PtrReq](ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.OrgID != nil {
		t.Errorf("OrgID: want nil, got %v", req.OrgID)
	}
	if req.Page != nil {
		t.Errorf("Page: want nil, got %v", req.Page)
	}
}

// ── 返回指针类型测试 ───────────────────────────────────────────────────────────

func TestDecode_PtrResult(t *testing.T) {
	ctx := &mockCtx{
		path:   map[string]string{"org_id": "acme"},
		header: map[string]string{},
		query:  map[string]string{},
	}

	req, err := bind.Decode[*SimpleReq](ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req == nil {
		t.Fatal("expected non-nil pointer")
	}
	if req.OrgID != "acme" {
		t.Errorf("OrgID: want %q, got %q", "acme", req.OrgID)
	}
}

// ── 错误处理测试 ───────────────────────────────────────────────────────────────

func TestDecode_InvalidInt(t *testing.T) {
	ctx := &mockCtx{
		path:   map[string]string{},
		header: map[string]string{},
		query:  map[string]string{"page": "notanumber"},
	}

	_, err := bind.Decode[SimpleReq](ctx)
	if err == nil {
		t.Error("expected error for invalid int, got nil")
	}
}

type DuplicateBodyReq struct {
	A struct{ X string } `body:"json"`
	B struct{ Y string } `body:"json"`
}

func TestDecode_DuplicateBody(t *testing.T) {
	ctx := &mockCtx{path: map[string]string{}, header: map[string]string{}, query: map[string]string{}}
	_, err := bind.Decode[DuplicateBodyReq](ctx)
	if err == nil {
		t.Error("expected error for duplicate body fields, got nil")
	}
}

// ── 缓存一致性测试 ─────────────────────────────────────────────────────────────

func TestDecode_CacheConsistency(t *testing.T) {
	for i := 0; i < 10; i++ {
		ctx := &mockCtx{
			path:   map[string]string{"org_id": "org42"},
			header: map[string]string{"Authorization": "tok"},
			query:  map[string]string{"page": "7"},
		}
		req, err := bind.Decode[SimpleReq](ctx)
		if err != nil {
			t.Fatalf("iter %d: unexpected error: %v", i, err)
		}
		if req.OrgID != "org42" || req.Page != 7 {
			t.Fatalf("iter %d: unexpected values: %+v", i, req)
		}
	}
}

// ── net/http 适配器集成测试 ───────────────────────────────────────────────────

func TestStdAdapter_GET(t *testing.T) {
	type ListUsersReq struct {
		OrgID string `path:"org_id"`
		Token string `header:"Authorization"`
		Page  int    `query:"page" default:"1"`
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/orgs/{org_id}/users", func(w http.ResponseWriter, r *http.Request) {
		req, err := stdbind.Decode[ListUsersReq](r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if req.OrgID != "acme" {
			t.Errorf("OrgID: want %q, got %q", "acme", req.OrgID)
		}
		if req.Token != "Bearer secret" {
			t.Errorf("Token: want %q, got %q", "Bearer secret", req.Token)
		}
		if req.Page != 3 {
			t.Errorf("Page: want 3, got %d", req.Page)
		}
		w.WriteHeader(http.StatusOK)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	httpReq, _ := http.NewRequest("GET", srv.URL+"/orgs/acme/users?page=3", nil)
	httpReq.Header.Set("Authorization", "Bearer secret")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status: want 200, got %d", resp.StatusCode)
	}
}

func TestStdAdapter_JSONBody(t *testing.T) {
	type CreateReq struct {
		OrgID string   `path:"org_id"`
		Body  BodyUser `body:"json"`
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/orgs/{org_id}/users", func(w http.ResponseWriter, r *http.Request) {
		req, err := stdbind.Decode[CreateReq](r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if req.OrgID != "acme" {
			t.Errorf("OrgID: want %q, got %q", "acme", req.OrgID)
		}
		if req.Body.Name != "Alice" {
			t.Errorf("Body.Name: want %q, got %q", "Alice", req.Body.Name)
		}
		w.WriteHeader(http.StatusCreated)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	payload, _ := json.Marshal(BodyUser{Name: "Alice", Email: "alice@example.com"})
	httpReq, _ := http.NewRequest("POST", srv.URL+"/orgs/acme/users", bytes.NewReader(payload))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("status: want 201, got %d", resp.StatusCode)
	}
}
