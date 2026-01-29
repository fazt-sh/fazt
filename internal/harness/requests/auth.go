package requests

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"

	"github.com/fazt-sh/fazt/internal/harness"
	"github.com/fazt-sh/fazt/internal/harness/gaps"
)

// AuthTest defines an authentication lifecycle test.
type AuthTest struct {
	Name        string
	Method      string
	Path        string
	Body        interface{}
	Headers     map[string]string
	Cookies     []*http.Cookie
	Expected    AuthExpected
	Description string
}

// AuthExpected defines expected authentication behavior.
type AuthExpected struct {
	Status          int
	SetCookie       bool   // Expects Set-Cookie header
	CookieName      string // Expected cookie name
	CookieCleared   bool   // Cookie should be cleared (MaxAge=0)
	BodyContains    string
	BodyNotContains string
}

// DefaultAuthTests returns standard authentication tests.
func DefaultAuthTests() []AuthTest {
	return []AuthTest{
		{
			Name:   "login_success",
			Method: "POST",
			Path:   "/api/login",
			Body:   map[string]string{"username": "admin", "password": "test-password"},
			Expected: AuthExpected{
				Status:     200,
				SetCookie:  true,
				CookieName: "session",
			},
			Description: "Successful login sets session cookie",
		},
		{
			Name:   "login_wrong_password",
			Method: "POST",
			Path:   "/api/login",
			Body:   map[string]string{"username": "admin", "password": "wrong-password"},
			Expected: AuthExpected{
				Status:          401,
				BodyContains:    "INVALID_CREDENTIALS",
				BodyNotContains: "session",
			},
			Description: "Wrong password returns 401",
		},
		{
			Name:   "login_wrong_username",
			Method: "POST",
			Path:   "/api/login",
			Body:   map[string]string{"username": "nonexistent", "password": "test"},
			Expected: AuthExpected{
				Status:       401,
				BodyContains: "INVALID_CREDENTIALS",
			},
			Description: "Wrong username returns 401",
		},
		{
			Name:   "login_missing_fields",
			Method: "POST",
			Path:   "/api/login",
			Body:   map[string]string{},
			Expected: AuthExpected{
				Status: 400,
			},
			Description: "Missing credentials returns 400",
		},
		{
			Name:   "protected_no_session",
			Method: "GET",
			Path:   "/api/apps",
			Expected: AuthExpected{
				Status: 401,
			},
			Description: "Protected endpoint without session returns 401",
		},
	}
}

// SessionLifecycleTests returns tests for full session lifecycle.
func SessionLifecycleTests() []AuthTest {
	return []AuthTest{
		{
			Name:   "session_refresh",
			Method: "GET",
			Path:   "/api/me",
			// Requires valid session cookie (set up in runner)
			Expected: AuthExpected{
				Status:       200,
				BodyContains: "username",
			},
			Description: "Session refresh extends expiry",
		},
		{
			Name:   "logout",
			Method: "POST",
			Path:   "/api/logout",
			Expected: AuthExpected{
				Status:        200,
				SetCookie:     true,
				CookieCleared: true,
			},
			Description: "Logout clears session cookie",
		},
		{
			Name:   "after_logout_401",
			Method: "GET",
			Path:   "/api/me",
			Expected: AuthExpected{
				Status: 401,
			},
			Description: "Request after logout returns 401",
		},
	}
}

// AuthRunner executes authentication tests.
type AuthRunner struct {
	client     *http.Client
	baseURL    string
	gapTracker *gaps.Tracker
	// Test credentials
	username string
	password string
}

// NewAuthRunner creates a new auth test runner.
func NewAuthRunner(baseURL, username, password string, gapTracker *gaps.Tracker) *AuthRunner {
	jar, _ := cookiejar.New(nil)
	return &AuthRunner{
		client: &http.Client{
			Timeout: 10 * time.Second,
			Jar:     jar,
		},
		baseURL:    baseURL,
		gapTracker: gapTracker,
		username:   username,
		password:   password,
	}
}

// Run executes all auth tests.
func (r *AuthRunner) Run(ctx context.Context) []harness.TestResult {
	tests := DefaultAuthTests()
	results := make([]harness.TestResult, 0, len(tests))

	// Update test credentials
	for i := range tests {
		if tests[i].Name == "login_success" {
			tests[i].Body = map[string]string{
				"username": r.username,
				"password": r.password,
			}
		}
	}

	for _, test := range tests {
		result := r.runTest(ctx, test)
		results = append(results, result)
	}

	return results
}

// RunSessionLifecycle runs the full session lifecycle test.
func (r *AuthRunner) RunSessionLifecycle(ctx context.Context) []harness.TestResult {
	var results []harness.TestResult

	// Step 1: Login
	loginResult := r.login(ctx)
	results = append(results, loginResult)
	if !loginResult.Passed {
		return results // Can't continue without login
	}

	// Step 2: Access protected endpoint
	protectedResult := r.runTest(ctx, AuthTest{
		Name:   "session_protected_access",
		Method: "GET",
		Path:   "/api/me",
		Expected: AuthExpected{
			Status:       200,
			BodyContains: "username",
		},
	})
	results = append(results, protectedResult)

	// Step 3: Logout
	logoutResult := r.runTest(ctx, AuthTest{
		Name:   "session_logout",
		Method: "POST",
		Path:   "/api/logout",
		Expected: AuthExpected{
			Status: 200,
		},
	})
	results = append(results, logoutResult)

	// Step 4: Verify can't access after logout
	afterLogoutResult := r.runTest(ctx, AuthTest{
		Name:   "session_after_logout",
		Method: "GET",
		Path:   "/api/me",
		Expected: AuthExpected{
			Status: 401,
		},
	})
	results = append(results, afterLogoutResult)

	return results
}

func (r *AuthRunner) login(ctx context.Context) harness.TestResult {
	return r.runTest(ctx, AuthTest{
		Name:   "session_login",
		Method: "POST",
		Path:   "/api/login",
		Body:   map[string]string{"username": r.username, "password": r.password},
		Expected: AuthExpected{
			Status:     200,
			SetCookie:  true,
			CookieName: "session",
		},
	})
}

func (r *AuthRunner) runTest(ctx context.Context, test AuthTest) harness.TestResult {
	start := time.Now()
	result := harness.TestResult{
		Name:     test.Name,
		Category: "auth",
	}

	var bodyReader io.Reader
	if test.Body != nil {
		bodyBytes, _ := json.Marshal(test.Body)
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, test.Method, r.baseURL+test.Path, bodyReader)
	if err != nil {
		result.Error = err
		result.Duration = time.Since(start)
		return result
	}

	if bodyReader != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range test.Headers {
		req.Header.Set(k, v)
	}

	resp, err := r.client.Do(req)
	result.Duration = time.Since(start)

	if err != nil {
		result.Error = err
		return result
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))

	result.Actual = harness.Actual{
		Status:  resp.StatusCode,
		Latency: result.Duration,
		Body:    string(body),
		Headers: make(map[string]string),
	}

	// Capture Set-Cookie
	if setCookie := resp.Header.Get("Set-Cookie"); setCookie != "" {
		result.Actual.Headers["Set-Cookie"] = setCookie
	}

	// Validate
	result.Passed = r.validate(test, resp, body)

	// Record gaps for security issues
	if !result.Passed && r.gapTracker != nil {
		severity := gaps.SeverityMedium
		if test.Expected.Status == 401 && resp.StatusCode != 401 {
			severity = gaps.SeverityCritical // Auth bypass
		}

		gap := gaps.Gap{
			Category:     gaps.CategorySecurity,
			Severity:     severity,
			Description:  test.Name + ": authentication behavior mismatch",
			DiscoveredBy: "auth_" + test.Name,
		}
		gapID := r.gapTracker.Record(gap)
		result.Gap = &gaps.Gap{ID: gapID}
	}

	return result
}

func (r *AuthRunner) validate(test AuthTest, resp *http.Response, body []byte) bool {
	// Status check
	if resp.StatusCode != test.Expected.Status {
		return false
	}

	// Set-Cookie check
	if test.Expected.SetCookie {
		setCookie := resp.Header.Get("Set-Cookie")
		if setCookie == "" {
			return false
		}
		if test.Expected.CookieName != "" && !strings.Contains(setCookie, test.Expected.CookieName) {
			return false
		}
		if test.Expected.CookieCleared && !strings.Contains(setCookie, "Max-Age=0") {
			return false
		}
	}

	// Body content check
	if test.Expected.BodyContains != "" {
		if !strings.Contains(string(body), test.Expected.BodyContains) {
			return false
		}
	}
	if test.Expected.BodyNotContains != "" {
		if strings.Contains(string(body), test.Expected.BodyNotContains) {
			return false
		}
	}

	return true
}

// APIKeyTests returns API key authentication tests.
func APIKeyTests() []AuthTest {
	return []AuthTest{
		{
			Name:    "api_key_auth",
			Method:  "GET",
			Path:    "/api/apps",
			Headers: map[string]string{"Authorization": "Bearer valid-api-key"},
			Expected: AuthExpected{
				Status: 200,
			},
			Description: "Valid API key grants access",
		},
		{
			Name:    "api_key_invalid",
			Method:  "GET",
			Path:    "/api/apps",
			Headers: map[string]string{"Authorization": "Bearer invalid-key"},
			Expected: AuthExpected{
				Status:       401,
				BodyContains: "INVALID_API_KEY",
			},
			Description: "Invalid API key returns 401",
		},
		{
			Name:    "api_key_malformed",
			Method:  "GET",
			Path:    "/api/apps",
			Headers: map[string]string{"Authorization": "NotBearer something"},
			Expected: AuthExpected{
				Status: 401,
			},
			Description: "Malformed auth header returns 401",
		},
	}
}

// RunAPIKeyTests tests API key authentication.
func (r *AuthRunner) RunAPIKeyTests(ctx context.Context, validKey string) []harness.TestResult {
	tests := APIKeyTests()
	var results []harness.TestResult

	// Create a fresh client without cookie jar
	client := &http.Client{Timeout: 10 * time.Second}
	originalClient := r.client
	r.client = client
	defer func() { r.client = originalClient }()

	for _, test := range tests {
		// Replace placeholder with actual key
		if strings.Contains(test.Headers["Authorization"], "valid-api-key") {
			test.Headers["Authorization"] = "Bearer " + validKey
		}

		result := r.runTest(ctx, test)
		results = append(results, result)
	}

	return results
}

// SessionSecurityTests tests session security properties.
func (r *AuthRunner) RunSecurityTests(ctx context.Context) []harness.TestResult {
	var results []harness.TestResult

	// Test: Session cookie should be HttpOnly
	loginResult := r.login(ctx)
	if loginResult.Actual.Headers["Set-Cookie"] != "" {
		setCookie := loginResult.Actual.Headers["Set-Cookie"]
		httpOnly := strings.Contains(strings.ToLower(setCookie), "httponly")

		results = append(results, harness.TestResult{
			Name:     "session_httponly",
			Category: "auth",
			Passed:   httpOnly,
			Actual:   harness.Actual{Body: setCookie},
		})

		if !httpOnly && r.gapTracker != nil {
			r.gapTracker.Record(gaps.Gap{
				Category:     gaps.CategorySecurity,
				Severity:     gaps.SeverityHigh,
				Description:  "Session cookie missing HttpOnly flag",
				DiscoveredBy: "auth_session_httponly",
				Remediation:  "Add HttpOnly to session cookie",
			})
		}

		// Test: Session cookie should have SameSite
		sameSite := strings.Contains(strings.ToLower(setCookie), "samesite")
		results = append(results, harness.TestResult{
			Name:     "session_samesite",
			Category: "auth",
			Passed:   sameSite,
		})

		if !sameSite && r.gapTracker != nil {
			r.gapTracker.Record(gaps.Gap{
				Category:     gaps.CategorySecurity,
				Severity:     gaps.SeverityMedium,
				Description:  "Session cookie missing SameSite attribute",
				DiscoveredBy: "auth_session_samesite",
				Remediation:  "Add SameSite=Lax or SameSite=Strict",
			})
		}
	}

	return results
}
