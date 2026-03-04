package status

import (
	"context"
	"strings"
	"time"

	"github.com/na0chan-go/home-dash/internal/ports"
)

const (
	defaultAppVersion = "unknown"
	tokyoTimezone     = "Asia/Tokyo"
	timestampLayout   = "2006-01-02T15:04:05Z07:00"
)

type GetStatusUseCase struct {
	clock           ports.Clock
	dbChecker       ports.DBHealthChecker
	garbageProvider ports.GarbageScheduleProvider
	backupReader    ports.BackupStatusReader
	dbPath          string
	authEnabled     bool
	appVersion      string
	startedAt       time.Time
}

func NewGetStatusUseCase(
	clock ports.Clock,
	dbChecker ports.DBHealthChecker,
	garbageProvider ports.GarbageScheduleProvider,
	backupReader ports.BackupStatusReader,
	dbPath string,
	authEnabled bool,
	appVersion string,
	startedAt time.Time,
) *GetStatusUseCase {
	version := strings.TrimSpace(appVersion)
	if version == "" {
		version = defaultAppVersion
	}

	if startedAt.IsZero() {
		startedAt = clock.NowUTC()
	}

	return &GetStatusUseCase{
		clock:           clock,
		dbChecker:       dbChecker,
		garbageProvider: garbageProvider,
		backupReader:    backupReader,
		dbPath:          dbPath,
		authEnabled:     authEnabled,
		appVersion:      version,
		startedAt:       startedAt.UTC(),
	}
}

func (u *GetStatusUseCase) Execute(ctx context.Context) (StatusDTO, error) {
	nowUTC := u.clock.NowUTC()
	nowTokyo := toTokyo(nowUTC)
	output := StatusDTO{
		AppVersion:    u.appVersion,
		UptimeSeconds: uptimeSeconds(nowUTC, u.startedAt),
		ServerTime:    nowTokyo.Format(timestampLayout),
		DB: DBStatusDTO{
			Path: u.dbPath,
			OK:   true,
		},
		Config: ConfigStatusDTO{
			GarbageScheduleLoaded: true,
		},
		Auth: AuthStatusDTO{
			Enabled: u.authEnabled,
		},
	}

	if err := u.dbChecker.Check(ctx); err != nil {
		output.DB.OK = false
		output.DBError = err.Error()
	}

	if _, err := u.garbageProvider.GetSchedule(ctx); err != nil {
		output.Config.GarbageScheduleLoaded = false
		output.ConfigError = err.Error()
	}

	if u.backupReader != nil {
		lastBackupAt, err := u.backupReader.LastBackupAt(ctx)
		if err != nil {
			output.LastBackupError = err.Error()
		} else if lastBackupAt != nil {
			value := lastBackupAt.In(nowTokyo.Location()).Format(timestampLayout)
			output.LastBackup = &value
		}
	}

	return output, nil
}

func toTokyo(ts time.Time) time.Time {
	loc, err := time.LoadLocation(tokyoTimezone)
	if err != nil {
		loc = time.FixedZone("JST", 9*60*60)
	}
	return ts.In(loc)
}

func uptimeSeconds(nowUTC, startedAt time.Time) int64 {
	if nowUTC.Before(startedAt) {
		return 0
	}
	return int64(nowUTC.Sub(startedAt).Seconds())
}
