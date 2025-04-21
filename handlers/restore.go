package handlers

import (
	log "github.com/sirupsen/logrus"
	"os"
	"uelei/capivara-sync/compressor"
	"uelei/capivara-sync/db"
	"uelei/capivara-sync/sources"
)

func Restore(origin sources.Source, destination sources.Source, snap_date string, clean bool) error {
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
	var snapshotp *db.SnapShotRecord
	var error error
	if snap_date == "" {
		log.Warning("SnapShot Date not provided, using the last snapshot")
		snapshotp, error = db.GetLastSnap(database)
		if error != nil {
			log.Fatal("Error getting snapshot:", error)
		}
	} else {
		log.Info("Searching SnapShot date:", snap_date)
		snapshotp, error = db.GetSnapByDate(database, snap_date)

		if error != nil {
			log.Fatal("Error getting snapshot:", error)
		}
	}
	snapshot := snapshotp
	log.Info("Restoring snapshot ID: ", snapshot.Id, " Date: ", snapshot.Date)

	files, err := db.ListFilesbySnapshot(database, snapshot.Id)

	if clean {

		log.Warn("Clean Flag activated - removing all files in origin that are not in the snapshot")

		localfiles := origin.ListFiles()
		if err != nil {
			log.Fatalf("Error listing files: %v", err)
		}

		for file := range localfiles {

			found := false
			for _, record := range files {
				if record.MD5 == file.Md5 {
					found = true
					break
				}
			}
			if !found {
				log.Warn("removing the file from origin storage: ", file.Path)
			}

		}
	}

	for _, file := range files {
		log.Debug("Restoring file:", file.Path)
		// Check if the file exists in the origin
		exists := origin.Exists(file.Path)
		hash, _ := origin.GetFileHash(file.Path)
		if !exists || hash != file.MD5 {
			// Get the file from the destination
			data, err := destination.GetFile("block_" + file.MD5 + ".zst")
			if err != nil {
				log.Fatal("Error getting file:", err)
			} else {
				log.Info("File restored from destination storage")
				datafile, error := compressor.DecompressZstd(data)
				if error != nil {
					log.Fatal("Error decompressing file:", error)
				}
				error = origin.SaveFile(file.Path, datafile, file.Permission)
				if error != nil {
					log.Fatal("Error saving file on local:", error)
				}
			}
		} else {

			log.Debug("File already exists in origin storage ", file.Path)
		}

	}

	return nil

}
