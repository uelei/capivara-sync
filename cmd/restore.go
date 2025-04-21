/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"uelei/capivara-sync/db"
	"uelei/capivara-sync/handlers"
)

var list, clean bool
var snap string

// restoreCmd represents the restore command
var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore a backup from a snapshot",
	Long:  `Restore the files from the dest to a origin Source.`,
	Run: func(cmd *cobra.Command, args []string) {

		destsource, error := BuildSource(dest, destpass)
		if error != nil {
			log.Warn("Error building origin source:", error)
		}

		if list {
			database, er := handlers.GetDatabaseFromRemote(destsource)
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
				saveerr := destsource.SaveFile("snapshot_files.db", db_file, "-rw-r--r--")
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

			fmt.Println("Listing snapshots")

			snaps, err := db.ListSnapShots(database)
			if err != nil {
				log.Fatal("Error listing snapshots:", err)
			}
			for _, snp := range snaps {
				fmt.Println("Snapshot ID:", snp.Id, "Date:", snp.Date)
			}

		} else {
			originsource, error := BuildSource(origin, originpass)
			if error != nil {
				log.Warn("Error building origin source:", error)
			}

			// starting the handler
			if err := handlers.Restore(originsource, destsource, snap, clean); err != nil {
				log.Fatal("Error restoring snapshot:", err)
			}
		}

	},
}

func init() {
	restoreCmd.Flags().BoolVarP(&list, "list", "l", false, "List snapshots dates")
	restoreCmd.Flags().BoolVarP(&clean, "clean", "c", false, "List snapshots dates")

	restoreCmd.Flags().StringVarP(&origin, "origin", "o", "", "origin: local or ssh (required)")
	restoreCmd.Flags().StringVarP(&dest, "dest", "d", "", "destination: local or ssh (required)")

	restoreCmd.Flags().StringVarP(&snap, "snap", "s", "", "snap date to restore if not latest (optional)")

	if err := restoreCmd.MarkFlagRequired("dest"); err != nil {
		log.Fatal(err)
	}

	restoreCmd.PersistentFlags().StringVar(&originpass, "origin-password", "", "SSH password (optional, will prompt if not provided)")
	restoreCmd.PersistentFlags().StringVar(&destpass, "dest-password", "", "SSH password (optional, will prompt if not provided)")

	rootCmd.AddCommand(restoreCmd)

}
