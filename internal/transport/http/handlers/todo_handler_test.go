package handlers_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"samokat_todo/internal/transport/http/handlers"
	"samokat_todo/internal/transport/http/router"
	"strconv"
	"testing"
	"time"

	"samokat_todo/internal/repository/memory"
	"samokat_todo/internal/service"
)

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	repo := memory.NewTodoRepository()
	svc := service.NewTodoService(repo)

	h := handlers.NewHandler(svc, 2*time.Second)
	r := router.NewRouter(h)

	return httptest.NewServer(r)
}

func doJSON(t *testing.T, client *http.Client, method, url string, body []byte) *http.Response {
	t.Helper()

	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	return resp
}

func readAll(t *testing.T, r io.Reader) []byte {
	t.Helper()
	b, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return b
}

func createTodo(t *testing.T, baseURL string) (id int) {
	t.Helper()

	client := &http.Client{Timeout: time.Second}
	resp := doJSON(t, client, http.MethodPost, baseURL+"/todos", []byte(`{"title":"a","description":"b","completed":false}`))
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected %d, got %d, body=%s", http.StatusCreated, resp.StatusCode, string(readAll(t, resp.Body)))
	}

	var got map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}

	v, ok := got["id"]
	if !ok {
		t.Fatalf("expected id in response, got: %#v", got)
	}

	f, ok := v.(float64)
	if !ok {
		t.Fatalf("expected id float64, got %T (%v)", v, v)
	}

	return int(f)
}

func TestPOST_Todos_Success(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	client := &http.Client{Timeout: time.Second}
	resp := doJSON(t, client, http.MethodPost, ts.URL+"/todos", []byte(`{"title":"buy milk","description":"2 liters","completed":false}`))
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected %d, got %d, body=%s", http.StatusCreated, resp.StatusCode, string(readAll(t, resp.Body)))
	}

	var got map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if _, ok := got["id"]; !ok {
		t.Fatalf("expected id in response, got: %#v", got)
	}
}

func TestPOST_Todos_ValidationError(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	client := &http.Client{Timeout: time.Second}
	resp := doJSON(t, client, http.MethodPost, ts.URL+"/todos", []byte(`{"title":"   ","description":"x","completed":false}`))
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d, body=%s", http.StatusBadRequest, resp.StatusCode, string(readAll(t, resp.Body)))
	}
}

func TestGET_Todos_List(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	_ = createTodo(t, ts.URL)
	_ = createTodo(t, ts.URL)

	client := &http.Client{Timeout: time.Second}
	resp, err := client.Get(ts.URL + "/todos")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected %d, got %d, body=%s", http.StatusOK, resp.StatusCode, string(readAll(t, resp.Body)))
	}

	var list []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 todos, got %d", len(list))
	}
}

func TestGET_TodosByID_Success(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	id := createTodo(t, ts.URL)

	client := &http.Client{Timeout: time.Second}
	resp, err := client.Get(ts.URL + "/todos/" + strconv.Itoa(id))
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected %d, got %d, body=%s", http.StatusOK, resp.StatusCode, string(readAll(t, resp.Body)))
	}
}

func TestGET_TodosByID_NotFound(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	client := &http.Client{Timeout: time.Second}
	resp, err := client.Get(ts.URL + "/todos/999")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected %d, got %d, body=%s", http.StatusNotFound, resp.StatusCode, string(readAll(t, resp.Body)))
	}
}

func TestPUT_Todos_ValidationError(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	id := createTodo(t, ts.URL)

	client := &http.Client{Timeout: time.Second}
	resp := doJSON(t, client, http.MethodPut, ts.URL+"/todos/"+strconv.Itoa(id), []byte(`{"title":"   ","description":"x","completed":false}`))
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d, body=%s", http.StatusBadRequest, resp.StatusCode, string(readAll(t, resp.Body)))
	}
}

func TestPUT_Todos_NotFound(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	client := &http.Client{Timeout: time.Second}
	resp := doJSON(t, client, http.MethodPut, ts.URL+"/todos/999", []byte(`{"title":"x","description":"y","completed":true}`))
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected %d, got %d, body=%s", http.StatusNotFound, resp.StatusCode, string(readAll(t, resp.Body)))
	}
}

func TestDELETE_Todos_Success(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	id := createTodo(t, ts.URL)

	client := &http.Client{Timeout: time.Second}
	req, err := http.NewRequest(http.MethodDelete, ts.URL+"/todos/"+strconv.Itoa(id), nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected %d, got %d, body=%s", http.StatusNoContent, resp.StatusCode, string(readAll(t, resp.Body)))
	}

	resp2, err := client.Get(ts.URL + "/todos/" + strconv.Itoa(id))
	if err != nil {
		t.Fatalf("get after delete: %v", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusNotFound {
		t.Fatalf("expected %d, got %d", http.StatusNotFound, resp2.StatusCode)
	}
}

func TestRequestID_HeaderIsSet(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	client := &http.Client{Timeout: time.Second}
	resp := doJSON(t, client, http.MethodPost, ts.URL+"/todos", []byte(`{"title":"a","description":"b","completed":false}`))
	defer resp.Body.Close()

	if rid := resp.Header.Get("X-Request-Id"); rid == "" {
		t.Fatalf("expected X-Request-Id header to be set")
	}
}
func TestGET_Todos_Pagination(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	createTodo(t, ts.URL)
	createTodo(t, ts.URL)
	createTodo(t, ts.URL)

	client := &http.Client{Timeout: time.Second}
	resp, err := client.Get(ts.URL + "/todos?limit=2&offset=1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var list []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 todos, got %d", len(list))
	}
}
