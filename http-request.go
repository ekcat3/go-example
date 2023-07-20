package example

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"regexp"
	"strings"
)

// Make a request
func Request(method string, apiUrl string, js interface{}, wantStatus []int, headers map[string]string) (resp *http.Response, err error) {
	var jb []byte
	if jb, err = json.Marshal(js); err != nil {
		return
	}

	var req *http.Request
	if req, err = http.NewRequest(method, apiUrl, bytes.NewBuffer(jb)); err != nil {
		return nil, err
	}

	if len(headers) == 0 {
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Accept", "application/json")
	} else {
		for k, v := range headers {
			req.Header.Add(k, v)
		}
	}

	if os.Getenv("DEBUG") == "true" {
		var db []byte
		if db, err = httputil.DumpRequest(req, true); err != nil {
			return
		}
		log.Print("REQUEST:\n", string(db))
	}
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	if os.Getenv("DEBUG") == "true" {
		if body, err := httputil.DumpResponse(resp, true); err == nil {
			log.Print("REPONSE:\n", string(body))
		}
	}

	var statusFound bool
	for _, status := range wantStatus {
		if resp.StatusCode == status {
			statusFound = true
			break
		}
	}
	if !statusFound {
		var body []byte
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			return
		}
		if strings.Contains(string(body), "<html><head>") {
			r := regexp.MustCompile(`(<title>)([^{}]*)(<\/title>)`)
			title := r.FindAllStringSubmatch(string(body), -1)
			if len(title) == 0 {
				err = fmt.Errorf("got: %d, wanted: %v", resp.StatusCode, wantStatus)
			} else {
				err = fmt.Errorf("got: %d, wanted: %v (%s)", resp.StatusCode, wantStatus, title[0][2])
			}

		} else {
			err = fmt.Errorf("got: %d, wanted: %v (%s)", resp.StatusCode, wantStatus, body)
		}
		return
	}
	return
}
