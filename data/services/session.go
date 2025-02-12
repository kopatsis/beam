package services

import (
	"beam/config"
	"beam/data/models"
	"beam/data/repositories"
	"beam/data/services/sessionhelp"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SessionService interface {
	AddSession(session *models.Session) error
	GetSessionByID(id string) (*models.Session, error)
	UpdateSession(session *models.Session) error
	DeleteSession(id string) error
	AddToSession(session *models.Session, line *models.SessionLine)
	SessionMiddleware(cookie *models.SessionCookie, customerID int, guestID, store string, c *gin.Context, tools *config.Tools)
	AffiliateMiddleware(cookie *models.AffiliateSession, sessionID, store string, c *gin.Context)

	AddAffiliateSale(dpi *DataPassIn, orderID string)
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

func (s *sessionService) SessionMiddleware(cookie *models.SessionCookie, customerID int, guestID, store string, c *gin.Context, tools *config.Tools) {
	if cookie == nil {
		cookie = &models.SessionCookie{
			Assigned: time.Now(),
		}
	}
	cookie.CustomerID = customerID

	if cookie.GuestID != guestID || cookie.Store != store || len(cookie.SessionID) < 2 || cookie.SessionID[:2] != "S-" {
		cookie.GuestID = guestID
		cookie.Store = store
		cookie.SessionID = "S-" + uuid.NewString()
		cookie.Assigned = time.Now()

		session := &models.Session{
			CustomerID: customerID,
			GuestID:    guestID,
			CreatedAt:  cookie.Assigned,
		}

		sessionhelp.CreateSessionDetails(c, tools, session)
		s.AddToSession(session, nil)
	}
}

func (s *sessionService) AffiliateMiddleware(cookie *models.AffiliateSession, sessionID, store string, c *gin.Context) {
	if cookie == nil {
		cookie = &models.AffiliateSession{}
	}

	affiliateCode := c.Query("affiliate")
	if affiliateCode == "" {
		return
	}

	if cookie.ID == 0 || cookie.ActualCode != affiliateCode {
		newCookie, err := s.sessionRepo.GetAffiliate(affiliateCode)
		if err != nil {
			log.Printf("Unable to query affiliate with ID: %s; Store: %s\n", affiliateCode, store)
			return
		}

		cookie.ID = newCookie.ID
		cookie.ActualCode = newCookie.ActualCode

		line := &models.AffiliateLine{
			AffiliateID: cookie.ID,
			Code:        cookie.ActualCode,
			SessionID:   sessionID,
			Timestamp:   time.Now(),
		}
		go func() {
			s.sessionRepo.AddAffiliateLine(line, store)
		}()
	}
}

func (s *sessionService) AddAffiliateSale(dpi *DataPassIn, orderID string) {
	if dpi.AffiliateID == 0 {
		return
	}

	use := models.AffiliateSale{
		AffiliateID: dpi.AffiliateID,
		Code:        dpi.AffiliateCode,
		SessionID:   dpi.SessionID,
		OrderID:     orderID,
		Timestamp:   time.Now(),
	}

	s.sessionRepo.AddAffiliateSale(&use, dpi.Store)
}
