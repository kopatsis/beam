package discount

import (
	"beam/config"
	"strconv"
)

// Welcome DISC name, Always up disc name
func SpecialDiscNames(storeSettings *config.SettingsMutex, store string) (string, float64, string, float64) {
	welcome := config.BASE_WELCOME_CODE
	always := config.BASE_ALWAYS_CODE

	welcomePct := config.DEFAULT_WELCOME_PCT
	alwaysPct := config.DEFAULT_ALWAYS_PCT

	storeSettings.Mu.RLock()
	pct, ok := storeSettings.Settings.WelcomePct[store]
	if ok {
		welcomePct = pct
	}
	welcome += strconv.Itoa(welcomePct)

	pct, ok = storeSettings.Settings.AlwaysWorksPct[store]
	if ok {
		alwaysPct = pct
	}
	always += strconv.Itoa(alwaysPct)

	storeSettings.Mu.RUnlock()

	return welcome, float64(welcomePct) / 100, always, float64(alwaysPct) / 100
}
