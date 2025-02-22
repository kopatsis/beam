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
	DeviceCookie string    `json:"d"`
	Set          time.Time `json:"s"`
	EmailSubbed  bool      `json:"u"`
}

type TwoFactorEmailParam struct {
	Param        string    `json:"t"`
	CustomerID   int       `json:"c"`
	Set          time.Time `json:"s"`
	SixDigitCode uint      `json:"x"`
	Tries        int       `json:"r"`
}
