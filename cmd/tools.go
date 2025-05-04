package cmd

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
	"strings"
	"syscall"
	"uelei/capivara-sync/sources"
)

func EnsureTrailingSlash(path string) string {
	if !strings.HasSuffix(path, "/") {
		return path + "/"
	}
	return path
}

func BuildSource(source_path string, password string, user string) (sources.Source, error) {

	if strings.Contains(source_path, "http") {
		host := ""
		if password == "" {
			fmt.Print("Enter password: ")
			bytePassword, err := term.ReadPassword(int(syscall.Stdin))
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println() // for newline
			password = string(bytePassword)

		}
		if strings.Contains(source_path, "@") && user == "" {

			sourceParts := strings.Split(source_path, "@")

			user = sourceParts[0]
			host = sourceParts[1]

		} else {
			if user == "" {
				log.Fatal("User not provided for SSH/DAV source")
			}
			host = source_path

		}

		return sources.NewWebDAVSource(host, user, password)

	}

	if strings.Contains(source_path, "@") {

		if password == "" {
			fmt.Print("Enter password: ")
			bytePassword, err := term.ReadPassword(int(syscall.Stdin))
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println() // for newline
			password = string(bytePassword)

		}
		auth := ssh.Password(password)
		originParts := strings.Split(source_path, "@")
		hostpaths := strings.Split(originParts[1], ":")
		originPath := EnsureTrailingSlash(hostpaths[1])
		return sources.NewSSHSource(originParts[0], hostpaths[0]+":22", originPath, auth)
	}

	return sources.Localsource{Localpath: EnsureTrailingSlash(source_path)}, nil
}
