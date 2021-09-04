package setting

import "time"

type ServerSettingS struct {
	RunMode string
	HttpPort string
	ReadTimeout time.Duration
	WriteTimeout time.Duration
}

type AppSettingS struct {
	DefaultPageSize int
	MaxPageSize int
	LogSavePath string
	LogFileName string
	LogFileExt string
}

type DatabaseSettingS struct {
	
}
