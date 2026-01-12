APP_NAME=backup_master
CMD_PATH=cmd/app/main.go
BUILD_DIR=build

# Определяем ОС
ifeq ($(OS),Windows_NT)
	EXT=.exe
	RM=del /Q
	MKDIR=mkdir
else
	EXT=
	RM=rm -f
	MKDIR=mkdir -p
endif

# ========================
# ЗАПУСК
# ========================
run:
	go run $(CMD_PATH)

# ========================
# СБОРКА
# ========================
build:
	$(MKDIR) $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(APP_NAME)$(EXT) $(CMD_PATH)

# ========================
# СБОРКА С ИНФО
# ========================
build-info:
	$(MKDIR) $(BUILD_DIR)
	go build -v -o $(BUILD_DIR)/$(APP_NAME)$(EXT) $(CMD_PATH)

# ========================
# ОЧИСТКА
# ========================
clean:
	$(RM) $(BUILD_DIR)/$(APP_NAME)$(EXT)

# ========================
# ПОЛНЫЙ РЕБИЛД
# ========================
rebuild: clean build

fyne-reload:
	sudo apt remove golang-go
	sudo apt autoremove
	sudo apt update
	sudo apt install golang-go
	source ~/.bashrc
	echo 'export PATH="$PATH:$(go env GOPATH)/bin"' >> ~/.bashrc
	go install fyne.io/fyne/v2/cmd/fyne@latest

# ========================
# ПРОВЕРКА
# ========================
check:
	go vet ./...
	go test ./...
