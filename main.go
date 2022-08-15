package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/su55y/yt_search_rofi_blocks/internal/blocks"
	"github.com/su55y/yt_search_rofi_blocks/internal/config"
	"github.com/su55y/yt_search_rofi_blocks/internal/consts"
	"github.com/su55y/yt_search_rofi_blocks/internal/search"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

var (
	conf    Config
	appConf config.AppConfig

	rawIn       string
	blocksInput blocks.BlocksIn
)

type Config struct {
	AppCachePath  string
	HomePath      string
	NextPageToken string
	ConfPathRoot  string
	CachePathRoot string
	ConfFullPath  string
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
	ctx := context.Background()
	yt, err := youtube.NewService(ctx, option.WithAPIKey(appConf.API_KEY))
	if err != nil {
		log.Fatalf("Unable to create YouTube service: %v", err)
	}

	searchService := search.NewSearchService(yt, appConf, conf.AppCachePath)

	// initial output
	blocksOutput := blocks.Blocks{
		Lines:   []blocks.Line{},
		Message: "enter for search",
	}

	j, err := json.Marshal(&blocksOutput)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(j))

	inDecoder := json.NewDecoder(os.Stdin)

	var runMPV bool

	for {
		if err := inDecoder.Decode(&blocksInput); err != nil {
			log.Fatal(err)
		}

		switch blocksInput.Name {
		case "execute custom input":
			searchService.NewQuery(blocksInput.Value)
			blocksOutput = searchService.DoSearch("")
		case "select entry":
			switch sel := blocks.ParseSelect(blocksInput.Data); sel.Action {
			case "open":
				blocksOutput.Message = sel.Message
				blocksOutput.ActEntr = sel.Selected
				if runMPV = openInMPV(sel.Id); !runMPV {
					blocksOutput.Message += " : error"
				}
			case "next", "prev":
				blocksOutput = searchService.DoSearch(sel.Id)
			case "clear":
				blocksOutput.Lines = []blocks.Line{}
				blocksOutput.Message = sel.Message
			case "err":
				blocksOutput.Message = sel.Message
			default:
				blocksOutput.Message = "unkwnown error"
			}
		}

		j, err = json.Marshal(&blocksOutput)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(string(j))
		if runMPV {
			time.Sleep(2 * time.Second)
			os.Exit(0)
		}

	}
}

func openInMPV(id string) bool {
	c := exec.Command("mpv", "https://www.youtube.com/watch?v="+id)
	if err := c.Start(); err != nil {
		log.Println(err.Error())
		return false
	}

	return c.Process.Pid > 0
}
