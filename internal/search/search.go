package search

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/su55y/yt_search_rofi_blocks/internal/blocks"
	"github.com/su55y/yt_search_rofi_blocks/internal/config"
	"google.golang.org/api/youtube/v3"
)

type SearchService struct {
	service      *youtube.Service
	appConf      config.AppConfig
	currentQuery string
	appCachePath string
}

func NewSearchService(s *youtube.Service, c config.AppConfig, p string) *SearchService {
	return &SearchService{
		service:      s,
		appConf:      c,
		appCachePath: p,
	}
}

func (s *SearchService) NewQuery(q string) {
	if len(q) > 0 {
		s.currentQuery = q
	}
}

func (s *SearchService) DoSearch(page string) (blocks.Blocks, blocks.Page) {
	call := s.service.Search.List([]string{"snippet"}).
		RegionCode(s.appConf.Region).
		MaxResults(s.appConf.MaxResults).
		Type("video").
		Q(s.currentQuery)

	if len(page) == 6 {
		call = call.PageToken(page)
	}

	res, err := call.Do()
	if err != nil {
		log.Println("search error:", err.Error())
	}

	return blocks.Blocks{
			Lines:   s.getLines(res),
			Message: getMesg(res.PageInfo),
		}, blocks.Page{
			NextToken: res.NextPageToken,
			PrevToken: res.PrevPageToken,
		}
}

func getMesg(pageInfo *youtube.PageInfo) string {
	return fmt.Sprintf(
		"current page count: %d, total: %d",
		pageInfo.ResultsPerPage,
		pageInfo.TotalResults,
	)
}

func (s *SearchService) getLines(result *youtube.SearchListResponse) []blocks.Line {
	var lines []blocks.Line
	var err error
	for i, v := range result.Items {
		l := blocks.Line{
			Text: fmt.Sprintf("%d) %s", i, v.Snippet.Title),
			Data: fmt.Sprintf("%d:%s", i, v.Id.VideoId),
		}

		if !s.appConf.ThumbOff {
			thumbUrl := v.Snippet.Thumbnails.Default.Url
			switch s.appConf.ThumbSize {
			case "high":
				if len(v.Snippet.Thumbnails.High.Url) > 0 {
					thumbUrl = v.Snippet.Thumbnails.High.Url
				}
			case "medium":
				if len(v.Snippet.Thumbnails.Medium.Url) > 0 {
					thumbUrl = v.Snippet.Thumbnails.Medium.Url
				}
			}

			if l.Icon, err = s.getThumb(v.Id.VideoId, thumbUrl); err != nil {
				log.Println(err.Error())
			}
		}

		lines = append(lines, l)
	}

	if len(result.NextPageToken) == 6 {
		lines = append(
			lines,
			blocks.Line{
				Text: "Next >",
				Data: "next:" + result.NextPageToken,
			},
		)
	}
	if len(result.PrevPageToken) == 6 {
		lines = append(
			lines,
			blocks.Line{
				Text: "< Prev",
				Data: "prev:" + result.PrevPageToken,
			},
		)
	}

	lines = append(
		lines,
		blocks.Line{
			Text: "Clear ",
			Data: "cmd:clear",
		},
	)

	return lines
}

func (s *SearchService) getThumb(id, url string) (string, error) {
	ext := filepath.Ext(filepath.Base(url))
	if len(ext) == 0 && len(filepath.Base(url)) != 0 {
		ext = ".jpg"
	}
	if len(filepath.Base(url)) == 0 {
		return "", fmt.Errorf("can't read file from url:'%s'", url)
	}

	thumbName := "t" + id + ext
	switch s.appConf.ThumbSize {
	case "high":
		thumbName = "h" + thumbName
	case "medium":
		thumbName = "m" + thumbName
	default:
		thumbName = "d" + thumbName
	}

	thumbPath := filepath.Join(s.appCachePath, thumbName)
	if _, err := os.Stat(thumbPath); err == nil {
		return thumbPath, nil
	}

	out, err := os.Create(thumbPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	log.Printf("download thumb: %s\n", url)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("get thumb status: %s", resp.Status)
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", err
	}

	return thumbPath, nil
}
