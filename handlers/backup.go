package handlers

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"uelei/capivara-sync/compressor"
	"uelei/capivara-sync/db"
	"uelei/capivara-sync/sources"
)
import log "github.com/sirupsen/logrus"

func Backup(origin sources.Source, destination sources.Source, skip bool) error {

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
			// newPath := "new_folder/snapshot_files.db"
			if err := os.Remove("snapshot_files.db"); err != nil {
				log.Error("Error removing local database file:", err)
			}
			// err := os.Rename("snapshot_files.db", newPath)
			// if err != nil {
			// 	log.Fatalf("Failed to move database file: %v", err)
		}
	}()
	snap_id, er := db.SaveSnapshot(database)
	if er != nil {
		log.Fatal("Error saving snapshot:", er)
	}
	files := origin.ListFiles()

	fmt.Println("Files in folder:")
	for file := range files {
		var error error

		log.Debug(file.Path, file.Md5, file.Filename)
		remote_filename := "block_" + file.Md5 + ".zst"
		exists := destination.Exists(remote_filename)
		// old file hash
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
		}
		if hf != nil && remote_hash != hf.RemoteHash && exists {
			reason = "Remote file hash does not match."
			if skip {
				upload = false
				status = "skip"
			}
		}

		if reason != "" && upload {
			log.Info("Backing up file: ", file.Path, " — reason: ", reason)
			remote_file_bytes, error := origin.GetFile(file.Path)
			if error != nil {
				fmt.Println("Error getting file:", error)
			}
			compresedfile, errr := compressor.CompressZstd(remote_file_bytes)
			if errr != nil {
				fmt.Println("Error compressing file:", errr)
			}
			hash := md5.Sum(compresedfile) // returns [16]byte
			remote_hash = hex.EncodeToString(hash[:])

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

	// Implement the backup logic here
	return nil
}
