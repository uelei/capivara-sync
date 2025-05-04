package sources

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestExists(t *testing.T) {
	ls := Localsource{Localpath: "./testdata/"}
	assert.True(t, ls.Exists("testfile.txt"))
	assert.False(t, ls.Exists("nonexistent.txt"))
}

func TestGetFileHash(t *testing.T) {
	ls := Localsource{Localpath: "./testdata/"}
	hash, err := ls.GetFileHash("testfile.txt")
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
}

func TestSaveAndRemoveFile(t *testing.T) {
	ls := Localsource{Localpath: "./testdata/"}
	data := []byte("test data")
	err := ls.SaveFile("newfile.txt", data, "-rw-r--r--")
	assert.NoError(t, err)

	assert.True(t, ls.Exists("newfile.txt"))

	err = ls.RemoveFile("newfile.txt")
	assert.NoError(t, err)
	assert.False(t, ls.Exists("newfile.txt"))
}

func TestGetFileLastModified(t *testing.T) {
	ls := Localsource{Localpath: "./testdata/"}
	modTime, err := ls.GetFileLastModified("testfile.txt")
	assert.NoError(t, err)
	assert.WithinDuration(t, time.Now(), modTime, time.Hour*24)
}

func TestCalculateFileHash(t *testing.T) {
	ls := Localsource{}
	data := []byte("test data")
	hash, err := ls.CalculateFileHash(data)
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
}

// func TestListFiles(t *testing.T) {
// 	ls := Localsource{Localpath: "./testdata/"}
// 	ch := ls.ListFiles()
//
// 	var files []FileInfo
// 	for file := range ch {
// 		files = append(files, file)
// 	}
//
// 	assert.NotEmpty(t, files)
// 	for _, file := range files {
// 		assert.NotEmpty(t, file.Path)
// 		assert.NotEmpty(t, file.Md5)
// 		assert.NotEmpty(t, file.Filename)
// 		assert.NotEmpty(t, file.Permission)
// 	}
// }

func TestFileModeFromString(t *testing.T) {
	mode, err := FileModeFromString("-rw-r--r--")
	assert.NoError(t, err)
	assert.Equal(t, os.FileMode(0644), mode)

	_, err = FileModeFromString("invalid")
	assert.Error(t, err)
}
