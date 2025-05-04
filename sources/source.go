package sources

import "time"

type Source interface {
	ListFiles() <-chan FileInfo
	GetFile(string) ([]byte, error)
	SaveFile(string, []byte, string) error
	Exists(string) bool
	GetFileHash(string) (string, error)
	RemoveFile(string) error
	CalculateFileHash([]byte) (string, error)
	GetFileLastModified(remote_path string) (time.Time, error)
}

type FileInfo struct {
	Path         string
	Md5          string
	Filename     string
	Permission   string
	RemoteHash   string
	LastModified time.Time
}
