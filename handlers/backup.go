package handlers

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"uelei/capivara-sync/compressor"
	"uelei/capivara-sync/db"
	"uelei/capivara-sync/sources"
)

func GetRemoteFileName(file sources.FileInfo) string {
	return "block_" + file.Md5 + ".zst"
}

func Backup(origin sources.Source, destination sources.Source, setting sources.Setting) error {

	database, er := GetDatabaseFromRemote(destination)
	if er != nil {
		log.Error("Error getting database from remote:", er)
	}

	defer func() {
		log.Info("Clean up environment")
		if err := database.Close(); err != nil {
			log.Fatal("Error closing database:", err)
		}

		db_file, _ := os.ReadFile("snapshot_files.db")
		// Move the database file to another folder
		log.Info("Saving database file to remote storage")
		saveerr := destination.SaveFile("snapshot_files.db", db_file, "-rw-r--r--")
		if saveerr != nil {
			log.Fatal("Error saving database file to local storage:", saveerr)
		} else {
			if err := os.Remove("snapshot_files.db"); err != nil {
				log.Error("Error removing local database file:", err)
			}
		}
	}()
	snap_id, er := db.SaveSnapshot(database)
	if er != nil {
		log.Fatal("Error saving snapshot:", er)
	}
	files := origin.ListFiles()

	fmt.Println("Files in folder:")
	for file := range files {
		var error, errr error

		log.Debug("File is ", file.Path, " MD5: ", file.Md5, " Filename: ", file.Filename)
		remote_filename := GetRemoteFileName(file)
		exists := destination.Exists(remote_filename)

		hf, error := db.GetFileByHash(database, file.Md5)
		if error != nil {
			fmt.Println("Error getting file by hash:", error)
		}
		reason := ""
		remote_hash := ""
		upload := true
		status := "upload"
		if exists {
			remote_hash, error = destination.GetFileHash(remote_filename)
			if error != nil {
				log.Error("Error getting file hash:", error)
				// remote_hash = file.RemoteHash
			}
		} else {
			reason = "File does not exist in remote storage."
		}

		if hf == nil {
			reason = "file has not been backed up previously."
		} else {
			log.Warn("file exists hash is :", remote_hash, " file hash is: ", hf.RemoteHash)
		}

		if hf != nil && remote_hash != hf.RemoteHash && exists {
			reason = "Remote file hash does not match."
			if setting.Skip_hash {
				upload = false
				status = "skip"
			}
		}

		if reason != "" && upload {
			log.Info("Backing up file: ", file.Path, " — reason: ", reason)
			origin_file_bytes, error := origin.GetFile(file.Path)
			if error != nil {
				fmt.Println("Error getting file:", error)
			}
			log.Debug("File size: ", len(origin_file_bytes))
			var compresedfile []byte
			compresedfile, errr = compressor.CompressZstd(origin_file_bytes)
			if errr != nil {
				fmt.Println("Error compressing file:", errr)
			}
			remote_hash, error = destination.CalculateFileHash(compresedfile)
			if error != nil {
				log.Error("Error calculating file hash:", error)
			}
			log.Debug(" size of file: ", len(compresedfile))
			log.Info("Writing file to remote:", remote_filename)
			if err := destination.SaveFile(remote_filename, compresedfile, "-rw-r--r--"); err != nil {
				log.Error("Error saving file to remote storage:", err)
			} else {
				log.Debug("File saved to remote storage successfully")
			}
		}

		if err := db.SaveFileInfo(database, file.Path, file.Md5, file.Permission, int(snap_id), remote_hash, status); err != nil {
			log.Error("Error saving file info to database:", err)
		} else {
			log.Debug("File info saved to database successfully")
		}
	}

	return nil
}
