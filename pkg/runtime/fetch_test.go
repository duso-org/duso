package runtime

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// redirectServer answers /login with a 302 that sets a cookie, and /dashboard
// with a plain 200.
func redirectServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: "session", Value: "abc"})
		http.Redirect(w, r, "/dashboard", http.StatusFound)
	})
	mux.HandleFunc("/dashboard", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("dashboard"))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func fetchResult(t *testing.T, args map[string]any) map[string]any {
	t.Helper()
	out, err := builtinFetch(nil, args)
	if err != nil {
		t.Fatalf("fetch failed: %v", err)
	}
	return out.(map[string]any)
}

func TestFetchFollowsRedirectsByDefault(t *testing.T) {
	srv := redirectServer(t)
	r := fetchResult(t, map[string]any{"0": srv.URL + "/login"})
	if r["status"] != float64(200) || r["body"] != "dashboard" {
		t.Errorf("got status %v body %q, want the followed 200", r["status"], r["body"])
	}
}

// follow_redirects = false returns the 3xx itself, so a script can read the
// Set-Cookie that a followed redirect would throw away.
func TestFetchFollowRedirectsFalse(t *testing.T) {
	srv := redirectServer(t)
	opts := map[string]any{"follow_redirects": false}
	r := fetchResult(t, map[string]any{"0": srv.URL + "/login", "1": opts})
	if r["status"] != float64(302) {
		t.Fatalf("status = %v, want 302", r["status"])
	}
	headers := r["headers"].(map[string]any)
	if headers["Location"] != "/dashboard" {
		t.Errorf("Location = %v, want /dashboard", headers["Location"])
	}
	if headers["Set-Cookie"] != "session=abc" {
		t.Errorf("Set-Cookie = %v, want session=abc", headers["Set-Cookie"])
	}
}
