package handlers

import (
	"fmt"
	"uelei/capivara-sync/sources"
)
import log "github.com/sirupsen/logrus"

func RSync(origin sources.Source, destination sources.Source, delete bool) error {

	files := origin.ListFiles()

	if delete {
		log.Info("deleting files on destination if not in the origin")
		// Remove the files not in the origin
		destination_files := destination.ListFiles()
		for file := range destination_files {
			oexists := origin.Exists(file.Path)
			if !oexists {
				log.Error("File not found in origin, removing from destination: ", file.Path)
				error := destination.RemoveFile(file.Path)
				if error != nil {
					log.Error("Error removing file from destination:", error)
				}
			}
		}
	}
	log.Info("Syncing files from origin to destination")
	for file := range files {
		var error error
		log.Debug("file : ", file.Path, " MD5: ", file.Md5, " Filename: ", file.Filename, " LT : ", file.LastModified)
		exists := destination.Exists(file.Path)
		reason := ""
		remote_hash := ""
		if exists {
			// get remote hash
			remote_hash, error = destination.GetFileHash(file.Path)
			if error != nil {
				log.Error("Error getting file hash:", error)
			}
			if remote_hash != file.Md5 {

				last_modified, error := destination.GetFileLastModified(file.Path)
				if error != nil {
					log.Error("Error getting file last modified:", error)

				}
				log.Info("local hash is :", file.Md5, " remote_hash is : ", remote_hash)
				if last_modified.After(file.LastModified) {

					log.Warn("The File: ", file.Path, " is older: ", TimeToString(file.LastModified), " then remote: ", TimeToString(last_modified))
					reason = ""
				} else {
					reason = "Remote file hash does not match."
				}
			} else {
				log.Debug("File already exists in remote storage, skipping upload. " + file.Path)
			}
		} else {
			reason = "File does not exist in remote storage."
		}

		if reason != "" {
			log.Info("Sync up file: ", file.Path, " — reason: ", reason)
			origin_file_bytes, error := origin.GetFile(file.Path)
			if error != nil {
				fmt.Println("Error getting file:", error)
			}

			log.Info("Writing file to remote:", file.Path)
			if err := destination.SaveFile(file.Path, origin_file_bytes, "-rw-r--r--"); err != nil {
				log.Error("Error saving file to remote storage:", err)
			} else {
				log.Debug("File saved to remote storage successfully")
			}
		}
		log.Debug("File synced successfully " + file.Path)
	}

	return nil
}
