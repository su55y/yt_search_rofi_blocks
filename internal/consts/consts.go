package consts

const (
	ENV_YT_API_KEY   = "YT_SEARCH_API_KEY"
	ENV_YT_CACHE_DIR = "YT_SEARCH_CACHE_DIR"
	ENV_CACHE_HOME   = "XDG_CACHE_HOME"
	ENV_CONFIG_HOME  = "XDG_CONFIG_HOME"
	APP_NAME         = "yt_search_rofi_blocks"
	APP_CONFIG_NAME  = "config.yaml"
	DEF_CACHE_PATH   = ".cache"
	DEF_CONFIG_PATH  = ".config/yt_search_rofi_blocks/config.yaml"
	DEF_CONFIG       = `# youtube api key (https://console.cloud.google.com/)
# api_key: "<YT_API_KEY>"
# api_key_path: "/path/to/api_key"

# ISO 3166-1 alpha-2 country code
region: "UA"

# max results
max_results: 10

# "/home/user/.cache/yt_search_rofi_blocks" by default
# cache_dir: "/path/to/cache"

# thumbnails are loaded into the cache directory with format '(h/m/d)(t)(video_id).ext' 
# you can disable thumbnails loading
thumbnails_disable: false

# thumbnails size: high(~15-30k),medium(~8-15k),default(~3-4k)
thumbnails_size: "default"`

	// info
	INF_NEW_CONFIG = "new config written to %s"

	// errors
	ERR_NO_API_KEY        = "api key was not found either in the config or in the env variable"
	ERR_NO_API_KEY_FILE   = "api key not found in '%s': %v"
	ERR_API_KEY_FILE_READ = "no api key in '%s'"
	ERR_CONFIG_LOAD       = "load config error: %v"
)
