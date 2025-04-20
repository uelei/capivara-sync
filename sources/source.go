package sources

type Source interface {
	ListFiles() ([]FileInfo, error)
	GetFile(string) ([]byte, error)
	SaveFile(string, []byte, string) error
	Exists(string) bool
	GetFileHash(string) (string, error)
	RemoveFile(string) error
}

type FileInfo struct {
	Path       string
	Md5        string
	Filename   string
	Permission string
	RemoteHash string
}
