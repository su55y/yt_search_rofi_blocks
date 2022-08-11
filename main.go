package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/su55y/yt_search_rofi_blocks/internal/config"
	"github.com/su55y/yt_search_rofi_blocks/internal/consts"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

var (
	conf    Config
	appConf config.AppConfig

	rawIn string
	j_in  BlocksIn

	vidRx = regexp.MustCompile("^id\\:[0-9a-zA-Z_-]{11}$")
	npRx  = regexp.MustCompile("^(next|prev)\\:[a-zA-Z0-9-_]{6}$")
)

type Blocks struct {
	Massage string `json:"message"`
	Overlay string `json:"overlay"`
	Prompt  string `json:"prompt"`
	Input   string `json:"input"`
	Lines   []Line `json:"lines"`
	ActEntr int    `json:"active entry"`
}

type Line struct {
	Text   string `json:"text"`
	Markup bool   `json:"markup"`
	Icon   string `json:"icon"`
	Data   string `json:"data"`
}

type BlocksIn struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	Data  string `json:"data"`
}

type Config struct {
	AppCachePath  string
	HomePath      string
	NextPageToken string
	ConfPathRoot  string
	CachePathRoot string
	ConfFullPath  string

	Q string
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist) && err == nil
}

func readEnv() {
	// set /home/user/.config
	if conf.ConfPathRoot = os.Getenv(consts.ENV_CONFIG_HOME); !exists(conf.ConfPathRoot) {
		conf.ConfPathRoot = filepath.Join(conf.HomePath, consts.DEF_CONFIG_PATH)
	}

	// set /home/user/.cache
	if conf.CachePathRoot = os.Getenv(consts.ENV_CACHE_HOME); !exists(conf.CachePathRoot) {
		conf.CachePathRoot = filepath.Join(conf.HomePath, consts.DEF_CACHE_PATH)
	}
}

func getAppConfig() {
	appConfDirPath := filepath.Join(
		conf.ConfPathRoot,
		consts.APP_NAME,
	)

	appConfFilePath := filepath.Join(
		appConfDirPath,
		consts.APP_CONFIG_NAME,
	)

	if !exists(appConfDirPath) {
		if err := os.MkdirAll(appConfDirPath, os.ModePerm); err != nil {
			log.Fatal(err)
		}
	}

	if _, err := os.Stat(appConfFilePath); errors.Is(err, os.ErrNotExist) {
		log.Printf(consts.INF_NEW_CONFIG, appConfFilePath)
		ioutil.WriteFile(appConfFilePath, []byte(consts.DEF_CONFIG), 0666)
	}

	var err error
	appConf, err = config.GetAppConfig(appConfFilePath)
	if err != nil {
		log.Printf(consts.ERR_CONFIG_LOAD, err)
	}
}

func openInMPV(u string) {
	err := exec.Command("setsid", []string{"-f", "mpv", u}...).Start()
	if err != nil {
		log.Fatal(err)
	}
}

func init() {
	var err error
	if conf.HomePath, err = os.UserHomeDir(); err != nil {
		log.Fatal(err)
	}

	readEnv()
	getAppConfig()

	if len(appConf.API_KEY) == 0 {
		if exists(appConf.ApiKeyPath) {
			apiBytes, err := ioutil.ReadFile(appConf.ApiKeyPath)
			if err != nil {
				log.Fatal(fmt.Errorf(consts.ERR_NO_API_KEY_FILE, appConf.ApiKeyPath, err))
			}

			if appConf.API_KEY = strings.TrimSpace(string(apiBytes)); len(appConf.API_KEY) == 0 {
				log.Fatal(fmt.Errorf(consts.ERR_API_KEY_FILE_READ, appConf.ApiKeyPath))
			}
		} else {
			if appConf.API_KEY = os.Getenv(consts.ENV_YT_API_KEY); len(appConf.API_KEY) == 0 {
				log.Fatal(fmt.Errorf("%s", consts.ERR_NO_API_KEY))
			}
		}
	}

	if !appConf.ThumbOff {
		conf.AppCachePath = filepath.Join(conf.CachePathRoot, consts.APP_NAME)
		if !exists(conf.AppCachePath) {
			if err := os.MkdirAll(conf.AppCachePath, os.ModePerm); err != nil {
				log.Fatal(err)
			}
		}
	}
}

func doSearch(service *youtube.Service, p string) *youtube.SearchListResponse {
	call := service.Search.List([]string{"snippet"}).
		Q(conf.Q).
		RegionCode(appConf.Region).
		MaxResults(appConf.MaxResults).
		Type("video")
	if len(p) == 6 {
		call = call.PageToken(p)
	}
	res, err := call.Do()
	if err != nil {
		log.Fatalf("doSearch Do error: %v\n", err.Error())
	}

	return res
}

func getMesg(pageInfo *youtube.PageInfo) string {
	return fmt.Sprintf(
		"current page count: %d, total: %d",
		pageInfo.ResultsPerPage,
		pageInfo.TotalResults,
	)
}

func showResult(list *youtube.SearchListResponse) []Line {
	var lines []Line
	for _, v := range list.Items {
		l := Line{
			Text: v.Snippet.Title,
			Data: "id:" + v.Id.VideoId,
		}

		if !appConf.ThumbOff {
			thumbUrl := v.Snippet.Thumbnails.Default.Url
			switch appConf.ThumbSize {
			case "high":
				if len(v.Snippet.Thumbnails.High.Url) > 0 {
					thumbUrl = v.Snippet.Thumbnails.High.Url
				}
			case "medium":
				if len(v.Snippet.Thumbnails.Medium.Url) > 0 {
					thumbUrl = v.Snippet.Thumbnails.Medium.Url
				}
			}

			thumb, err := getThumb(v.Id.VideoId, thumbUrl)
			if err != nil {
				log.Fatal(err.Error())
			}

			l.Icon = thumb
		}

		lines = append(lines, l)
	}

	return lines
}

func main() {
	f, err := os.OpenFile(
		filepath.Join(conf.AppCachePath, "log"),
		os.O_RDWR|os.O_CREATE|os.O_APPEND,
		0666,
	)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	// initial output
	b := Blocks{
		Lines:   []Line{},
		Massage: "enter for search",
	}

	j, err := json.Marshal(&b)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(j))

	ctx := context.Background()
	yt, err := youtube.NewService(ctx, option.WithAPIKey(appConf.API_KEY))
	if err != nil {
		log.Fatalf("Unable to create YouTube service: %v", err)
	}

	inDecoder := json.NewDecoder(os.Stdin)

	for {
		if err := inDecoder.Decode(&j_in); err != nil {
			log.Fatal(err)
		}

		switch j_in.Name {
		case "execute custom input":
			b.Lines = []Line{}
			conf.Q = j_in.Value
			list := doSearch(yt, "")
			b.Massage = getMesg(list.PageInfo)
			b.Lines = showResult(list)
			b.Lines = append(b.Lines, Line{Text: "Next ->", Data: "next:" + list.NextPageToken})
		case "select entry":
			if res, ok := parseToken(j_in.Data); ok {
				switch res[0] {
				case "id":
					openInMPV("https://www.youtube.com/watch?v=" + res[1])
					os.Exit(0)
				case "next":
					list := doSearch(yt, res[1])
					b.Massage = getMesg(list.PageInfo)
					b.Lines = showResult(list)
					if len(list.NextPageToken) == 6 {
						b.Lines = append(
							b.Lines,
							Line{Text: "Next ->", Data: "next:" + list.NextPageToken},
						)
					}
					if len(list.PrevPageToken) == 6 {
						b.Lines = append(
							b.Lines,
							Line{Text: "<- Prev", Data: "prev:" + list.PrevPageToken},
						)
					}
				case "prev":
					list := doSearch(yt, res[1])
					b.Massage = getMesg(list.PageInfo)
					b.Lines = showResult(list)
					if len(list.NextPageToken) == 6 {
						b.Lines = append(
							b.Lines,
							Line{Text: "Next ->", Data: "next:" + list.NextPageToken},
						)
					}
					if len(list.PrevPageToken) == 6 {
						b.Lines = append(
							b.Lines,
							Line{Text: "<- Prev", Data: "prev:" + list.PrevPageToken},
						)
					}
				default:
					b.Lines = []Line{}
					b.Massage = "unkwnown input"
				}
			}
		}

		b.Prompt = ""

		j, err = json.Marshal(&b)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(string(j))

	}

}

func parseToken(s string) ([]string, bool) {
	if sub := strings.Split(s, ":"); len(sub) == 2 {
		switch sub[0] {
		case "id":
			if vidRx.MatchString(s) && len(sub[1]) == 11 {
				return sub, true
			}
		case "next", "prev":
			if npRx.MatchString(s) && len(sub[1]) == 6 {
				return sub, true
			}
		}
	}

	return []string{}, false
}

func getThumb(id, url string) (string, error) {
	ext := filepath.Ext(filepath.Base(url))
	if len(ext) == 0 && len(filepath.Base(url)) != 0 {
		ext = ".jpg"
	}
	if len(filepath.Base(url)) == 0 {
		return "", fmt.Errorf("can't read file from url:'%s'", url)
	}

	thumbName := "t" + id + ext
	switch appConf.ThumbSize {
	case "high":
		thumbName = "h" + thumbName
	case "medium":
		thumbName = "m" + thumbName
	default:
		thumbName = "d" + thumbName
	}

	thumbPath := filepath.Join(conf.AppCachePath, thumbName)
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
