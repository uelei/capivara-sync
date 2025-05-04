package cmd

import (
	"fmt"
	"uelei/capivara-sync/handlers"
	"uelei/capivara-sync/sources"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var origin, dest, originpass, destpass, originuser, destuser string
var skip, compress bool

// backupCmd represents the backup command
var backupCmd = &cobra.Command{

	Use:   "backup",
	Short: "Backup a origin folder to a destination",
	Long:  `Backup a origin folder to a destination. using the given source and remote`,
	Run: func(cmd *cobra.Command, args []string) {

		originsource, error := BuildSource(origin, originpass, originuser)
		if error != nil {
			log.Warn("Error building origin source:", error)
		}

		destsource, error := BuildSource(dest, destpass, destuser)
		if error != nil {
			log.Warn("Error building origin source:", error)
		}
		log.Warn("compress mode is ", compress)
		if error := handlers.Backup(originsource, destsource, sources.Setting{Compress: compress, Skip_hash: skip}); error != nil {
			log.Fatal("Error backing up:", error)
		}

		fmt.Println("Backup completed successfully")
	},
}

func init() {

	// log.SetLevel(log.DebugLevel)
	// Here you will define your flags and configuration settings.
	backupCmd.Flags().BoolVarP(&skip, "skip", "s", false, "Skip mode no check remote checksum")
	backupCmd.Flags().BoolVar(&compress, "x", true, "Compress mode, compress files before sending to remote")

	// Flags
	backupCmd.Flags().StringVarP(&origin, "origin", "", "o", "origin: local or ssh (required)")
	backupCmd.Flags().StringVarP(&dest, "dest", "", "o", "destination: local or ssh (required)")

	if error := backupCmd.MarkFlagRequired("origin"); error != nil {
		log.Fatal("Error marking origin flag as required:", error)
	}
	if error := backupCmd.MarkFlagRequired("dest"); error != nil {
		log.Fatal("Error marking dest flag as required:", error)
	}

	backupCmd.PersistentFlags().StringVar(&originpass, "origin-password", "", "SSH/DAV password (optional, will prompt if not provided)")
	backupCmd.PersistentFlags().StringVar(&destpass, "dest-password", "", "SSH/DAV password (optional, will prompt if not provided)")

	backupCmd.PersistentFlags().StringVar(&originuser, "origin-user", "", "SSH/DAV user (optional, will prompt if not provided)")
	backupCmd.PersistentFlags().StringVar(&destuser, "dest-user", "", "SSH/DAV user (optional, will prompt if not provided)")

	rootCmd.AddCommand(backupCmd)

}
