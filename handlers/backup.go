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

func Backup(origin sources.Source, destination sources.Source) error {

	db_file, err := destination.GetFile("snapshot_files.db")

	if err == nil {
		log.Info("Database file already exists in remote storage")
		if err := os.WriteFile("snapshot_files.db", db_file, 0644); err != nil {
			return fmt.Errorf("failed to write file: %v", err)
		}
		destination.RemoveFile("snapshot_files.db")

	}

	database, err := db.InitDB("snapshot_files.db")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		log.Info("Clean up environment")
		// database.Close()
		db_file, _ := os.ReadFile("snapshot_files.db")
		// Move the database file to another folder
		log.Info("Saving database file to remote storage")
		saveerr := destination.SaveFile("snapshot_files.db", db_file, "-rw-r--r--")
		if saveerr != nil {
			log.Fatal("Error saving database file to local storage:", saveerr)
		} else {
			// newPath := "new_folder/snapshot_files.db"
			os.Remove("snapshot_files.db")
			// err := os.Rename("snapshot_files.db", newPath)
			// if err != nil {
			// 	log.Fatalf("Failed to move database file: %v", err)
		}
	}()

	snap_id, snap_error := db.SaveSnapshot(database)
	if snap_error != nil {
		log.Fatal("Error saving snapshot:", snap_error)
	}
	files, err := origin.ListFiles()
	if err != nil {
		log.Fatalf("Error listing files: %v", err)
	}

	fmt.Println("Files in folder:")
	for _, file := range files {
		var error error

		log.Debug(file.Path, file.Md5, file.Filename)
		remote_filename := "block_" + file.Md5 + ".zst"
		exists := destination.Exists(remote_filename)
		hf, error := db.GetFileByHash(database, file.Md5)
		if error != nil {
			fmt.Println("Error getting file by hash:", error)
		}
		reason := ""
		remote_hash := ""
		if exists == true {
			remote_hash, error = destination.GetFileHash(remote_filename)
			if error != nil {
				log.Error("Error getting file hash:", error)
				// remote_hash = file.RemoteHash
			}
		} else {
			reason = "File does not exist in remote storage"
		}

		if hf == nil {
			reason = "File does not exist in database"
		}
		if hf != nil && remote_hash != hf.RemoteHash && exists == true {
			reason = "Remote file hash does not match"
		}
		if reason != "" {
			log.Info("File is being backuped: ", file.Path, " because: ", reason)
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
			destination.SaveFile(remote_filename, compresedfile, "-rw-r--r--")
		}

		db.SaveFileInfo(database, file.Path, file.Md5, file.Permission, int(snap_id), remote_hash)
	}

	// Implement the backup logic here
	return nil
}
