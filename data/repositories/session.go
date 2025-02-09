package repositories

import (
	"beam/config"
	"beam/data/models"
	"context"
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type SessionRepository interface {
	Create(session *models.Session) error
	Read(id string) (*models.Session, error)
	Update(session *models.Session) error
	Delete(id string) error
	AddToBatch(session *models.Session, line *models.SessionLine)
	FlushBatch()
	GetAffiliate(code string) (models.AffiliateSession, error)
	AddAffiliateLine(line *models.AffiliateLine, store string)
	AddAffiliateSale(line *models.AffiliateSale, store string)
}

type sessionRepo struct {
	db         *gorm.DB
	rdb        *redis.Client
	store      string
	key        string
	saveTicker *time.Ticker
}

func NewSessionRepository(db *gorm.DB, rdb *redis.Client, store string, ct, len int) SessionRepository {
	repo := &sessionRepo{
		db:         db,
		rdb:        rdb,
		saveTicker: time.NewTicker(time.Duration(config.BATCH) * time.Second),
		store:      store,
	}

	go func() {
		defer repo.saveTicker.Stop()

		if len > 0 && ct >= 0 {
			delayFactor := float64(ct) / float64(len)
			if delayFactor > 1 {
				delayFactor = 1
			} else if delayFactor < 0 {
				delayFactor = 0
			}

			initialDelay := time.Duration(float64(config.BATCH) * delayFactor * float64(time.Second))
			time.Sleep(initialDelay)
		}

		for range repo.saveTicker.C {
			repo.FlushBatch()
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

	if session == nil && line == nil {
		return
	}

	fks, err := r.rdb.Get(context.Background(), r.store+"::SLN").Result()
	if err != nil {
		log.Printf("Unable to get key to push event to redis for store: %s; err: %v\n", r.store, err)
	}

	var flushKey config.FlushKey
	if err := json.Unmarshal([]byte(fks), &flushKey); err != nil {
		log.Printf("Unable to convert key to push event to redis for store: %s; err: %v\n", r.store, err)
	}

	if session != nil {
		data, err := json.Marshal(session)
		if err == nil {
			if err := r.rdb.LPush(context.Background(), r.store+"::SSN::"+flushKey.ActualKey, data); err != nil {
				log.Printf("Unable to push session to redis for store: %s; err: %v\n", r.store, err)
			}
		} else {
			log.Printf("Unable to create session for redis for store: %s; err: %v\n", r.store, err)
		}
	}

	if line != nil {
		data, err := json.Marshal(line)
		if err == nil {
			if err := r.rdb.LPush(context.Background(), r.store+"::SSL::"+flushKey.ActualKey, data); err != nil {
				log.Printf("Unable to push session line to redis for store: %s; err: %v\n", r.store, err)
			}
		} else {
			log.Printf("Unable to create session line for redis for store: %s; err: %v\n", r.store, err)
		}
	}
}

func (r *sessionRepo) FlushBatch() {
	oldKey := r.key
	r.key = strconv.FormatInt(time.Now().Unix(), 10)

	ctx := context.Background()

	sessionsData, err := r.rdb.LRange(ctx, r.store+"::SSN::"+oldKey, 0, -1).Result()
	if err != nil {
		log.Printf("Error retrieving sessions from Redis for store: %s; key: %s; error: %v", r.store, oldKey, err)
	} else {
		if err := r.rdb.Del(ctx, r.store+"::SSN::"+oldKey).Err(); err != nil {
			log.Printf("Error deleting sessions for store: %s; key: %s; error: %v", r.store, oldKey, err)
		}
	}

	linesData, err := r.rdb.LRange(ctx, r.store+"::SSL::"+oldKey, 0, -1).Result()
	if err != nil {
		log.Printf("Error retrieving lines for store: %s; key: %s; error: %v", r.store, oldKey, err)
	} else {
		if err := r.rdb.Del(ctx, r.store+"::SSL::"+oldKey).Err(); err != nil {
			log.Printf("Error deleting session lines for store: %s; key: %s; error: %v", r.store, oldKey, err)
		}
	}

	var sessionsToSave []models.Session
	var linesToSave []models.SessionLine

	if len(sessionsData) == 0 && len(linesData) == 0 {
		return
	}

	for _, session := range sessionsData {
		var s models.Session
		if err := json.Unmarshal([]byte(session), &s); err != nil {
			log.Printf("Error unmarshalling session for store: %s; key: %s; error: %v", r.store, oldKey, err)
		} else {
			sessionsToSave = append(sessionsToSave, s)
		}
	}

	for _, line := range linesData {
		var l models.SessionLine
		if err := json.Unmarshal([]byte(line), &l); err != nil {
			log.Printf("Error unmarshalling session line for store: %s; key: %s; error: %v", r.store, oldKey, err)
		} else {
			linesToSave = append(linesToSave, l)
		}
	}

	if len(sessionsToSave) == 0 && len(linesToSave) == 0 {
		return
	}

	if err := r.db.Transaction(func(tx *gorm.DB) error {
		if len(sessionsToSave) > 0 {
			if err := tx.Save(sessionsToSave).Error; err != nil {
				return err
			}
		}
		if len(linesToSave) > 0 {
			if err := tx.Save(linesToSave).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		log.Printf("Unable to save sessions and lines for store: %s; key: %s; error: %v", r.store, oldKey, err)
		if len(sessionsToSave) > 0 {
			if err := r.rdb.LPush(ctx, r.store+"::SSN::"+r.key, sessionsData).Err(); err != nil {
				log.Printf("Error pushing back sessions for store: %s; key: %s; error: %v", r.store, r.key, err)
			}
		}
		if len(linesToSave) > 0 {
			if err := r.rdb.LPush(ctx, r.store+"::SSL::"+r.key, linesData).Err(); err != nil {
				log.Printf("Error pushing back session lines for store: %s; key: %s; error: %v", r.store, r.key, err)
			}
		}
	}
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
