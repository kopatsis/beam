package models

import (
	"time"

	"github.com/lib/pq"
)

type Review struct {
	PK          int       `gorm:"primaryKey"`
	CustomerID  int       `gorm:"index"`
	ProductID   int       `gorm:"index"`
	CreatedAt   time.Time `gorm:"index"`
	Status      string    `gorm:"index"`
	DisplayName string
	LastEdited  time.Time
	Stars       int
	JustStar    bool
	Subject     string
	Body        string
	ImageURLs   pq.StringArray
	HelpfulTr   pq.Int64Array
	UnhelpfulTr pq.Int64Array
	Helpful     int
	Unhelpful   int
}

func (r *Review) CheckCust(customerID int) int {
	customerID64 := int64(customerID)

	inHelpful := false
	inUnhelpful := false

	for _, v := range r.HelpfulTr {
		if v == customerID64 {
			inHelpful = true
			break
		}
	}

	for _, v := range r.UnhelpfulTr {
		if v == customerID64 {
			inUnhelpful = true
			break
		}
	}

	if inHelpful && inUnhelpful {
		return 0
	} else if inHelpful {
		return 1
	} else if inUnhelpful {
		return -1
	}

	return 0
}

func (r *Review) SetCust(customerID int, helpful bool) {
	customerID64 := int64(customerID)
	if helpful {
		if contains(r.UnhelpfulTr, customerID64) {
			r.UnhelpfulTr = remove(r.UnhelpfulTr, customerID64)
		}
		r.HelpfulTr = append(r.HelpfulTr, customerID64)
	} else {
		if contains(r.HelpfulTr, customerID64) {
			r.HelpfulTr = remove(r.HelpfulTr, customerID64)
		}
		r.UnhelpfulTr = append(r.UnhelpfulTr, customerID64)
	}
}

func (r *Review) UnsetCust(customerID int) {
	customerID64 := int64(customerID)
	r.HelpfulTr = remove(r.HelpfulTr, customerID64)
	r.UnhelpfulTr = remove(r.UnhelpfulTr, customerID64)
}

func contains(arr pq.Int64Array, val int64) bool {
	for _, v := range arr {
		if v == val {
			return true
		}
	}
	return false
}

func remove(arr pq.Int64Array, val int64) pq.Int64Array {
	var result pq.Int64Array
	for _, v := range arr {
		if v != val {
			result = append(result, v)
		}
	}
	return result
}
