package router

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/go-chi/chi"
	"github.com/kbannyi/shortener/internal/config"
	"github.com/kbannyi/shortener/internal/dto"
	"github.com/kbannyi/shortener/internal/logger"
)

type URLRouter struct {
	chi.Router
	Service Service
	Flags   config.Flags
}

type Service interface {
	Create(value string) (ID string, err error)
	Get(ID string) (string, bool)
}

func NewURLRouter(s Service, c config.Flags) *URLRouter {
	r := URLRouter{chi.NewRouter(), s, c}

	r.Get("/{id}", r.getByID)
	r.Post("/", r.createFromText)

	// JSON-based:
	r.Post("/api/shorten", r.create)

	return &r
}

func (router *URLRouter) createFromText(w http.ResponseWriter, r *http.Request) {
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

	linkid, err := router.Service.Create(link)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	shorturl, err := url.JoinPath(router.Flags.RedirectBaseAddr, linkid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	_, err = io.WriteString(w, shorturl)
	if err != nil {
		logger.Log.Errorf("Response write failed: %v", err)
		return
	}
}

func (router *URLRouter) getByID(w http.ResponseWriter, r *http.Request) {
	linkid := chi.URLParam(r, "id")

	if len(linkid) == 0 {
		http.Error(w, "Link id can't be empty", http.StatusBadRequest)
		return
	}

	link, ok := router.Service.Get(linkid)
	if !ok {
		http.Error(w, "Unknown link", http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, link, http.StatusTemporaryRedirect)
}

func (router *URLRouter) create(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var reqmodel dto.ShortenRequest
	if err := decoder.Decode(&reqmodel); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if reqmodel.URL == "" {
		http.Error(w, "url is required", http.StatusBadRequest)
		return
	}

	linkid, err := router.Service.Create(reqmodel.URL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	shorturl, err := url.JoinPath(router.Flags.RedirectBaseAddr, linkid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	encoder := json.NewEncoder(w)
	resmodel := dto.ShortenResponse{Result: shorturl}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := encoder.Encode(&resmodel); err != nil {
		logger.Log.Errorf("Response write failed: %v", err)
		return
	}
}
