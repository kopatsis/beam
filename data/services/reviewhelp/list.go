package reviewhelp

import (
	"net/url"
	"strconv"
)

func ParseQueryParams(values url.Values) (string, bool, int) {
	sort := "stars"
	desc := true
	page := 1

	if s := values.Get("sort"); s == "created_at" || s == "stars" {
		sort = s
	}

	if d := values.Get("desc"); d == "true" || d == "false" {
		desc = d == "true"
	}

	if p := values.Get("page"); p != "" {
		if num, err := strconv.Atoi(p); err == nil && num > 0 {
			page = num
		}
	}

	return sort, desc, page
}
