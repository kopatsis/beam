package custhelp

import (
	"beam/config"
	"fmt"
	"time"
)

// Valid for device, valid for ip, valid for account, error
func RateLimitLogin(store, deviceID, ip string, customerID int, tools *config.Tools) (bool, bool, bool, error) {
	if unmaxed, err := config.RateLimit(tools.Redis, store, "LGNS", fmt.Sprintf("%d", customerID), config.LOGIN_ATTEMPTS_ACCOUNT, time.Second); err != nil {
		return true, true, true, err
	} else if !unmaxed {
		return true, true, false, nil
	}

	if unmaxed, err := config.RateLimit(tools.Redis, store, "LGNM", fmt.Sprintf("%d", customerID), config.LOGIN_ATTEMPTS_MINUTE, time.Minute); err != nil {
		return true, true, true, err
	} else if !unmaxed {
		return true, false, true, nil
	}

	if unmaxed, err := config.RateLimit(tools.Redis, store, "LGNH", fmt.Sprintf("%d", customerID), config.LOGIN_ATTEMPTS_HOUR, time.Hour); err != nil {
		return true, true, true, err
	} else if !unmaxed {
		return true, false, true, nil
	}

	return true, true, true, nil
}
