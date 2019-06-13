package box

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"time"
)

// UploadOptions contains the options for uploading data
type UploadOptions struct {
	Parent      *PathEntry
	Filename    string
	ContentType string
	Content     []byte
}

// Upload uploads data to Box.com
func (client *Client) Upload(ctx context.Context, options *UploadOptions) (*FileCollection, error) {
	log := client.Logger.Scope("upload").Child()

	// TODO: Create real errors
	if options == nil {
		return nil, fmt.Errorf("Missing options")
	}
	if len(options.Content) == 0 {
		return nil, fmt.Errorf("Missing Content")
	}
	if len(options.Filename) == 0 {
		return nil, fmt.Errorf("Missing filename")
	}
	if client.Token == nil {
		return nil, fmt.Errorf("Not Authenticated")
	}

	if len(options.ContentType) == 0 {
		options.ContentType = "application/octet-stream"
	}

	reqBody := &bytes.Buffer{}
	writer  := multipart.NewWriter(reqBody)

	writer.WriteField("name", options.Filename)
	if options.Parent == nil {
		writer.WriteField("parent_id", "0")
	} else {
		writer.WriteField("parent_id", options.Parent.ID)
	}
	part, _ := writer.CreateFormFile("file", options.Filename)
	_, err  := part.Write(options.Content)
	if err != nil {
		return nil, fmt.Errorf("Failed to write content to the form: %s", err)
	}
	writer.Close()

	req, _ := http.NewRequest("POST", "https://upload.box.com/api/2.0/files/content", reqBody)
	req.Header.Add("User-Agent",    "BOX Client v." + VERSION)
	req.Header.Add("Content-Type",  writer.FormDataContentType())
	req.Header.Add("Authorization", "Bearer " + client.Token.AccessToken)

	httpclient := http.DefaultClient
	if client.Proxy != nil {
		httpclient.Transport = &http.Transport{Proxy: http.ProxyURL(client.Proxy)}
	}

	log.Debugf("HTTP POST %s", "https://api.box.com/oauth2/token")
	start    := time.Now()
	res, err := httpclient.Do(req)
	duration := time.Since(start)
	if err != nil {
		return nil, fmt.Errorf("Failed to send request to Box API: %s", err)
	}
	defer res.Body.Close()

	resBody, err := ioutil.ReadAll(res.Body) // read the body no matter what
	if err != nil {
		return nil, fmt.Errorf("Failed to read response body: %s", err)
	}
	log.Debugf("Response in %s\nproto: %s,\nstatus: %s,\nheaders: %#v", duration, res.Proto, res.Status, res.Header)
	log.Tracef("Response body: %s", string(resBody))

	if res.StatusCode >= 300 {
		// TODO: Handle token expiration
		return nil, fmt.Errorf("HTTP Error: %s", res.Status)
	}

	result := FileCollection{}
	if err = json.Unmarshal(resBody, &result); err != nil {
		return nil, fmt.Errorf("Failed to decode response body: %s", err)
	}
	return &result, nil
}