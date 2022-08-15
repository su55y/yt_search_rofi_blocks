package blocks_test

import (
	"testing"

	"github.com/su55y/yt_search_rofi_blocks/internal/blocks"
)

func TestParseSelect(t *testing.T) {
    ids := map[string]string{
        "video":"dQw4w9WgXcQ",
        "page":"a2b_c-",
    }
	cases := []struct {
        input string
        result blocks.Select
    }{
        { "cmd:clear", blocks.Select{Action:"clear"} },
        { "next:"+ids["page"], blocks.Select{Action:"next",Id:ids["page"]} },
        { "prev:"+ids["page"], blocks.Select{Action:"prev",Id:ids["page"]} },
        { "3:"+ids["video"], blocks.Select{Action:"open",Id:ids["video"],Selected:3} },
	}

	for _, c := range cases {
        r := blocks.ParseSelect(c.input)
        if r.Action != c.result.Action ||
            r.Id != c.result.Id ||
            r.Selected != c.result.Selected {
            t.Errorf("unexpected result ( %#+v ) in case ( %#+v )", r, c)
        }
	}
}
