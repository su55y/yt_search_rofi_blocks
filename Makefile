APP_BIN = $(HOME)/code/rofi/yt_search_rofi_blocks/yt_search_rofi_blocks
# APP_BIN = yt_search_rofi_blocks

build: clean $(APP_BIN)

$(APP_BIN):
	go build -o $(APP_BIN) ./main.go
	# CGO_ENABLED=0 go build -o $(APP_BIN) ./main.go

clean:
	rm $(APP_BIN) || true

