package model

import "time"

type FileInfo struct {
	FileID    string
	FileName  string
	FileURL   string
	FileSize  int64
	FileType  string
	CreatedAt time.Time
}
