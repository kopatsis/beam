package config

import (
	"net"
	"strings"
)

func GetLocation(ipStr string, tools *Tools) (string, string) {

	if ipStr == "" || tools.Geo == nil {
		return "", ""
	}

	var city string
	var country string

	if commaIndex := strings.Index(ipStr, ","); commaIndex != -1 {
		ipStr = ipStr[:commaIndex]
	}

	ip := net.ParseIP(ipStr)
	if ip != nil {
		record, err := tools.Geo.City(ip)
		if err == nil && record != nil {
			city = record.City.Names["en"]
			country = record.Country.Names["en"]
		}
	}

	return city, country
}
