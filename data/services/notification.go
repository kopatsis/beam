package services

import (
	"beam/data/models"
	"beam/data/repositories"
)

type NotificationService interface {
	AddNotification(notification models.Notification) error
	GetNotificationByID(id string) (*models.Notification, error)
	UpdateNotification(notification models.Notification) error
	DeleteNotification(id string) error
}

type notificationService struct {
	notificationRepo repositories.NotificationRepository
}

func NewNotificationService(notificationRepo repositories.NotificationRepository) NotificationService {
	return &notificationService{notificationRepo: notificationRepo}
}

func (s *notificationService) AddNotification(notification models.Notification) error {
	return s.notificationRepo.Create(notification)
}

func (s *notificationService) GetNotificationByID(id string) (*models.Notification, error) {
	return s.notificationRepo.Read(id)
}

func (s *notificationService) UpdateNotification(notification models.Notification) error {
	return s.notificationRepo.Update(notification)
}

func (s *notificationService) DeleteNotification(id string) error {
	return s.notificationRepo.Delete(id)
}
