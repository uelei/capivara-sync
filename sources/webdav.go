package sources

import (
	"bytes"
	"crypto/sha1"
	"encoding/xml"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

type WebDAVSource struct {
	Server   string
	Username string
	Password string
}

func NewWebDAVSource(server, username, password string) (*WebDAVSource, error) {
	return &WebDAVSource{
		Server:   server,
		Username: username,
		Password: password,
	}, nil
}

func (w *WebDAVSource) CalculateFileHash(filebyte []byte) (string, error) {
	hash := sha1.New()
	hash.Write(filebyte)
	sha1Sum := hash.Sum(nil)
	sha1Hex := fmt.Sprintf("%x", sha1Sum)
	return sha1Hex, nil
}

func ConvertTimeFromRFC3339(timeString string) (time.Time, error) {

	parsedTime, err := time.Parse(time.RFC1123, timeString)
	if err != nil {
		fmt.Printf("Error parsing time: %v\n", err)
		return time.Time{}, err
	}

	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		fmt.Printf("Error loading time zone: %v\n", err)
		return time.Time{}, err
	}

	localTime := parsedTime.In(loc)
	return localTime, nil
	// formattedTime := localTime.Format("2006-01-02 15:04:05")
	// return formattedTime, nil

}

func (w *WebDAVSource) ListFiles() <-chan FileInfo {
	ch := make(chan FileInfo)
	go func() {
		defer close(ch)
		// Create an HTTP request to list files
		req, err := http.NewRequest("PROPFIND", w.Server, nil)
		if err != nil {
			log.Error("Error creating request: ", err)
			return
		}
		req.SetBasicAuth(w.Username, w.Password)
		req.Header.Set("Depth", "1000")

		// Perform the HTTP request
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Error("Error performing request: ", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusMultiStatus {
			log.Error("Unexpected status code: ", resp.StatusCode)
			return
		}

		// Parse the XML response
		var multistatus struct {
			Responses []struct {
				Href  string `xml:"href"`
				Props struct {
					DisplayName string `xml:"displayname"`
					Permissions string `xml:"getcontentlength"`
				} `xml:"propstat>prop"`
			} `xml:"response"`
		}
		if err := xml.NewDecoder(resp.Body).Decode(&multistatus); err != nil {
			log.Error("Error decoding response: ", err)
			return
		}

		basePath := multistatus.Responses[0].Href // Assuming the first response contains the base path

		log.Info("Base path: ", basePath)
		// Send FileInfo objects to the channel
		for _, response := range multistatus.Responses {

			if strings.HasSuffix(response.Href, "/") {
				log.Debug("Skipping directory: ", response.Href)
			} else {

				remote_path := strings.TrimPrefix(response.Href, basePath)
				// Perform a HEAD request to get the checksum

				remote_path, err := url.QueryUnescape(remote_path)
				if err != nil {
					fmt.Printf("Error decoding string: %v\n", err)
					return
				}

				md5, error := w.GetFileHash(remote_path)
				if error != nil {
					log.Error("Error getting file hash:", error)
				}
				LastModified, error := w.GetFileLastModified(remote_path)
				if error != nil {
					log.Error("Error getting file last modified:", error)
				}
				// LastModifiedTm, error := ConvertTimeFromRFC3339(LastModified)
				// if error != nil {
				// 	log.Error("Error converting file last modified:", error)
				// }

				ch <- FileInfo{
					Path:         remote_path,
					Md5:          md5,
					Filename:     path.Base(remote_path),
					Permission:   response.Props.Permissions,
					RemoteHash:   "",
					LastModified: LastModified,
				}
			}
		}
	}()
	return ch
}

func (w *WebDAVSource) GetFile(path string) ([]byte, error) {
	log.Warn("Performing GET request for: ", path, " on ", w.Server)
	req, err := http.NewRequest("GET", w.Server+path, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(w.Username, w.Password)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to fetch file")
	}

	return io.ReadAll(resp.Body)
}

func (w *WebDAVSource) SaveFile(path string, data []byte, permission string) error {
	log.Info("Saving file to WebDAV: ", w.Server, "pall  ", path)
	body := bytes.NewReader(data)
	req, err := http.NewRequest("PUT", w.Server+path, body)
	if err != nil {
		log.Error("Error creating request:", err)
		return err
	}
	req.SetBasicAuth(w.Username, w.Password)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error("Error sending request:", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		log.Error("Error saving file:", resp.Status)
		return errors.New("failed to save file")
	}

	return nil
}

func (w *WebDAVSource) Exists(path string) bool {
	req, err := http.NewRequest("HEAD", w.Server+path, nil)
	if err != nil {
		return false
	}
	req.SetBasicAuth(w.Username, w.Password)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// DAVResponse represents the WebDAV multistatus response
type DAVResponse struct {
	XMLName   xml.Name   `xml:"multistatus"`
	Responses []Response `xml:"response"`
}

// Response represents a single resource response
type Response struct {
	Href     string   `xml:"href"`
	Propstat Propstat `xml:"propstat"`
}

// Propstat contains the properties and status
type Propstat struct {
	Prop   Prop   `xml:"prop"`
	Status string `xml:"status"`
}

// Prop contains the requested properties
type Prop struct {
	GetLastModified string    `xml:"getlastmodified"`
	Checksums       Checksums `xml:"http://owncloud.org/ns checksums"`
}

// Checksums contains the checksum entries
type Checksums struct {
	Checksum []string `xml:"http://owncloud.org/ns checksum"`
}

func (w *WebDAVSource) GetFileHash(remote_path string) (string, error) {
	client := &http.Client{}
	body := `<?xml version="1.0" encoding="utf-8" ?>
		<d:propfind xmlns:d="DAV:" xmlns:oc="http://owncloud.org/ns">
			<d:prop><oc:checksums/></d:prop>
		</d:propfind>`
	headReq, err := http.NewRequest("PROPFIND", w.Server+remote_path, strings.NewReader(body))
	if err != nil {
		log.Error("Error creating PROPFIND request: ", err)
		return "", err
	}
	headReq.Header.Set("Content-Type", "application/xml")
	headReq.SetBasicAuth(w.Username, w.Password)

	headResp, err := client.Do(headReq)
	if err != nil {
		log.Error("Error performing PROPFIND request: ", err)
		return "", err
	}
	defer headResp.Body.Close()
	if headResp.StatusCode != http.StatusMultiStatus {
		return "", fmt.Errorf("unexpected status: %s", headResp.Status)
	}

	// Parse XML response
	var davResp DAVResponse
	if err := xml.NewDecoder(headResp.Body).Decode(&davResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Extract MD5 checksum
	for _, response := range davResp.Responses {
		for _, checksum := range response.Propstat.Prop.Checksums.Checksum {
			for _, csum := range strings.Split(checksum, " ") {
				if strings.HasPrefix(csum, "MD5:") {
					return strings.TrimPrefix(csum, "MD5:"), nil
				}
			}
		}
	}

	return "", fmt.Errorf("MD5 checksum not found")
}

func (w *WebDAVSource) RemoveFile(path string) error {
	req, err := http.NewRequest("DELETE", w.Server+path, nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(w.Username, w.Password)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return errors.New("failed to remove file")
	}

	return nil
}

func (w *WebDAVSource) GetFileLastModified(remote_path string) (time.Time, error) {
	client := &http.Client{}
	body := `<?xml version="1.0" encoding="utf-8" ?>
		<d:propfind xmlns:d="DAV:">
			<d:prop><d:getlastmodified/></d:prop>
		</d:propfind>`
	req, err := http.NewRequest("PROPFIND", w.Server+remote_path, strings.NewReader(body))
	if err != nil {
		log.Error("Error creating PROPFIND request: ", err)
		return time.Time{}, err
	}
	req.Header.Set("Content-Type", "application/xml")
	req.SetBasicAuth(w.Username, w.Password)

	resp, err := client.Do(req)
	if err != nil {
		log.Error("Error performing PROPFIND request: ", err)
		return time.Time{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMultiStatus {
		return time.Time{}, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	// Parse XML response
	var davResp DAVResponse
	if err := xml.NewDecoder(resp.Body).Decode(&davResp); err != nil {
		return time.Time{}, fmt.Errorf("failed to parse response: %w", err)
	}

	// Extract last modified time
	for _, response := range davResp.Responses {
		if response.Propstat.Prop.GetLastModified != "" {
			parsedTime, err := time.Parse(time.RFC1123, response.Propstat.Prop.GetLastModified)
			if err != nil {
				fmt.Printf("Error parsing time: %v\n", err)
				return time.Time{}, err
			}
			return parsedTime, nil
		}
	}

	return time.Time{}, fmt.Errorf("last modified time not found")
}
