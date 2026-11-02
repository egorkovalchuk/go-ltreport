package confluence

import (
	"testing"
)

func TestAddContentQueryParams_Empty(t *testing.T) {
	query := ContentQuery{}
	result := addContentQueryParams(query)
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
}

func TestAddContentQueryParams_SpaceKey(t *testing.T) {
	query := ContentQuery{SpaceKey: "TEST"}
	result := addContentQueryParams(query)
	if result.Get("spaceKey") != "TEST" {
		t.Errorf("Expected spaceKey='TEST', got '%s'", result.Get("spaceKey"))
	}
}

func TestAddContentQueryParams_Limit(t *testing.T) {
	query := ContentQuery{Limit: 50}
	result := addContentQueryParams(query)
	if result.Get("limit") != "50" {
		t.Errorf("Expected limit='50', got '%s'", result.Get("limit"))
	}
}

func TestAddContentQueryParams_Status(t *testing.T) {
	query := ContentQuery{Status: "current"}
	result := addContentQueryParams(query)
	if result.Get("status") != "current" {
		t.Errorf("Expected status='current', got '%s'", result.Get("status"))
	}
}

func TestAddContentQueryParams_Type(t *testing.T) {
	query := ContentQuery{Type: "page"}
	result := addContentQueryParams(query)
	if result.Get("type") != "page" {
		t.Errorf("Expected type='page', got '%s'", result.Get("type"))
	}
}

func TestAddContentQueryParams_Title(t *testing.T) {
	query := ContentQuery{Title: "Test Page"}
	result := addContentQueryParams(query)
	if result.Get("title") != "Test Page" {
		t.Errorf("Expected title='Test Page', got '%s'", result.Get("title"))
	}
}

func TestAddContentQueryParams_Version(t *testing.T) {
	query := ContentQuery{Version: 5}
	result := addContentQueryParams(query)
	if result.Get("version") != "5" {
		t.Errorf("Expected version='5', got '%s'", result.Get("version"))
	}
}

func TestAddContentQueryParams_Expand(t *testing.T) {
	query := ContentQuery{Expand: []string{"body", "version"}}
	result := addContentQueryParams(query)
	expected := "body,version"
	if result.Get("expand") != expected {
		t.Errorf("Expected expand='%s', got '%s'", expected, result.Get("expand"))
	}
}

func TestNewAPI_EmptyURL(t *testing.T) {
	_, err := NewAPI("", "", "", "", "")
	if err == nil {
		t.Error("Expected error for empty URL")
	}
}

func TestNewAPI_InvalidURL(t *testing.T) {
	_, err := NewAPI("not-a-url", "", "", "", "")
	if err == nil {
		t.Error("Expected error for invalid URL")
	}
}

func TestNewAPI_ValidURL(t *testing.T) {
	api, err := NewAPI("https://example.atlassian.net/wiki", "user", "pass", "", "")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if api == nil {
		t.Fatal("Expected non-nil API")
	}
	if api.Url.String() != "https://example.atlassian.net/wiki" {
		t.Errorf("Expected URL 'https://example.atlassian.net/wiki', got '%s'", api.Url.String())
	}
}

func TestNewAPI_WithToken(t *testing.T) {
	api, err := NewAPI("https://example.atlassian.net/wiki", "", "", "my-token", "")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if api.token != "my-token" {
		t.Errorf("Expected token 'my-token', got '%s'", api.token)
	}
}

func TestNewAPI_WithProxy(t *testing.T) {
	proxyURL := "http://proxy.example.com:8080"
	api, err := NewAPI("https://example.atlassian.net/wiki", "u", "p", "", proxyURL)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if api == nil {
		t.Fatal("Expected non-nil API")
	}
}
