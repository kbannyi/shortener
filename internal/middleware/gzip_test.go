package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kbannyi/shortener/internal/config"
	"github.com/kbannyi/shortener/internal/router"
	"github.com/stretchr/testify/require"
)

type MockService struct{}

func (s *MockService) Create(value string) (ID string) {
	return "mockid"
}

func (s *MockService) Get(ID string) (string, bool) {
	return "redirect", ID == "mockid"
}

func TestGzipCompression(t *testing.T) {
	handler := GZIPMiddleware(router.NewURLRouter(&MockService{}, config.Flags{
		RedirectBaseAddr: "http://localhost:8080/",
	}))

	srv := httptest.NewServer(handler)
	defer srv.Close()

	requestBody := `{"url": "https://go.dev/doc/effective_go#allocation_new"}`
	successBody := `{"result":"http://localhost:8080/mockid"}`

	t.Run("sends_gzip", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		zb := gzip.NewWriter(buf)
		_, err := zb.Write([]byte(requestBody))
		require.NoError(t, err)
		err = zb.Close()
		require.NoError(t, err)

		r := httptest.NewRequest("POST", srv.URL+"/api/shorten", buf)
		r.RequestURI = ""
		r.Header.Set("Content-Encoding", "gzip")
		r.Header.Set("Accept-Encoding", "")

		resp, err := http.DefaultClient.Do(r)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		defer resp.Body.Close()

		b, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.JSONEq(t, successBody, string(b))
	})

	t.Run("accepts_gzip", func(t *testing.T) {
		buf := bytes.NewBufferString(requestBody)
		r := httptest.NewRequest("POST", srv.URL+"/api/shorten", buf)
		r.RequestURI = ""
		r.Header.Set("Accept-Encoding", "gzip")

		resp, err := http.DefaultClient.Do(r)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		defer resp.Body.Close()

		zr, err := gzip.NewReader(resp.Body)
		require.NoError(t, err)

		b, err := io.ReadAll(zr)
		require.NoError(t, err)

		require.JSONEq(t, successBody, string(b))
	})
}
