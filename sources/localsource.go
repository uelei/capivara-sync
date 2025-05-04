package sources

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"crypto/md5"

	log "github.com/sirupsen/logrus"

	"encoding/hex"
)

type Localsource struct {
	Localpath string
}

func (l Localsource) Exists(path string) bool {
	log.Debug("Checking if file exists on remote:", l.Localpath+path)
	_, err := os.Stat(l.Localpath + path)
	return err == nil

}

func (l Localsource) GetFileHash(path string) (string, error) {

	filePath := l.Localpath + path
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Error("Error closing database file:", err)
		}
	}()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func (l Localsource) GetFile(path string) ([]byte, error) {
	return os.ReadFile(l.Localpath + path)
}

func (l Localsource) RemoveFile(path string) error {
	return os.Remove(l.Localpath + path)
}

func (l Localsource) SaveFile(path string, data []byte, permission string) error {

	filePath := l.Localpath + path
	dir := filepath.Dir(filePath)

	// Create directories if they don't exist
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directories: %v", err)
	}

	// Write the data to the file
	perm, err := FileModeFromString(permission)
	if err != nil {
		return fmt.Errorf("failed to parse permission string: %v", err)

	}
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	err = os.Chmod(filePath, perm)
	if err != nil {
		log.Fatal(err)
	}
	return nil

}

func FileModeFromString(permStr string) (os.FileMode, error) {
	if len(permStr) != 10 {
		return 0, fmt.Errorf("invalid permission string: %q", permStr)
	}

	var mode os.FileMode

	// Owner
	if permStr[1] == 'r' {
		mode |= 0400
	}
	if permStr[2] == 'w' {
		mode |= 0200
	}
	if permStr[3] == 'x' || permStr[3] == 's' {
		mode |= 0100
	}

	// Group
	if permStr[4] == 'r' {
		mode |= 0040
	}
	if permStr[5] == 'w' {
		mode |= 0020
	}
	if permStr[6] == 'x' || permStr[6] == 's' {
		mode |= 0010
	}

	// Others
	if permStr[7] == 'r' {
		mode |= 0004
	}
	if permStr[8] == 'w' {
		mode |= 0002
	}
	if permStr[9] == 'x' || permStr[9] == 't' {
		mode |= 0001
	}

	return mode, nil
}

func (l Localsource) CalculateFileHash(filebyte []byte) (string, error) {
	hash := md5.Sum(filebyte) // returns [16]byte
	remote_hash := hex.EncodeToString(hash[:])

	return remote_hash, nil
}

func (l Localsource) ListFiles() <-chan FileInfo {
	ch := make(chan FileInfo)
	go func() {
		defer close(ch)

		err := filepath.WalkDir(l.Localpath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() {
				relative_path := strings.ReplaceAll(path, l.Localpath, "")
				md5sum, _ := l.GetFileHash(relative_path)
				info, err := d.Info()
				if err != nil {
					log.Fatal(err)
				}
				modTime := info.ModTime()
				log.Debug("File: ", path, " ", md5sum, " ", relative_path, " ", info.Mode().String())
				ch <- FileInfo{Path: relative_path, Md5: md5sum, Filename: d.Name(), Permission: info.Mode().Perm().String(), LastModified: modTime}
			}
			return nil
		})

		if err != nil {
			fmt.Printf("Error walking directory: %v\n", err)
			// return nil
		}
	}()

	return ch
}

func (l Localsource) GetFileLastModified(remote_path string) (time.Time, error) {
	filePath := l.Localpath + remote_path
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		log.Error("Error getting file info:", err)
	}
	return fileInfo.ModTime(), nil
}
