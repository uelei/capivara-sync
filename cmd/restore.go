/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"uelei/capivara-sync/db"
	"uelei/capivara-sync/sources"

	"github.com/spf13/cobra"

	"strings"
	"syscall"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
	"uelei/capivara-sync/handlers"

	log "github.com/sirupsen/logrus"
)

var list, clean bool
var snap string

// restoreCmd represents the restore command
var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore a backup from a snapshot",
	Long:  `Restore the files from the dest to a origin Source.`,
	Run: func(cmd *cobra.Command, args []string) {
		var error error
		var originauth, destauth ssh.AuthMethod

		database, err := db.InitDB("snapshot_files.db")
		if err != nil {
			log.Fatal(err)
		}
		defer database.Close()
		if list {
			fmt.Println("Listing snapshots")

			snaps, err := db.ListSnapShots(database)
			if err != nil {
				log.Fatal("Error listing snapshots:", err)
			}
			for _, snp := range snaps {
				fmt.Println("Snapshot ID:", snp.Id, "Date:", snp.Date)
			}

		} else {
			var destsource, originsource sources.Source

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
			handlers.Restore(originsource, destsource, snap, clean)
		}

		fmt.Println("restore called")
	},
}

func init() {
	restoreCmd.Flags().BoolVarP(&list, "list", "l", false, "List snapshots dates")
	restoreCmd.Flags().BoolVarP(&clean, "clean", "c", false, "List snapshots dates")

	restoreCmd.Flags().StringVarP(&origin, "origin", "o", "", "origin: local or ssh (required)")
	restoreCmd.Flags().StringVarP(&dest, "dest", "d", "", "destination: local or ssh (required)")

	// restoreCmd.MarkFlagRequired("origin")
	// restoreCmd.MarkFlagRequired("dest")

	restoreCmd.Flags().StringVarP(&snap, "snap", "s", "", "snap date to restore if not latest (optional)")

	restoreCmd.PersistentFlags().StringVar(&originpass, "origin-password", "", "SSH password (optional, will prompt if not provided)")
	restoreCmd.PersistentFlags().StringVar(&destpass, "dest-password", "", "SSH password (optional, will prompt if not provided)")

	rootCmd.AddCommand(restoreCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// restoreCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// restoreCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
