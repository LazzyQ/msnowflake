package config

// LogConfig 默认log 配置
type LogConfig struct {
	Filename       string `json:"filename"`
	LogSuffix      string `json:"log_suffix"`
	Level          string `json:"level"`
	MaxRemainCount uint   `json:"max_remain_count"`
}

// GetPath 日志文件路径
func (l LogConfig) GetLevel() string {
	return l.Level
}

func (l LogConfig) GetFilename() string {
	return l.Filename
}

func (l LogConfig) GetMaxRemainCount() uint {
	return l.MaxRemainCount
}


func (l LogConfig) GetLogSuffix() string {
	return l.LogSuffix
}
