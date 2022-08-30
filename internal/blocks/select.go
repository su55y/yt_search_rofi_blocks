package blocks

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	vidRx = regexp.MustCompile("^[0-9]{1,}\\:[a-zA-Z0-9_-]{11}$")
)

// parse "select entry" event
func ParseSelect(s string) Select {
	var sel Select
	if sub := strings.Split(s, ":"); len(sub) == 2 {
		if vidRx.MatchString(s) && len(sub[1]) == 11 {
			if d, err := strconv.Atoi(sub[0]); err == nil && d >= 0 {
				sel.Action = "open"
				sel.Id = sub[1]
				sel.Message = "open " + sub[1]
				sel.Selected = d
			}
		}
		return sel
	}

	return Select{
		Action:  "err",
		Message: "Something went wrong...",
	}
}
