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
	SixDigitCode []uint    `json:"x"`
	HasCustomer  bool      `json:"h"`
	CustomerID   int       `json:"c"`
	Tries        int       `json:"r"`
	NewCodeReqs  int       `json:"n"`
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
	SixDigitCode []uint    `json:"x"`
	Tries        int       `json:"r"`
	NewCodeReqs  int       `json:"n"`
}

type LoginSpecificParams struct {
	Param        string    `json:"p"`
	ReturnHandle string    `json:"r"`
	DraftID      string    `json:"d"`
	OrderID      string    `json:"o"`
	CartID       int       `json:"c"`
	Date         time.Time `json:"t"`
}
