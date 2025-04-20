package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/cobra"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
	"syscall"
	"uelei/capivara-sync/handlers"
	"uelei/capivara-sync/sources"
)

var origin, dest, originpass, destpass string

// backupCmd represents the backup command
var backupCmd = &cobra.Command{

	Use:   "backup",
	Short: "Backup a origin folder to a destination",
	Long:  `Backup a origin folder to a destination. using the given source and remote`,
	Run: func(cmd *cobra.Command, args []string) {

		var destsource, originsource sources.Source
		var error error
		var originauth, destauth ssh.AuthMethod

		if strings.Contains(origin, "@") {

			originParts := strings.Split(origin, "@")
			originUser := originParts[0]
			hostpaths := strings.Split(originParts[1], ":")
			originHost := hostpaths[0]
			originPath := EnsureTrailingSlash(hostpaths[1])
			fmt.Println("Origin User:", originUser)
			fmt.Println("Origin Host:", originHost)
			fmt.Println("Origin Path:", originPath)
			// Ask for password if not set via flag
			if originpass == "" {
				fmt.Print("Enter password: ")
				bytePassword, err := term.ReadPassword(int(syscall.Stdin))
				if err != nil {
					log.Fatal(err)
				}
				fmt.Println() // for newline
				originpass = string(bytePassword)

				originauth = ssh.Password(originpass)
			}
			originsource, error = sources.NewSSHSource(originParts[0], originHost+":22", originPath, originauth)

			if error != nil {
				log.Fatal(error)
			}

		} else {
			originsource = sources.Localsource{Localpath: EnsureTrailingSlash(origin)}

		}
		if strings.Contains(dest, "@") {
			destParts := strings.Split(dest, "@")
			destUser := destParts[0]
			hostpaths := strings.Split(destParts[1], ":")
			destHost := hostpaths[0]
			destPath := EnsureTrailingSlash(hostpaths[1])
			fmt.Println("Dest User:", destUser)
			fmt.Println("Dest Host:", destHost)
			fmt.Println("Dest Path:", destPath)

			// Ask for password if not set via flag
			if destpass == "" {
				fmt.Print("Enter password: ")
				bytePassword, err := term.ReadPassword(int(syscall.Stdin))
				if err != nil {
					log.Fatal(err)
				}
				fmt.Println() // for newline
				destpass = string(bytePassword)

				destauth = ssh.Password(destpass)
			}
			destsource, error = sources.NewSSHSource(destParts[0], destHost+":22", destPath, destauth)

			if error != nil {
				log.Fatal(error)
			}
		} else {
			destsource = sources.Localsource{Localpath: EnsureTrailingSlash(dest)}
		}

		handlers.Backup(originsource, destsource)
		fmt.Println("Backup completed successfully")
	},
}

// fmt.Println("starting")
// database, err := db.InitDB("snapshot_files.db")
//
//	if err != nil {
//		log.Fatal(err)
//	}
//
// defer database.Close()
//
// // // Backup
// // localsource := sources.Localsource{Localpath: "/tmp/storage/"}
// // remotesource := sources.Localsource{Localpath: "/tmp/restored/"}
// // sources.Backup(localsource, remotesource, database)
//
// // // Restore
// remotesource := sources.Localsource{Localpath: "/tmp/restored/"}
// // vaisalvar := sources.Localsource{Localpath: "/tmp/novo/"}
// // sources.Restore(remotesource, vaisalvar, database)
//
// // // Backup
// // localsource := sources.Localsource{Localpath: "/tmp/storage/"}
// auth := ssh.Password("uc3r2l53i")
// remote, error := sources.NewSSHSource("uelei", "192.168.1.10:22", "/mnt/1tb_disk/vai/", auth)
//
//	if error != nil {
//		log.Fatal(error)
//	}
//
// sources.Restore(remotesource, remote, database)
func init() {

	// Here you will define your flags and configuration settings.
	// backupCmd.Flags().BoolP("clean", "c", false, "Clean remote files")

	// Flags
	backupCmd.Flags().StringVarP(&origin, "origin", "", "o", "origin: local or ssh (required)")
	backupCmd.Flags().StringVarP(&dest, "dest", "", "o", "destination: local or ssh (required)")

	// backupCmd.Flags().StringVarP(&dest, "clean", "", "o", "destination: local or ssh (required)")

	backupCmd.MarkFlagRequired("origin")
	backupCmd.MarkFlagRequired("dest")

	backupCmd.PersistentFlags().StringVar(&originpass, "origin-password", "", "SSH password (optional, will prompt if not provided)")
	backupCmd.PersistentFlags().StringVar(&destpass, "dest-password", "", "SSH password (optional, will prompt if not provided)")

	rootCmd.AddCommand(backupCmd)
}
