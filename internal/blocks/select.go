package blocks

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	vidRx  = regexp.MustCompile("^[0-9]{1,}\\:[a-zA-Z0-9_-]{11}$")
	pageRx = regexp.MustCompile("^(next|prev)\\:[a-zA-Z0-9-_]{6}$")
)

// parse "select entry" event
func ParseSelect(s string) Select {
	var sel Select
	if sub := strings.Split(s, ":"); len(sub) == 2 {
		switch sub[0] {
		case "cmd":
			if sub[1] == "clear" {
				sel.Action = sub[1]
				sel.Message = "enter for search"
			}
		case "next", "prev":
			if pageRx.MatchString(s) && len(sub[1]) == 6 {
				sel.Action = sub[0]
				sel.Id = sub[1]
			}
		default:
			if vidRx.MatchString(s) && len(sub[1]) == 11 {
				if d, err := strconv.Atoi(sub[0]); err == nil && d >= 0 {
					sel.Action = "open"
					sel.Id = sub[1]
					sel.Message = "open " + sub[1]
					sel.Selected = d
				}
			}
		}
		return sel
	}

	return Select{
		Action:  "err",
		Message: "Something went wrong...",
	}
}
