package metrics

import "testing"

func TestRouteLabel(t *testing.T) {
	t.Cleanup(func() { routeRe = nil })

	if got := routeLabel("/anything"); got != unmatchedRoute {
		t.Fatalf("no regex: got %q, want %q", got, unmatchedRoute)
	}

	if err := SetRouteRegex(`^/$|^/health$|^/assets/`); err != nil {
		t.Fatal(err)
	}
	cases := map[string]string{
		"/":              "/",
		"/health":        "/health",
		"/assets/app.js": "/assets/app.js",
		"/about":         unmatchedRoute,
		"/assets":        unmatchedRoute,
	}
	for path, want := range cases {
		if got := routeLabel(path); got != want {
			t.Errorf("path %q: got %q, want %q", path, got, want)
		}
	}

	if err := SetRouteRegex(`^(/api/[^/]+)`); err != nil {
		t.Fatal(err)
	}
	if got := routeLabel("/api/users/42/profile"); got != "/api/users" {
		t.Errorf("capture group: got %q, want /api/users", got)
	}
	if got := routeLabel("/login"); got != unmatchedRoute {
		t.Errorf("unmatched capture: got %q, want %q", got, unmatchedRoute)
	}
}
