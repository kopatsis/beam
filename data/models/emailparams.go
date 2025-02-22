package models

import "time"

type VerificationEmailParam struct {
	Param       string    `json:"p"`
	EmailAtTime string    `json:"e"`
	CustomerID  int       `json:"c"`
	Set         time.Time `json:"s"`
}

type SignInEmailParam struct {
	Param       string    `json:"p"`
	EmailAtTime string    `json:"e"`
	Set         time.Time `json:"s"`
}
