package listhelp

import (
	"net/url"
)

func ParseQueryParams(values url.Values) (string, bool) {
	sort := "updated_at"
	desc := true

	if s := values.Get("sort"); s == "created_at" || s == "updated_at" || s == "length" {
		sort = s
	}

	if d := values.Get("desc"); d == "true" || d == "false" {
		desc = d == "true"
	}

	return sort, desc
}
