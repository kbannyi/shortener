package url

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/kbannyi/shortener/internal/service/url"
)

type URLHandler struct {
	service url.URLService
}

func NewHandler() *URLHandler {
	return &URLHandler{
		service: *url.NewService(),
	}
}

func (h *URLHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleGet(w, r)
	case http.MethodPost:
		h.handlePost(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusBadRequest)
		return
	}
}

func (h *URLHandler) handlePost(w http.ResponseWriter, r *http.Request) {
	linkBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Couldn't read body", http.StatusBadRequest)
		return
	}
	link := string(linkBytes)
	if len(link) == 0 {
		http.Error(w, "Link string can't be empty", http.StatusBadRequest)
		return
	}

	fmt.Printf("Shortening link %v\n", link)
	w.WriteHeader(http.StatusCreated)
	linkid := h.service.Create(link)
	_, err = io.WriteString(w, fmt.Sprintf("http://localhost:8080/%v", linkid))
	if err != nil {
		panic(err)
	}
}

func (h *URLHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	linkid := strings.TrimPrefix(r.URL.String(), "/")
	fmt.Printf("Getting link %v\n", linkid)

	if len(linkid) == 0 {
		http.Error(w, "Link id can't be empty", http.StatusBadRequest)
		return
	}

	link, ok := h.service.Get(linkid)
	if !ok {
		http.Error(w, "Unknown link", http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, link, http.StatusTemporaryRedirect)
}
