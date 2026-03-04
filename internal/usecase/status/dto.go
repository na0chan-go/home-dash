package status

type StatusDTO struct {
	AppVersion             string          `json:"appVersion"`
	UptimeSeconds          int64           `json:"uptimeSeconds"`
	ServerTime             string          `json:"serverTime"`
	DB                     DBStatusDTO     `json:"db"`
	Config                 ConfigStatusDTO `json:"config"`
	Auth                   AuthStatusDTO   `json:"auth"`
	LastBackup             *string         `json:"lastBackup,omitempty"`
	LastDashboardRefreshAt *string         `json:"lastDashboardRefreshAt,omitempty"`

	DBError         string `json:"-"`
	ConfigError     string `json:"-"`
	LastBackupError string `json:"-"`
}

type DBStatusDTO struct {
	Path string `json:"path"`
	OK   bool   `json:"ok"`
}

type ConfigStatusDTO struct {
	GarbageScheduleLoaded bool `json:"garbageScheduleLoaded"`
}

type AuthStatusDTO struct {
	Enabled bool `json:"enabled"`
}
