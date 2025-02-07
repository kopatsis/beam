package repositories

import (
	"beam/config"
	"beam/data/models"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
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
	GetAffiliate(code string) (models.AffiliateSession, error)
	AddAffiliateLine(line *models.AffiliateLine, store string)
	AddAffiliateSale(line *models.AffiliateSale, store string)
}

type sessionRepo struct {
	db         *gorm.DB
	store      string
	sessionMu  sync.Mutex
	lineMu     sync.Mutex
	sessions   []models.Session
	lines      []models.SessionLine
	saveTicker *time.Ticker
}

func NewSessionRepository(db *gorm.DB, store string) SessionRepository {
	repo := &sessionRepo{
		db:         db,
		sessions:   make([]models.Session, 0),
		lines:      make([]models.SessionLine, 0),
		saveTicker: time.NewTicker(time.Duration(config.BATCH) * time.Second),
		store:      store,
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		defer repo.saveTicker.Stop()
		defer repo.FlushBatch()

		for {
			select {
			case <-repo.saveTicker.C:
				repo.FlushBatch()
			case <-sigChan:
				return
			}
		}
	}()

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
		r.sessions = append(r.sessions, *session)
		r.sessionMu.Unlock()
	}

	if line != nil {
		r.lineMu.Lock()
		r.lines = append(r.lines, *line)
		r.lineMu.Unlock()
	}
}

func (r *sessionRepo) FlushBatch() {
	r.sessionMu.Lock()
	sessionsToSave := append([]models.Session{}, r.sessions...)
	r.sessions = r.sessions[:0]
	r.sessionMu.Unlock()

	r.lineMu.Lock()
	linesToSave := append([]models.SessionLine{}, r.lines...)
	r.lines = r.lines[:0]
	r.lineMu.Unlock()

	if err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(sessionsToSave).Error; err != nil {
			return err
		}
		if err := tx.Save(linesToSave).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.Printf("Unable to save sessions and lines in store %s due to error: %v", r.store, err)

		r.sessionMu.Lock()
		r.sessions = append(r.sessions, sessionsToSave...)
		r.sessionMu.Unlock()

		r.lineMu.Lock()
		r.lines = append(r.lines, linesToSave...)
		r.lineMu.Unlock()
	}
}

func (r *sessionRepo) SaveBatch(sessions []*models.Session, lines []*models.SessionLine) (error, error) {
	errSessions := r.db.Save(sessions).Error
	errLines := r.db.Save(lines).Error
	return errSessions, errLines
}

func (r *sessionRepo) GetAffiliate(code string) (models.AffiliateSession, error) {
	var aff models.Affiliate
	if err := r.db.Where("code = ? AND valid = true", code).First(&aff).Error; err != nil {
		return models.AffiliateSession{}, err
	}

	session := models.AffiliateSession{ID: aff.ID, ActualCode: aff.Code}

	go func() {
		r.db.Model(&models.Affiliate{}).Where("id = ?", aff.ID).Update("last_used", time.Now())
	}()

	return session, nil
}

func (r *sessionRepo) AddAffiliateLine(line *models.AffiliateLine, store string) {

	if line == nil {
		log.Printf("Nil affiliate line attempted to save.\n")
		return
	}

	if err := r.db.Save(line).Error; err != nil {
		log.Printf("Unable to save affiliate line, error: %v; Store: %s; ID: %d; Code: %s; SessionID: %s\n", err, store, line.AffiliateID, line.Code, line.SessionID)
	}
}

func (r *sessionRepo) AddAffiliateSale(line *models.AffiliateSale, store string) {
	if line == nil {
		log.Printf("Nil affiliate sale attempted to save.\n")
		return
	}

	if err := r.db.Save(line).Error; err != nil {
		log.Printf("Unable to save affiliate line, error: %v; Store: %s; ID: %d; Code: %s; SessionID: %s; Order ID: %s\n", err, store, line.AffiliateID, line.Code, line.SessionID, line.OrderID)
	}
}
