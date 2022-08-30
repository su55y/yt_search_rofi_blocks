package blocks

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/su55y/yt_search_rofi_blocks/internal/consts"
)

var (
	vidRx = regexp.MustCompile("^[0-9]{1,}\\:[a-zA-Z0-9_-]{11}$")
)

// parse "select entry" event
func ParseSelect(s string) Select {
	if sub := strings.Split(s, ":"); len(sub) == 2 &&
		len(sub[1]) == 11 &&
		vidRx.MatchString(s) {
		if d, err := strconv.Atoi(sub[0]); err == nil && d >= 0 {
			return Select{
				Action:   "open",
				Id:       sub[1],
				Message:  "open " + sub[1],
				Selected: d,
			}
		}
	}

	return Select{
		Action:  "err",
		Message: consts.INF_SELECT_PARSE,
	}
}
