package backup

type BackupDTO struct {
	FilePath  string `json:"filePath"`
	CreatedAt string `json:"createdAt"`
	Removed   int    `json:"removed"`
}
