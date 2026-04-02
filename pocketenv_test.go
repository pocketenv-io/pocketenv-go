package pocketenv

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func newTestServer(t *testing.T, routes map[string]http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	mux := http.NewServeMux()
	for path, handler := range routes {
		mux.HandleFunc(path, handler)
	}
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	c, err := New(WithBaseURL(srv.URL), WithToken("test-token"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return c, srv
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func assertBearer(t *testing.T, r *http.Request, want string) {
	t.Helper()
	got := r.Header.Get("Authorization")
	if got != "Bearer "+want {
		t.Errorf("Authorization header = %q, want %q", got, "Bearer "+want)
	}
}

func TestNew_defaults(t *testing.T) {
	c, err := New(WithToken("test"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if c.baseURL != defaultBaseURL {
		t.Errorf("baseURL = %q, want %q", c.baseURL, defaultBaseURL)
	}
}

func TestNew_noToken(t *testing.T) {
	// Ensure no token file exists for this test by pointing home somewhere empty.
	t.Setenv("HOME", t.TempDir())
	_, err := New()
	if err == nil {
		t.Fatal("expected error when no token provided")
	}
}

func TestWithBaseURL(t *testing.T) {
	c, err := New(WithBaseURL("https://api.example.com"), WithToken("test"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if c.baseURL != "https://api.example.com" {
		t.Errorf("baseURL = %q", c.baseURL)
	}
}

func TestWithToken(t *testing.T) {
	c, err := New(WithToken("my-token"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if c.token != "my-token" {
		t.Errorf("token = %q", c.token)
	}
}

func TestDo_HTTPError(t *testing.T) {
	c, _ := newTestServer(t, map[string]http.HandlerFunc{
		"/xrpc/io.pocketenv.sandbox.getSandbox": func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		},
	})
	_, err := c.GetSandbox("missing")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "404") {
		t.Errorf("error should mention HTTP 404, got: %v", err)
	}
}

func TestCreateSandbox(t *testing.T) {
	want := SandboxView{ID: "s1", Name: "my-sandbox", Status: "RUNNING"}
	c, _ := newTestServer(t, map[string]http.HandlerFunc{
		"/xrpc/io.pocketenv.sandbox.createSandbox": func(w http.ResponseWriter, r *http.Request) {
			assertBearer(t, r, "test-token")
			if r.Method != http.MethodPost {
				t.Errorf("method = %s, want POST", r.Method)
			}
			var body CreateSandboxInput
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body.Base != "openclaw" {
				t.Errorf("base = %q, want openclaw", body.Base)
			}
			writeJSON(w, want)
		},
	})
	got, err := c.CreateSandbox(CreateSandboxInput{Base: "openclaw"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != want.ID || got.Name != want.Name {
		t.Errorf("got %+v, want %+v", got.SandboxView, want)
	}
}

func TestGetSandbox(t *testing.T) {
	want := SandboxView{ID: "s1", Name: "box"}
	c, _ := newTestServer(t, map[string]http.HandlerFunc{
		"/xrpc/io.pocketenv.sandbox.getSandbox": func(w http.ResponseWriter, r *http.Request) {
			assertBearer(t, r, "test-token")
			if r.URL.Query().Get("id") != "s1" {
				t.Errorf("id param missing")
			}
			writeJSON(w, map[string]any{"sandbox": want})
		},
	})
	got, err := c.GetSandbox("s1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != want.ID {
		t.Errorf("got %+v, want %+v", got.SandboxView, want)
	}
}

func TestListSandboxes(t *testing.T) {
	boxes := []SandboxView{{ID: "s1"}, {ID: "s2"}}
	c, _ := newTestServer(t, map[string]http.HandlerFunc{
		"/xrpc/io.pocketenv.sandbox.getSandboxes": func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query()
			if q.Get("offset") != "0" || q.Get("limit") != "10" {
				t.Errorf("unexpected pagination params: %v", q)
			}
			writeJSON(w, map[string]any{"sandboxes": boxes, "total": 2})
		},
	})
	got, total, err := c.ListSandboxes(0, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 || len(got) != 2 {
		t.Errorf("got %d/%d sandboxes, want 2/2", len(got), total)
	}
	// each item is a usable handle
	if got[0].ID != "s1" || got[1].ID != "s2" {
		t.Errorf("unexpected IDs: %v %v", got[0].ID, got[1].ID)
	}
}

func TestSandbox_StartStop(t *testing.T) {
	called := map[string]bool{}
	c, _ := newTestServer(t, map[string]http.HandlerFunc{
		"/xrpc/io.pocketenv.sandbox.startSandbox": func(w http.ResponseWriter, r *http.Request) {
			called["start"] = true
			if r.URL.Query().Get("id") != "s1" {
				t.Errorf("missing id param")
			}
			writeJSON(w, map[string]any{})
		},
		"/xrpc/io.pocketenv.sandbox.stopSandbox": func(w http.ResponseWriter, r *http.Request) {
			called["stop"] = true
			writeJSON(w, map[string]any{})
		},
	})
	sb := c.Sandbox("s1")
	if err := sb.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if err := sb.Stop(); err != nil {
		t.Fatalf("Stop: %v", err)
	}
	if !called["start"] || !called["stop"] {
		t.Errorf("not all endpoints were called: %v", called)
	}
}

func TestSandbox_Delete(t *testing.T) {
	called := false
	c, _ := newTestServer(t, map[string]http.HandlerFunc{
		"/xrpc/io.pocketenv.sandbox.deleteSandbox": func(w http.ResponseWriter, r *http.Request) {
			called = true
			writeJSON(w, map[string]any{})
		},
	})
	if err := c.Sandbox("s1").Delete(); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if !called {
		t.Error("deleteSandbox was not called")
	}
}

func TestSandbox_Exec(t *testing.T) {
	c, _ := newTestServer(t, map[string]http.HandlerFunc{
		"/xrpc/io.pocketenv.sandbox.exec": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]string
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body["command"] != "echo hello" {
				t.Errorf("command = %q", body["command"])
			}
			writeJSON(w, ExecResult{Stdout: "hello\n", Stderr: "", ExitCode: 0})
		},
	})
	res, err := c.Sandbox("s1").Exec("echo", "hello")
	if err != nil {
		t.Fatalf("Exec: %v", err)
	}
	if res.Stdout != "hello\n" || res.ExitCode != 0 {
		t.Errorf("unexpected result: %+v", res)
	}
}

func TestSandbox_ExposeUnexposePort(t *testing.T) {
	called := map[string]bool{}
	c, _ := newTestServer(t, map[string]http.HandlerFunc{
		"/xrpc/io.pocketenv.sandbox.exposePort": func(w http.ResponseWriter, r *http.Request) {
			called["expose"] = true
			writeJSON(w, map[string]any{})
		},
		"/xrpc/io.pocketenv.sandbox.unexposePort": func(w http.ResponseWriter, r *http.Request) {
			called["unexpose"] = true
			writeJSON(w, map[string]any{})
		},
	})
	sb := c.Sandbox("s1")
	if err := sb.ExposePort(3000, "app"); err != nil {
		t.Fatalf("ExposePort: %v", err)
	}
	if err := sb.UnexposePort(3000); err != nil {
		t.Fatalf("UnexposePort: %v", err)
	}
	if !called["expose"] || !called["unexpose"] {
		t.Errorf("not all port endpoints called: %v", called)
	}
}

func TestSandbox_GetExposedPorts(t *testing.T) {
	want := []ExposedPort{{Port: 8080, Description: "web", URL: "https://x.example.com"}}
	c, _ := newTestServer(t, map[string]http.HandlerFunc{
		"/xrpc/io.pocketenv.sandbox.getExposedPorts": func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, map[string]any{"ports": want})
		},
	})
	got, err := c.Sandbox("s1").GetExposedPorts()
	if err != nil {
		t.Fatalf("GetExposedPorts: %v", err)
	}
	if len(got) != 1 || got[0].Port != 8080 {
		t.Errorf("got %+v", got)
	}
}

func TestSandbox_SshKeys(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	want := SshKeysView{ID: "k1", PublicKey: "ssh-rsa AAAA", PrivateKey: "-----BEGIN", CreatedAt: now, UpdatedAt: now}
	c, _ := newTestServer(t, map[string]http.HandlerFunc{
		"/xrpc/io.pocketenv.sandbox.getSshKeys": func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, map[string]any{"sshKeys": want})
		},
		"/xrpc/io.pocketenv.sandbox.putSshKeys": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]string
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body["publicKey"] == "" || body["privateKey"] == "" {
				t.Errorf("missing ssh key fields: %v", body)
			}
			writeJSON(w, map[string]any{})
		},
	})
	sb := c.Sandbox("s1")
	got, err := sb.GetSshKeys()
	if err != nil {
		t.Fatalf("GetSshKeys: %v", err)
	}
	if got.PublicKey != want.PublicKey {
		t.Errorf("got %+v, want %+v", got, want)
	}
	if err := sb.PutSshKeys("ssh-rsa AAAA", "-----BEGIN"); err != nil {
		t.Fatalf("PutSshKeys: %v", err)
	}
}

func TestSecretClient_CRUD(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	s := SecretView{ID: "sec1", Name: "DB_PASS", Value: "secret", SandboxID: "s1", CreatedAt: now, UpdatedAt: now}

	c, _ := newTestServer(t, map[string]http.HandlerFunc{
		"/xrpc/io.pocketenv.secret.addSecret": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			inner, _ := body["secret"].(map[string]any)
			if inner["name"] != "DB_PASS" {
				t.Errorf("unexpected body: %v", body)
			}
			writeJSON(w, s)
		},
		"/xrpc/io.pocketenv.secret.getSecrets": func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, map[string]any{"secrets": []SecretView{s}, "total": 1})
		},
		"/xrpc/io.pocketenv.secret.getSecret": func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("id") != "sec1" {
				t.Errorf("id param = %q", r.URL.Query().Get("id"))
			}
			writeJSON(w, s)
		},
		"/xrpc/io.pocketenv.secret.updateSecret": func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, s)
		},
		"/xrpc/io.pocketenv.secret.deleteSecret": func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("id") != "sec1" {
				t.Errorf("id param = %q", r.URL.Query().Get("id"))
			}
			writeJSON(w, map[string]any{})
		},
	})

	sc := c.Sandbox("s1").Secrets()

	added, err := sc.Add("DB_PASS", "secret")
	if err != nil || added.ID != "sec1" {
		t.Fatalf("Add: %v, got %+v", err, added)
	}

	secrets, total, err := sc.List(0, 10)
	if err != nil || total != 1 || len(secrets) != 1 {
		t.Fatalf("List: %v, total=%d len=%d", err, total, len(secrets))
	}

	got, err := sc.Get("sec1")
	if err != nil || got.ID != "sec1" {
		t.Fatalf("Get: %v, got %+v", err, got)
	}

	updated, err := sc.Update("sec1", "DB_PASS", "new-secret")
	if err != nil || updated.ID != "sec1" {
		t.Fatalf("Update: %v, got %+v", err, updated)
	}

	if err := sc.Delete("sec1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

func TestVariableClient_CRUD(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	v := VariableView{ID: "var1", Name: "PORT", Value: "8080", SandboxID: "s1", CreatedAt: now, UpdatedAt: now}

	c, _ := newTestServer(t, map[string]http.HandlerFunc{
		"/xrpc/io.pocketenv.variable.addVariable": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			inner, _ := body["variable"].(map[string]any)
			if inner["name"] != "PORT" {
				t.Errorf("unexpected name: %v", inner)
			}
			writeJSON(w, v)
		},
		"/xrpc/io.pocketenv.variable.getVariables": func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, map[string]any{"variables": []VariableView{v}, "total": 1})
		},
		"/xrpc/io.pocketenv.variable.getVariable": func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, v)
		},
		"/xrpc/io.pocketenv.variable.updateVariable": func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, v)
		},
		"/xrpc/io.pocketenv.variable.deleteVariable": func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, map[string]any{})
		},
	})

	vc := c.Sandbox("s1").Variables()

	added, err := vc.Add("PORT", "8080")
	if err != nil || added.ID != "var1" {
		t.Fatalf("Add: %v, got %+v", err, added)
	}

	vars, total, err := vc.List(0, 10)
	if err != nil || total != 1 || len(vars) != 1 {
		t.Fatalf("List: %v, total=%d len=%d", err, total, len(vars))
	}

	got, err := vc.Get("var1")
	if err != nil || got.Name != "PORT" {
		t.Fatalf("Get: %v, got %+v", err, got)
	}

	updated, err := vc.Update("var1", "PORT", "9090")
	if err != nil || updated.ID != "var1" {
		t.Fatalf("Update: %v", err)
	}

	if err := vc.Delete("var1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

// ── Files ─────────────────────────────────────────────────────────────────────

func TestFileClient_CRUD(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	f := FileView{ID: "f1", Path: "/app/main.go", SandboxID: "s1", CreatedAt: now, UpdatedAt: now}

	c, _ := newTestServer(t, map[string]http.HandlerFunc{
		"/xrpc/io.pocketenv.file.addFile": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			inner, _ := body["file"].(map[string]any)
			if inner["path"] != "/app/main.go" {
				t.Errorf("unexpected path: %v", inner)
			}
			writeJSON(w, map[string]any{})
		},
		"/xrpc/io.pocketenv.file.getFiles": func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("sandboxId") != "s1" {
				t.Errorf("sandboxId param missing")
			}
			writeJSON(w, map[string]any{"files": []FileView{f}, "total": 1})
		},
		"/xrpc/io.pocketenv.file.getFile": func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, map[string]any{"file": f})
		},
		"/xrpc/io.pocketenv.file.updateFile": func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, map[string]any{})
		},
		"/xrpc/io.pocketenv.file.deleteFile": func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("id") != "f1" {
				t.Errorf("id param = %q", r.URL.Query().Get("id"))
			}
			writeJSON(w, map[string]any{})
		},
	})

	fc := c.Sandbox("s1").Files()

	if err := fc.Add("/app/main.go", "package main"); err != nil {
		t.Fatalf("Add: %v", err)
	}

	files, total, err := fc.List(0, 20)
	if err != nil || total != 1 || len(files) != 1 {
		t.Fatalf("List: %v, total=%d len=%d", err, total, len(files))
	}

	got, err := fc.Get("f1")
	if err != nil || got.Path != "/app/main.go" {
		t.Fatalf("Get: %v, got %+v", err, got)
	}

	if err := fc.Update("f1", "/app/main.go", "package main\n"); err != nil {
		t.Fatalf("Update: %v", err)
	}

	if err := fc.Delete("f1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

// ── Volumes ───────────────────────────────────────────────────────────────────

func TestVolumeClient_CRUD(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	vol := VolumeView{ID: "v1", Name: "data", Path: "/data", SandboxID: "s1", CreatedAt: now, UpdatedAt: now}

	c, _ := newTestServer(t, map[string]http.HandlerFunc{
		"/xrpc/io.pocketenv.volume.addVolume": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			inner, _ := body["volume"].(map[string]any)
			if inner["path"] != "/data" {
				t.Errorf("unexpected path: %v", inner)
			}
			writeJSON(w, map[string]any{})
		},
		"/xrpc/io.pocketenv.volume.getVolumes": func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("sandboxId") != "s1" {
				t.Errorf("sandboxId param missing")
			}
			writeJSON(w, map[string]any{"volumes": []VolumeView{vol}, "total": 1})
		},
		"/xrpc/io.pocketenv.volume.getVolume": func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, map[string]any{"volume": vol})
		},
		"/xrpc/io.pocketenv.volume.updateVolume": func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, map[string]any{})
		},
		"/xrpc/io.pocketenv.volume.deleteVolume": func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("id") != "v1" {
				t.Errorf("id param = %q", r.URL.Query().Get("id"))
			}
			writeJSON(w, map[string]any{})
		},
	})

	vc := c.Sandbox("s1").Volumes()

	if err := vc.Add("data", "/data"); err != nil {
		t.Fatalf("Add: %v", err)
	}

	vols, total, err := vc.List(0, 20)
	if err != nil || total != 1 || len(vols) != 1 {
		t.Fatalf("List: %v, total=%d len=%d", err, total, len(vols))
	}

	got, err := vc.Get("v1")
	if err != nil || got.Name != "data" {
		t.Fatalf("Get: %v, got %+v", err, got)
	}

	if err := vc.Update("v1", "data", "/mnt/data"); err != nil {
		t.Fatalf("Update: %v", err)
	}

	if err := vc.Delete("v1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

func TestServiceClient_CRUD(t *testing.T) {
	svc := ServiceView{ID: "svc1", Name: "web", Command: "node server.js", Status: "RUNNING", SandboxID: "s1"}

	c, _ := newTestServer(t, map[string]http.HandlerFunc{
		"/xrpc/io.pocketenv.service.addService": func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("sandboxId") != "s1" {
				t.Errorf("sandboxId param missing, query: %v", r.URL.RawQuery)
			}
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			inner, _ := body["service"].(map[string]any)
			if inner["name"] != "web" {
				t.Errorf("unexpected service name: %v", inner)
			}
			writeJSON(w, map[string]any{})
		},
		"/xrpc/io.pocketenv.service.getServices": func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("sandboxId") != "s1" {
				t.Errorf("sandboxId param missing")
			}
			writeJSON(w, map[string]any{"services": []ServiceView{svc}})
		},
		"/xrpc/io.pocketenv.service.updateService": func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("serviceId") != "svc1" {
				t.Errorf("serviceId param = %q", r.URL.Query().Get("serviceId"))
			}
			writeJSON(w, map[string]any{})
		},
		"/xrpc/io.pocketenv.service.startService": func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("serviceId") != "svc1" {
				t.Errorf("serviceId param = %q", r.URL.Query().Get("serviceId"))
			}
			writeJSON(w, map[string]any{})
		},
		"/xrpc/io.pocketenv.service.stopService": func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, map[string]any{})
		},
		"/xrpc/io.pocketenv.service.restartService": func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, map[string]any{})
		},
		"/xrpc/io.pocketenv.service.deleteService": func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("serviceId") != "svc1" {
				t.Errorf("serviceId param = %q", r.URL.Query().Get("serviceId"))
			}
			writeJSON(w, map[string]any{})
		},
	})

	sc := c.Sandbox("s1").Services()

	if err := sc.Add(AddServiceInput{Name: "web", Command: "node server.js", Ports: []int{3000}}); err != nil {
		t.Fatalf("Add: %v", err)
	}

	svcs, err := sc.List()
	if err != nil || len(svcs) != 1 || svcs[0].ID != "svc1" {
		t.Fatalf("List: %v, got %+v", err, svcs)
	}

	if err := sc.Update("svc1", UpdateServiceInput{Name: "web-v2"}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	if err := sc.Start("svc1"); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if err := sc.Stop("svc1"); err != nil {
		t.Fatalf("Stop: %v", err)
	}
	if err := sc.Restart("svc1"); err != nil {
		t.Fatalf("Restart: %v", err)
	}
	if err := sc.Delete("svc1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

// ── Helpers scoped to Client ───────────────────────────────────────────────────

func TestClient_SandboxHandle(t *testing.T) {
	c, err := New(WithToken("test"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	sb := c.Sandbox("abc")
	if sb.ID != "abc" {
		t.Errorf("ID = %q, want abc", sb.ID)
	}
}

func TestClient_get_queryParams(t *testing.T) {
	var gotQuery url.Values
	c, _ := newTestServer(t, map[string]http.HandlerFunc{
		"/xrpc/io.pocketenv.sandbox.getSandboxes": func(w http.ResponseWriter, r *http.Request) {
			gotQuery = r.URL.Query()
			writeJSON(w, map[string]any{"sandboxes": []SandboxView{}, "total": 0})
		},
	})
	_, _, err := c.ListSandboxes(5, 25)
	if err != nil {
		t.Fatalf("ListSandboxes: %v", err)
	}
	if gotQuery.Get("offset") != "5" || gotQuery.Get("limit") != "25" {
		t.Errorf("unexpected query params: %v", gotQuery)
	}
}
