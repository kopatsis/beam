package services

import (
	"beam/data/models"
	"beam/data/repositories"
)

type SessionService interface {
	AddSession(session *models.Session) error
	GetSessionByID(id string) (*models.Session, error)
	UpdateSession(session *models.Session) error
	DeleteSession(id string) error
	AddToSession(session *models.Session, line *models.SessionLine)
}

type sessionService struct {
	sessionRepo repositories.SessionRepository
}

func NewSessionService(sessionRepo repositories.SessionRepository) SessionService {
	return &sessionService{sessionRepo: sessionRepo}
}

func (s *sessionService) AddSession(session *models.Session) error {
	return s.sessionRepo.Create(session)
}

func (s *sessionService) GetSessionByID(id string) (*models.Session, error) {
	return s.sessionRepo.Read(id)
}

func (s *sessionService) UpdateSession(session *models.Session) error {
	return s.sessionRepo.Update(session)
}

func (s *sessionService) DeleteSession(id string) error {
	return s.sessionRepo.Delete(id)
}

func (s *sessionService) AddToSession(session *models.Session, line *models.SessionLine) {
	s.sessionRepo.AddToBatch(session, line)
}
