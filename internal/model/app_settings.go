package model

type AppSettings struct {
	BackupRootPath  string
	MaxStorageBytes int64
	ThemeMode       string // system | light | dark
}
