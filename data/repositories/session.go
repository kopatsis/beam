package repositories

import (
	"beam/config"
	"beam/data/models"
	"log"
	"sync"
	"time"

	"gorm.io/gorm"
)

type SessionRepository interface {
	Create(session *models.Session) error
	Read(id string) (*models.Session, error)
	Update(session *models.Session) error
	Delete(id string) error
	SaveBatch(sessions []*models.Session, lines []*models.SessionLine) (error, error)
	AddToBatch(session *models.Session, line *models.SessionLine)
	FlushBatch()
}

type sessionRepo struct {
	db         *gorm.DB
	store      string
	sessionMu  sync.Mutex
	lineMu     sync.Mutex
	sessions   []*models.Session
	lines      []*models.SessionLine
	saveTicker *time.Ticker
}

func NewSessionRepository(db *gorm.DB, store string) SessionRepository {
	repo := &sessionRepo{
		db:         db,
		sessions:   make([]*models.Session, 0),
		lines:      make([]*models.SessionLine, 0),
		saveTicker: time.NewTicker(time.Duration(config.SESSIONBATCH) * time.Second),
		store:      store,
	}

	go func() {
		for range repo.saveTicker.C {
			repo.FlushBatch()
		}
	}()
	defer repo.saveTicker.Stop()

	return repo
}

func (r *sessionRepo) Create(session *models.Session) error {
	r.AddToBatch(session, nil)
	return nil
}

func (r *sessionRepo) Read(id string) (*models.Session, error) {
	var session models.Session
	err := r.db.First(&session, id).Error
	return &session, err
}

func (r *sessionRepo) Update(session *models.Session) error {
	return r.db.Save(session).Error
}

func (r *sessionRepo) Delete(id string) error {
	return r.db.Delete(&models.Session{}, id).Error
}

func (r *sessionRepo) AddToBatch(session *models.Session, line *models.SessionLine) {
	if session != nil {
		r.sessionMu.Lock()
		r.sessions = append(r.sessions, session)
		r.sessionMu.Unlock()
	}

	if line != nil {
		r.lineMu.Lock()
		r.lines = append(r.lines, line)
		r.lineMu.Unlock()
	}
}

func (r *sessionRepo) FlushBatch() {
	var errSessions, errLines error

	r.sessionMu.Lock()
	if len(r.sessions) > 0 {
		errSessions = r.db.Save(r.sessions).Error
		if errSessions == nil {
			r.sessions = nil
		}
	}
	r.sessionMu.Unlock()

	r.lineMu.Lock()
	if len(r.lines) > 0 {
		errLines = r.db.Save(r.lines).Error
		if errLines == nil {
			r.lines = nil
		}
	}
	r.lineMu.Unlock()

	if errSessions != nil {
		log.Printf("Unable to save the sessions in store %s here due to error: %v", r.store, errSessions)
	}

	if errLines != nil {
		log.Printf("Unable to save the session lines in store %s here due to error: %v", r.store, errLines)
	}
}

func (r *sessionRepo) SaveBatch(sessions []*models.Session, lines []*models.SessionLine) (error, error) {
	errSessions := r.db.Save(sessions).Error
	errLines := r.db.Save(lines).Error
	return errSessions, errLines
}
