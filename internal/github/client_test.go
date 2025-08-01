package github

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	client := NewClient("mytoken")
	assert.Equal(t, "mytoken", client.Token)
	assert.NotNil(t, client.httpClient)
}

func TestNewRequest(t *testing.T) {
	client := NewClient("token123")
	method := "GET"
	path := "/test/path"
	req, err := client.newRequest(method, path)
	assert.NoError(t, err)
	assert.Equal(t, "GET", req.Method)
	assert.Equal(t, "https://api.github.com/test/path", req.URL.String())
	assert.Equal(t, "Bearer token123", req.Header.Get("Authorization"))
}

func TestDo(t *testing.T) {
	// Use a test server to mock responses
	ts := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("ok"))
		assert.NoError(t, err)
	})
	server := http.Server{Handler: ts}
	defer func() {
		err := server.Close()
		assert.NoError(t, err)
	}()

	client := NewClient("")
	client.httpClient = &http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       http.NoBody,
			Header:     make(http.Header),
		}
		return resp, nil
	})}

	req, _ := http.NewRequest("GET", "http://example.com", nil)
	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
