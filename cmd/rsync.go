package cmd

import (
	"fmt"
	"uelei/capivara-sync/handlers"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var delete bool

// syncCmd represents the backup command
var syncCmd = &cobra.Command{

	Use:   "rsync",
	Short: "Rsync a origin folder to a destination",
	Long:  `Rsync a origin folder to a destination. using the given source and remote`,
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
		if error := handlers.RSync(originsource, destsource, delete); error != nil {
			log.Fatal("Error backing up:", error)
		}

		fmt.Println("Backup completed successfully")
	},
}

func init() {

	syncCmd.Flags().BoolVarP(&delete, "delete", "d", false, "Delete files on destination if not on the origin")
	// Flags
	syncCmd.Flags().StringVarP(&origin, "origin", "", "o", "origin: local or ssh (required)")
	syncCmd.Flags().StringVarP(&dest, "dest", "", "o", "destination: local or ssh (required)")

	if error := syncCmd.MarkFlagRequired("origin"); error != nil {
		log.Fatal("Error marking origin flag as required:", error)
	}
	if error := syncCmd.MarkFlagRequired("dest"); error != nil {
		log.Fatal("Error marking dest flag as required:", error)
	}

	syncCmd.PersistentFlags().StringVar(&originpass, "origin-password", "", "SSH/DAV password (optional, will prompt if not provided)")
	syncCmd.PersistentFlags().StringVar(&destpass, "dest-password", "", "SSH/DAV password (optional, will prompt if not provided)")

	syncCmd.PersistentFlags().StringVar(&originuser, "origin-user", "", "SSH/DAV user (optional, will prompt if not provided)")
	syncCmd.PersistentFlags().StringVar(&destuser, "dest-user", "", "SSH/DAV user (optional, will prompt if not provided)")

	rootCmd.AddCommand(syncCmd)
}
