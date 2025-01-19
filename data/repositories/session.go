package repositories

import (
	"beam/data/models"

	"gorm.io/gorm"
)

type SessionRepository interface {
	Create(session *models.Session) error
	Read(id string) (*models.Session, error)
	Update(session *models.Session) error
	Delete(id string) error
}

type sessionRepo struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) SessionRepository {
	return &sessionRepo{db: db}
}

func (r *sessionRepo) Create(session *models.Session) error {
	return r.db.Create(session).Error
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
	return r.db.Delete(id).Error
}
