package httpc

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/geekpratyush/conduit/internal/plugin"
)

func TestSendGETWithParamsAndHeaders(t *testing.T) {
	var gotQuery, gotHeader string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query().Get("active")
		gotHeader = r.Header.Get("X-Trace")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		io.WriteString(w, `{"ok":true}`)
	}))
	defer srv.Close()

	c := New()
	resp, err := c.Send(context.Background(), Request{
		Method:  "GET",
		URL:     srv.URL + "/users",
		Params:  []KeyValue{{Key: "active", Value: "true", Enabled: true}, {Key: "skip", Value: "x", Enabled: false}},
		Headers: []KeyValue{{Key: "X-Trace", Value: "abc", Enabled: true}},
	})
	if err != nil {
		t.Fatalf("send: %v", err)
	}
	if resp.Status != 200 {
		t.Fatalf("status: got %d", resp.Status)
	}
	if gotQuery != "true" {
		t.Fatalf("enabled param not sent: %q", gotQuery)
	}
	if gotHeader != "abc" {
		t.Fatalf("header not sent: %q", gotHeader)
	}
	if resp.Elapsed <= 0 {
		t.Fatal("elapsed should be measured")
	}
	if resp.Size != int64(len(resp.Body)) {
		t.Fatalf("size %d != body len %d", resp.Size, len(resp.Body))
	}
}

func TestSendDisabledParamOmitted(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Has("skip") {
			t.Errorf("disabled param should not be sent")
		}
		w.WriteHeader(204)
	}))
	defer srv.Close()
	c := New()
	if _, err := c.Send(context.Background(), Request{
		URL:    srv.URL,
		Params: []KeyValue{{Key: "skip", Value: "x", Enabled: false}},
	}); err != nil {
		t.Fatalf("send: %v", err)
	}
}

func TestBasicAuth(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		if !ok || u != "ada" || p != "pw" {
			w.WriteHeader(401)
			return
		}
		w.WriteHeader(200)
	}))
	defer srv.Close()
	c := New()
	resp, err := c.Send(context.Background(), Request{
		URL:  srv.URL,
		Auth: Auth{Method: plugin.AuthBasic, Username: "ada", Password: "pw"},
	})
	if err != nil || resp.Status != 200 {
		t.Fatalf("basic auth failed: status=%d err=%v", statusOf(resp), err)
	}
}

func TestBearerAuth(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer t0k" {
			w.WriteHeader(403)
			return
		}
		w.WriteHeader(200)
	}))
	defer srv.Close()
	c := New()
	resp, err := c.Send(context.Background(), Request{
		URL:  srv.URL,
		Auth: Auth{Method: plugin.AuthBearer, Token: "t0k"},
	})
	if err != nil || resp.Status != 200 {
		t.Fatalf("bearer auth failed: status=%d err=%v", statusOf(resp), err)
	}
}

func TestAPIKeyInQuery(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("api_key") != "secret" {
			w.WriteHeader(401)
			return
		}
		w.WriteHeader(200)
	}))
	defer srv.Close()
	c := New()
	resp, err := c.Send(context.Background(), Request{
		URL:  srv.URL,
		Auth: Auth{Method: plugin.AuthAPIKey, APIKeyName: "api_key", APIKeyValue: "secret", APIKeyIn: "query"},
	})
	if err != nil || resp.Status != 200 {
		t.Fatalf("api-key(query) failed: status=%d err=%v", statusOf(resp), err)
	}
}

func TestPostJSONBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ct := r.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
			t.Errorf("content-type: got %q", ct)
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("decode body: %v", err)
		}
		if payload["name"] != "ada" {
			t.Errorf("body not received: %v", payload)
		}
		w.WriteHeader(201)
	}))
	defer srv.Close()
	c := New()
	resp, err := c.Send(context.Background(), Request{
		Method:   "POST",
		URL:      srv.URL,
		BodyType: BodyJSON,
		Body:     `{"name":"ada"}`,
	})
	if err != nil || resp.Status != 201 {
		t.Fatalf("post json failed: status=%d err=%v", statusOf(resp), err)
	}
}

func TestConnectorTest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer srv.Close()
	c := New()
	res := c.Test(context.Background(), plugin.ConnectionConfig{URL: srv.URL})
	if !res.OK {
		t.Fatalf("Test should report OK: %+v", res)
	}
}

func TestSendRejectsRelativeURL(t *testing.T) {
	c := New()
	if _, err := c.Send(context.Background(), Request{URL: "/relative/only"}); err == nil {
		t.Fatal("expected error for non-absolute URL")
	}
}

func statusOf(r *Response) int {
	if r == nil {
		return 0
	}
	return r.Status
}
