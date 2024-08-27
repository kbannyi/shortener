package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"

	"github.com/go-chi/chi"
	"github.com/kbannyi/shortener/internal/auth"
	"github.com/kbannyi/shortener/internal/config"
	"github.com/kbannyi/shortener/internal/domain"
	"github.com/kbannyi/shortener/internal/dto"
	"github.com/kbannyi/shortener/internal/logger"
	"github.com/kbannyi/shortener/internal/middleware"
	"github.com/kbannyi/shortener/internal/models"
	"github.com/kbannyi/shortener/internal/repository"
)

type URLHandler struct {
	Service Service
	Flags   config.Flags
}

type Service interface {
	Create(ctx context.Context, value string) (ID string, err error)
	Get(ID string) (string, error)
	GetByUser(ctx context.Context) ([]*domain.URL, error)
	DeleteByUser(ctx context.Context, ids []string) error
	BatchCreate(ctx context.Context, correlated []models.CorrelatedURL) (map[string]*domain.URL, error)
}

func NewURLHandler(s Service, c config.Flags) http.Handler {
	r := chi.NewRouter()
	h := URLHandler{s, c}

	// Public
	r.Group(func(r chi.Router) {
		r.Get("/{id}", h.getByID)
	})

	// Auto auth
	r.Group(func(r chi.Router) {
		r.Use(middleware.AutoAuthMiddleware)

		r.Post("/", h.createFromText)
		r.Post("/api/shorten", h.create)
		r.Post("/api/shorten/batch", h.batchCreate)
	})

	// Require auth
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuthMiddleware)

		r.Get("/api/user/urls", h.getByUser)
		r.Delete("/api/user/urls", h.deleteByUser)
	})

	return r
}

func (handler *URLHandler) createFromText(w http.ResponseWriter, r *http.Request) {
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

	linkid, err := handler.Service.Create(r.Context(), link)
	if err != nil {
		var dupErr *repository.ErrDuplicateURL
		if errors.As(err, &dupErr) {
			w.WriteHeader(http.StatusConflict)
			handler.writeCreateFromTextResult(w, dupErr.URL.ID)
			return
		}

		logger.Log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	handler.writeCreateFromTextResult(w, linkid)
}

func (handler *URLHandler) writeCreateFromTextResult(w http.ResponseWriter, linkid string) {
	shorturl, err := url.JoinPath(handler.Flags.RedirectBaseAddr, linkid)
	if err != nil {
		logger.Log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = io.WriteString(w, shorturl)
	if err != nil {
		logger.Log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (handler *URLHandler) getByID(w http.ResponseWriter, r *http.Request) {
	linkid := chi.URLParam(r, "id")

	if len(linkid) == 0 {
		http.Error(w, "Link id can't be empty", http.StatusBadRequest)
		return
	}

	link, err := handler.Service.Get(linkid)
	if errors.Is(err, repository.ErrNotFound) {
		http.Error(w, "link by id not found", http.StatusBadRequest)
		return
	}
	if errors.Is(err, repository.ErrDeleted) {
		w.WriteHeader(http.StatusGone)
		return
	}
	if err != nil {
		logger.Log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, link, http.StatusTemporaryRedirect)
}

func (handler *URLHandler) create(w http.ResponseWriter, r *http.Request) {
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

	linkid, err := handler.Service.Create(r.Context(), reqmodel.URL)
	if err != nil {
		var dupErr *repository.ErrDuplicateURL
		if errors.As(err, &dupErr) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			handler.writeCreateResult(w, dupErr.URL.ID)
			return
		}

		logger.Log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	handler.writeCreateResult(w, linkid)
}

func (handler *URLHandler) writeCreateResult(w http.ResponseWriter, linkid string) {
	shorturl, err := url.JoinPath(handler.Flags.RedirectBaseAddr, linkid)
	if err != nil {
		logger.Log.Error(err.Error())
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

func (handler *URLHandler) batchCreate(w http.ResponseWriter, r *http.Request) {
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

	urls, err := handler.Service.BatchCreate(r.Context(), correlated)
	if err != nil {
		logger.Log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	response := make([]dto.BatchResponseURL, 0, len(correlated))
	for _, orig := range request {
		shorturl, err := url.JoinPath(handler.Flags.RedirectBaseAddr, urls[orig.CorrelationID].Short)
		if err != nil {
			logger.Log.Error(err.Error())
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

func (handler *URLHandler) getByUser(w http.ResponseWriter, r *http.Request) {
	urls, err := handler.Service.GetByUser(r.Context())
	if err != nil {
		logger.Log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(urls) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	response := make([]dto.UserResponseURL, 0, len(urls))
	for _, u := range urls {
		shorturl, err := url.JoinPath(handler.Flags.RedirectBaseAddr, u.Short)
		if err != nil {
			logger.Log.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		response = append(response, dto.UserResponseURL{
			OriginalURL: u.Original,
			ShortURL:    shorturl,
		})
	}
	encoder := json.NewEncoder(w)
	w.Header().Set("Content-Type", "application/json")
	if err := encoder.Encode(&response); err != nil {
		logger.Log.Errorf("Response write failed: %v", err)
		return
	}
}

func (handler *URLHandler) deleteByUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var request []string
	if err := decoder.Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if len(request) == 0 {
		http.Error(w, "request is empty", http.StatusBadRequest)
		return
	}

	err := handler.Service.DeleteByUser(r.Context(), request)
	if errors.Is(err, repository.ErrNotFound) {
		http.Error(w, "link by id not found", http.StatusBadRequest)
		return
	}
	if errors.Is(err, auth.ErrNotAuthorized) {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	if err != nil {
		logger.Log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusAccepted)
}
