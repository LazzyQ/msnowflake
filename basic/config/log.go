package config

// LogConfig 默认log 配置
type LogConfig struct {
	Path    string `json:"path"`
	Filename string `json:"filename"`
}

// GetPath 日志文件路径
func (l LogConfig) GetPath() string {
	return l.Path
}


func (l LogConfig) GetFilename() string {
	return l.Filename
}