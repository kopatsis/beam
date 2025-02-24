package custhelp

import (
	"beam/config"
	"context"
	"fmt"
	"time"
)

// Is usage issue/failure + Failure type = hour, minute, second, ip, device + any error
func RateLimitLoginPtl(store, deviceID, ip, email string, tools *config.Tools) (bool, string, error) {
	if unmaxed, err := config.RateLimit(tools.Redis, store, "LGNS", email, config.LOGIN_ATTEMPTS_ACCOUNT, time.Second); err != nil {
		return false, "", err
	} else if !unmaxed {
		return true, "hour", nil
	}

	if unmaxed, err := config.RateLimit(tools.Redis, store, "LGNM", email, config.LOGIN_ATTEMPTS_MINUTE, time.Minute); err != nil {
		return false, "", err
	} else if !unmaxed {
		return true, "hour", nil
	}

	if unmaxed, err := config.RateLimit(tools.Redis, store, "LGNH", email, config.LOGIN_ATTEMPTS_HOUR, time.Hour); err != nil {
		return false, "", err
	} else if !unmaxed {
		return true, "hour", nil
	}

	if unmaxed, err := config.RateLimit(tools.Redis, store, "LGNI", fmt.Sprintf("%s::%s", email, ip), config.LOGIN_ATTEMPTS_IP, time.Second); err != nil {
		return false, "", err
	} else if !unmaxed {
		return true, "hour", nil
	}

	if unmaxed, err := config.RateLimit(tools.Redis, store, "LGND", fmt.Sprintf("%s::%s", email, deviceID), config.LOGIN_ATTEMPTS_DEVICE, time.Second); err != nil {
		return false, "", err
	} else if !unmaxed {
		return true, "hour", nil
	}

	return false, "", nil
}

func SetTempBanDeviceAndCustomer(store, deviceID string, email string, tools *config.Tools) error {
	key := fmt.Sprintf("%s::TPBD::%s::%s", store, email, deviceID)
	duration := config.LOCKOUT_MINUTES_DEVICE * time.Minute

	return tools.Redis.Set(context.Background(), key, "1", duration).Err()
}

func SetTempBanIPAndCustomer(store, ip string, email string, tools *config.Tools) error {
	key := fmt.Sprintf("%s::TPBI::%s::%s", store, email, ip)
	duration := config.LOCKOUT_MINUTES_IP * time.Minute

	return tools.Redis.Set(context.Background(), key, "1", duration).Err()
}

func SetTempBanCustomer(store, failurePoint string, email string, tools *config.Tools) error {
	key := fmt.Sprintf("%s::TPBC::%s", store, email)
	var duration time.Duration
	if failurePoint == "hour" {
		duration = config.LOCKOUT_MINUTES_HOUR * time.Minute
	} else if failurePoint == "minute" {
		duration = config.LOCKOUT_MINUTES_MINUTE * time.Minute
	} else {
		duration = config.LOCKOUT_MINUTES_ACCOUNT * time.Minute
	}

	return tools.Redis.Set(context.Background(), key, "1", duration).Err()
}

func SetTempBanFull(store, deviceID, ip, failurePoint string, email string, tools *config.Tools) error {
	if failurePoint == "device" {
		return SetTempBanDeviceAndCustomer(store, deviceID, email, tools)
	} else if failurePoint == "ip" {
		return SetTempBanIPAndCustomer(store, ip, email, tools)
	}
	return SetTempBanCustomer(store, failurePoint, email, tools)
}

func GetTempBanDeviceAndCustomer(store, deviceID string, email string, tools *config.Tools) (bool, error) {
	key := fmt.Sprintf("%s::TPBD::%s::%s", store, email, deviceID)

	exists, err := tools.Redis.Exists(context.Background(), key).Result()
	return exists > 0, err
}

func GetTempBanIPAndCustomer(store, ip string, email string, tools *config.Tools) (bool, error) {
	key := fmt.Sprintf("%s::TPBI::%s::%s", store, email, ip)

	exists, err := tools.Redis.Exists(context.Background(), key).Result()
	return exists > 0, err
}

func GetTempBanCustomer(store string, email string, tools *config.Tools) (bool, error) {
	key := fmt.Sprintf("%s::TPBC::%s", store, email)

	exists, err := tools.Redis.Exists(context.Background(), key).Result()
	return exists > 0, err
}

func GetTempBanFull(store, deviceID, ip string, email string, tools *config.Tools) (bool, error) {
	if banned, err := GetTempBanCustomer(store, email, tools); err != nil {
		return false, err
	} else if banned {
		return true, nil
	}

	if banned, err := GetTempBanIPAndCustomer(store, ip, email, tools); err != nil {
		return false, err
	} else if banned {
		return true, nil
	}

	if banned, err := GetTempBanDeviceAndCustomer(store, deviceID, email, tools); err != nil {
		return false, err
	} else if banned {
		return true, nil
	}

	return false, nil
}

func FullLoginRateLimiting(store, deviceID, ip string, email string, tools *config.Tools) (bool, error) {
	if banned, err := GetTempBanFull(store, deviceID, ip, email, tools); err != nil {
		return false, err
	} else if banned {
		return true, nil
	}

	banned, failurePoint, err := RateLimitLoginPtl(store, deviceID, ip, email, tools)
	if err != nil {
		return false, err
	}

	if banned {
		return true, SetTempBanFull(store, deviceID, ip, failurePoint, email, tools)
	}

	return false, nil
}
