package router

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"

	"github.com/go-chi/chi"
	"github.com/kbannyi/shortener/internal/config"
	"github.com/kbannyi/shortener/internal/domain"
	"github.com/kbannyi/shortener/internal/dto"
	"github.com/kbannyi/shortener/internal/logger"
	"github.com/kbannyi/shortener/internal/models"
	"github.com/kbannyi/shortener/internal/repository"
)

type URLRouter struct {
	chi.Router
	Service Service
	Flags   config.Flags
}

type Service interface {
	Create(value string) (ID string, err error)
	Get(ID string) (string, bool)
	BatchCreate(ctx context.Context, correlated []models.CorrelatedURL) (map[string]*domain.URL, error)
}

func NewURLRouter(s Service, c config.Flags) *URLRouter {
	r := URLRouter{chi.NewRouter(), s, c}

	r.Get("/{id}", r.getByID)
	r.Post("/", r.createFromText)

	// JSON-based:
	r.Post("/api/shorten", r.create)
	r.Post("/api/shorten/batch", r.batchCreate)

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
		var dupErr *repository.DuplicateURLError
		if errors.As(err, &dupErr) {
			w.WriteHeader(http.StatusConflict)
			router.writeCreateFromTextResult(w, dupErr.URL.ID)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	router.writeCreateFromTextResult(w, linkid)
}

func (router *URLRouter) writeCreateFromTextResult(w http.ResponseWriter, linkid string) {
	shorturl, err := url.JoinPath(router.Flags.RedirectBaseAddr, linkid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	io.WriteString(w, shorturl)
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
		var dupErr *repository.DuplicateURLError
		if errors.As(err, &dupErr) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			router.writeCreateResult(w, dupErr.URL.ID)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	router.writeCreateResult(w, linkid)
}

func (router *URLRouter) writeCreateResult(w http.ResponseWriter, linkid string) {
	shorturl, err := url.JoinPath(router.Flags.RedirectBaseAddr, linkid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	encoder := json.NewEncoder(w)
	resmodel := dto.ShortenResponse{Result: shorturl}
	if err := encoder.Encode(&resmodel); err != nil {
		logger.Log.Errorf("Response write failed: %v", err)
		return
	}
}

func (router *URLRouter) batchCreate(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var request []dto.BatchRequestURL
	if err := decoder.Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if len(request) == 0 {
		http.Error(w, "request is empty", http.StatusBadRequest)
		return
	}

	correlated := make([]models.CorrelatedURL, 0, len(request))
	for _, url := range request {
		if url.CorrelationID == "" || url.OriginalURL == "" {
			http.Error(w, "correlation_id and original_id are required", http.StatusBadRequest)
			return
		}
		correlated = append(correlated, models.CorrelatedURL{
			CorrelationID: url.CorrelationID,
			Value:         url.OriginalURL,
		})
	}

	urls, err := router.Service.BatchCreate(r.Context(), correlated)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	response := make([]dto.BatchResponseURL, 0, len(correlated))
	for _, orig := range request {
		shorturl, err := url.JoinPath(router.Flags.RedirectBaseAddr, urls[orig.CorrelationID].Short)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		response = append(response, dto.BatchResponseURL{
			CorrelationID: orig.CorrelationID,
			ShortURL:      shorturl,
		})
	}
	encoder := json.NewEncoder(w)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := encoder.Encode(&response); err != nil {
		logger.Log.Errorf("Response write failed: %v", err)
		return
	}
}
