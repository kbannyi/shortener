package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kbannyi/shortener/internal/config"
	"github.com/kbannyi/shortener/internal/domain"
	"github.com/kbannyi/shortener/internal/models"
	"github.com/stretchr/testify/assert"
)

type MockService struct{}

func (s *MockService) Create(value string) (ID string, err error) {
	return "mockid", nil
}

func (s *MockService) BatchCreate(ctx context.Context, urls []models.CorrelatedURL) (map[string]*domain.URL, error) {
	return map[string]*domain.URL{"1": {Short: "1"}, "2": {Short: "2"}}, nil
}

func (s *MockService) Get(ID string) (string, bool) {
	return "redirect", ID == "mockid"
}

func TestURLRouter(t *testing.T) {
	tests := []struct {
		method       string
		request      string
		body         string
		expectedCode int
		expectedBody string
	}{
		{
			method:       http.MethodPost,
			request:      "/",
			expectedCode: http.StatusBadRequest,
		},
		{
			method:       http.MethodGet,
			request:      "/unknownid",
			expectedCode: http.StatusBadRequest,
		},
		{
			method:       http.MethodGet,
			request:      "/mockid",
			expectedCode: http.StatusTemporaryRedirect,
		},
		{
			method:       http.MethodPost,
			request:      "/",
			body:         "ya.ru",
			expectedBody: "http://localhost:8080/mockid",
			expectedCode: http.StatusCreated,
		},
		{
			method:       http.MethodPost,
			request:      "/api/shorten",
			body:         `{"url": "https://go.dev/doc/effective_go#allocation_new"}`,
			expectedBody: `{"result":"http://localhost:8080/mockid"}`,
			expectedCode: http.StatusCreated,
		},
		{
			method:       http.MethodPost,
			request:      "/api/shorten/batch",
			body:         `[{"correlation_id": "1","original_url":"1"},{"correlation_id": "2","original_url":"2"}]`,
			expectedBody: `[{"correlation_id":"1","short_url":"http://localhost:8080/1"},{"correlation_id":"2","short_url":"http://localhost:8080/2"}]`,
			expectedCode: http.StatusCreated,
		},
		{
			method:       http.MethodPut,
			request:      "/",
			expectedCode: http.StatusMethodNotAllowed,
		},
		{
			method:       http.MethodDelete,
			request:      "/",
			expectedCode: http.StatusMethodNotAllowed,
		},
	}

	for _, tc := range tests {
		t.Run(tc.method, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, tc.request, strings.NewReader(tc.body))
			w := httptest.NewRecorder()
			router := NewURLHandler(&MockService{}, config.Flags{
				RedirectBaseAddr: "http://localhost:8080/",
			})

			router.ServeHTTP(w, r)
			assert.Equal(t, tc.expectedCode, w.Code)
			if tc.expectedBody != "" {
				assert.Equal(t, strings.Trim(tc.expectedBody, "\n"), strings.Trim(w.Body.String(), "\n"))
			}
		})
	}
}
