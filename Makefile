APP_NAME=backup_master
CMD_PATH=cmd/app/main.go
BUILD_DIR=build

# ========================
# ЗАПУСК
# ========================
run:
	go run $(CMD_PATH)

# ========================
# СБОРКА Linux
# ========================
build-linux:
	fyne-cross linux -arch amd64 -name BackupMaster -app-id com.example.backupmaster ./cmd/app

# ========================
# СБОРКА Windows
# ========================
build-windows:
	fyne-cross windows -name BackupMaster -app-id com.example.backupmaster ./cmd/app


# ========================
# ОЧИСТКА
# ========================
clean:
	$(RM) $(BUILD_DIR)/$(APP_NAME)$(EXT)

# ========================
# ПОЛНЫЙ РЕБИЛД
# ========================
rebuild: clean build

# ========================
# ПРОВЕРКА
# ========================
check:
	go test -short ./...
