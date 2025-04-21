package handlers

import (
	"database/sql"
	"os"
	"uelei/capivara-sync/db"
	"uelei/capivara-sync/sources"

	log "github.com/sirupsen/logrus"
)

func GetDatabaseFromRemote(destination sources.Source) (*sql.DB, error) {
	db_file, err := destination.GetFile("snapshot_files.db")

	if err == nil {
		log.Info("Database file already exists in remote storage")
		if err := os.WriteFile("snapshot_files.db", db_file, 0644); err != nil {
			log.Error("failed to write file:", err)

		}
		if err := destination.RemoveFile("snapshot_files.db"); err != nil {
			log.Error("Error removing file from remote storage:", err)
		}

	} else {
		log.Error("Error getting database file from remote storage:", err)
	}

	database, err := db.InitDB("snapshot_files.db")
	if err != nil {
		log.Fatal(err)
	}
	return database, nil
}
