package models

import "time"

type VerificationEmailParam struct {
	Param       string    `json:"p"`
	EmailAtTime string    `json:"e"`
	CustomerID  int       `json:"c"`
	Set         time.Time `json:"s"`
}

type SignInEmailParam struct {
	Param        string    `json:"p"`
	EmailAtTime  string    `json:"e"`
	Set          time.Time `json:"s"`
	SixDigitCode uint      `json:"x"`
	HasCustomer  bool      `json:"h"`
	CustomerID   int       `json:"c"`
	Tries        int       `json:"r"`
}

type ResetEmailParam struct {
	Param       string    `json:"p"`
	EmailAtTime string    `json:"e"`
	CustomerID  int       `json:"c"`
	SecretCode  string    `json:"r"`
	Set         time.Time `json:"s"`
}

type TwoFactorEmailParam struct {
	Param        string    `json:"p"`
	CustomerID   int       `json:"c"`
	Set          time.Time `json:"s"`
	SixDigitCode uint      `json:"x"`
	Tries        int       `json:"r"`
}
