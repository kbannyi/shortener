package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/kbannyi/shortener/internal/auth"
	"github.com/kbannyi/shortener/internal/domain"
	"github.com/kbannyi/shortener/internal/logger"
	"github.com/kbannyi/shortener/internal/models"
	"go.uber.org/zap"
)

type Repository interface {
	Save(ctx context.Context, url *domain.URL) error
	BatchSave(ctx context.Context, urls []*domain.URL) error
	Get(ctx context.Context, id string) (*domain.URL, error)
	GetByUser(ctx context.Context, id string) ([]*domain.URL, error)
	GetList(ctx context.Context, ids []string) ([]*domain.URL, error)
	DeleteIDs(ctx context.Context, ids []string) error
}

type URLService struct {
	Repository Repository
	delChan    chan string
	ctx        context.Context
}

func NewService(ctx context.Context, r Repository) *URLService {
	instance := &URLService{
		r,
		make(chan string, 100),
		ctx,
	}

	go instance.flushBatchDelete()

	return instance
}

func (s *URLService) Create(ctx context.Context, value string) (ID string, err error) {
	u, err := auth.FromContext(ctx)
	if err != nil {
		return "", err
	}

	URL := domain.NewURLUser(value, u.UserID)
	if err := s.Repository.Save(ctx, URL); err != nil {
		return "", fmt.Errorf("coudln't save domain.URL: %w", err)
	}

	return URL.ID, nil
}

func (s *URLService) Get(ID string) (string, error) {
	v, err := s.Repository.Get(context.TODO(), ID)
	if err != nil {
		return "", err
	}

	return v.Original, nil
}

func (s *URLService) BatchCreate(ctx context.Context, correlated []models.CorrelatedURL) (map[string]*domain.URL, error) {
	u, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	results := make(map[string]*domain.URL, len(correlated))
	batch := make([]*domain.URL, 0, len(correlated))
	for _, orig := range correlated {
		_, ok := results[orig.CorrelationID]
		if ok {
			return nil, errors.New("correlationId duplicate")
		}
		url := domain.NewURLUser(orig.Value, u.UserID)
		results[orig.CorrelationID] = url
		batch = append(batch, url)
	}
	err = s.Repository.BatchSave(ctx, batch)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func (s *URLService) GetByUser(ctx context.Context) ([]*domain.URL, error) {
	u, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	return s.Repository.GetByUser(ctx, u.UserID)
}

func (s *URLService) DeleteByUser(ctx context.Context, ids []string) error {
	u, err := auth.FromContext(ctx)
	if err != nil {
		return err
	}
	urls, err := s.Repository.GetList(ctx, ids)
	if err != nil {
		return err
	}

	for _, url := range urls {
		if url.UserID == nil || *url.UserID != u.UserID {
			return auth.ErrNotAuthorized
		}
	}
	for _, url := range urls {
		s.delChan <- url.ID
	}

	return nil
}

func (s *URLService) flushBatchDelete() {
	ticker := time.NewTicker(5 * time.Second)

	var ids []string

	for {
		select {
		case id := <-s.delChan:
			ids = append(ids, id)
		case <-ticker.C:
			if len(ids) == 0 {
				continue
			}
			logger.Log.Debugf("deleting urls: %+q", ids)
			err := s.Repository.DeleteIDs(s.ctx, ids)
			if err != nil {
				logger.Log.Error("cannot delete urls", zap.Error(err))
				continue
			}
			ids = nil
		case <-s.ctx.Done():
			logger.Log.Debug("flushBatchDelete stopped")
			return
		}
	}
}
