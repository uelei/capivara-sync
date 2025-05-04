package sources

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/pkg/sftp"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"io"
	"path"
	"strings"
	"time"
)

type SSHSource struct {
	Client   *ssh.Client
	SFTP     *sftp.Client
	BasePath string
}

func NewSSHSource(user, addr, basePath string, authMethod ssh.AuthMethod) (*SSHSource, error) {
	config := &ssh.ClientConfig{
		User:            user,
		Auth:            []ssh.AuthMethod{authMethod},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // For production, replace with proper callback
		Timeout:         10 * time.Second,
	}

	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, fmt.Errorf("SSH connection failed: %w", err)
	}

	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return nil, fmt.Errorf("SFTP client init failed: %w", err)
	}

	return &SSHSource{
		Client:   client,
		SFTP:     sftpClient,
		BasePath: basePath,
	}, nil
}

func (s *SSHSource) GetFileHash(path string) (string, error) {
	session, err := s.Client.NewSession()
	if err != nil {
		log.Error("Failed to create SSH session:", err)
		return "", err
	}
	defer session.Close()

	// Escape the path to avoid issues with special characters
	escapedPath := fmt.Sprintf("%q", s.BasePath+path)

	output, err := session.CombinedOutput(fmt.Sprintf("md5sum %s", escapedPath))
	if err != nil {
		log.Errorf("Failed to execute md5sum command on path '%s': %v", path, err)
		log.Debugf("Command output: %s", string(output))
		return "", err
	}

	// md5sum output format: "<hash>  <filename>"
	parts := strings.Fields(string(output))
	if len(parts) == 0 {
		return "", fmt.Errorf("unexpected md5sum output: %s", output)
	}

	return parts[0], nil
}

func (s *SSHSource) ListFiles() <-chan FileInfo {
	ch := make(chan FileInfo)
	go func() {
		defer close(ch)
		walker := s.SFTP.Walk(s.BasePath)
		for walker.Step() {
			if err := walker.Err(); err != nil {
				continue
			}
			stat := walker.Stat()
			// Get last modification time
			modTime := stat.ModTime()
			ch <- FileInfo{
				Path:         walker.Path(),
				Permission:   stat.Mode().Perm().String(),
				LastModified: modTime,
			}
		}
	}()
	return ch
}

func (s *SSHSource) GetFile(path string) ([]byte, error) {
	f, err := s.SFTP.Open(s.BasePath + path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return io.ReadAll(f)
}

func ensureRemoteDir(sftpClient *sftp.Client, remotePath string) error {
	dirs := strings.Split(path.Clean(path.Dir(remotePath)), "/")
	curr := "/"
	for _, dir := range dirs {
		if dir == "" {
			continue
		}
		curr = path.Join(curr, dir)
		_, err := sftpClient.Stat(curr)
		if err != nil {
			if err := sftpClient.Mkdir(curr); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *SSHSource) SaveFile(path string, data []byte, perm string) error {
	filePath := s.BasePath + path

	if err := ensureRemoteDir(s.SFTP, filePath); err != nil {
		return fmt.Errorf("failed to create remote dirs: %w", err)
	}

	f, err := s.SFTP.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	log.Debug("Writing file to remote:", filePath)
	if _, err := f.Write(data); err != nil {
		return err
	}

	// Step 3: Apply permissions if provided
	permission, err := FileModeFromString(perm)
	if err != nil {
		log.Error("Error parsing permission string:", perm, err)
	}
	if err := s.SFTP.Chmod(filePath, permission); err != nil {
		return fmt.Errorf("failed to chmod remote file: %w", err)
	}
	// Optionally apply permissions using sftp.Chmod
	return nil
}

func (s *SSHSource) Exists(path string) bool {
	_, err := s.SFTP.Stat(s.BasePath + path)
	return err == nil
}

func (s *SSHSource) RemoveFile(path string) error {
	return s.SFTP.Remove(s.BasePath + path)
}

func (s *SSHSource) CalculateFileHash(filebyte []byte) (string, error) {
	hash := md5.Sum(filebyte) // returns [16]byte
	remote_hash := hex.EncodeToString(hash[:])

	return remote_hash, nil
}

func (s *SSHSource) GetFileLastModified(remote_path string) (time.Time, error) {
	session, err := s.Client.NewSession()
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	// Command to get the last modified time of the file
	cmd := fmt.Sprintf("stat -c %%y %s", remote_path)
	println("cmd : ", cmd)
	output, err := session.Output(cmd)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to execute command: %w", err)
	}
	result_date := strings.TrimSpace(string(output))

	const layout = "2006-01-02 15:04:05.999999999 -0700"
	parsedTime, err := time.Parse(layout, result_date)
	if err != nil {
		fmt.Printf("Error parsing time: %v\n", err)
		return time.Time{}, err
	}
	// Return the output as the last modified time
	return parsedTime, nil

}
