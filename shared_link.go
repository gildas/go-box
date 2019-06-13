package box

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gildas/go-core"
)

// SharedLinkOptions contains the shared link options
type SharedLinkOptions struct {
	Access      string      `json:"access"`
	UnsharedAt  *core.Time  `json:"unshared_at,omitempty"`
	Password    string      `json:"password,omitempty"`
	Permissions Permissions `json:"permissions,omitempty"`
}

// MarshalJSON marshals this into JSON
func (slo SharedLinkOptions) MarshalJSON() ([]byte, error) {
	type surrogate SharedLinkOptions
	type sharedLink struct {
		surrogate
	}
	return json.Marshal(struct {
		SL sharedLink `json:"shared_link"`
	}{
		sharedLink{
			surrogate(slo),
		},
	})
}

// CreateSharedLink creates a shared link for a given File entry
func (client *Client) CreateSharedLink(ctx context.Context, entry *FileEntry, options *SharedLinkOptions) (*SharedLink, error) {
	log := client.Logger.Scope("createsharedlink").Child()
	if entry == nil {
		return nil, fmt.Errorf("Missing entry")
	}
	if options == nil {
		options = &SharedLinkOptions{
			Access:      "open",
			Permissions: Permissions{ CanDownload: true },
		}
	}

	// TODO: Create real errors
	if client.Token == nil {
		return nil, fmt.Errorf("Not Authenticated")
	}

	// TODO: Validate Access (open, company, collaborators)

	payload, err := json.Marshal(options)
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal sharedLink: %s", err)
	}
	log.Record("req", options).Tracef("Payload: %s", string(payload))
	req, _ := http.NewRequest("PUT", fmt.Sprintf("https://api.box.com/2.0/files/%s?fields=shared_link", entry.ID), bytes.NewBuffer(payload))
	req.Header.Add("User-Agent",    "BOX Client v." + VERSION)
	req.Header.Add("Content-Type",  "application/json")
	req.Header.Add("Authorization", "Bearer " + client.Token.AccessToken)

	httpclient := http.DefaultClient
	if client.Proxy != nil {
		httpclient.Transport = &http.Transport{Proxy: http.ProxyURL(client.Proxy)}
	}

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

	result := FileEntry{}
	if err = json.Unmarshal(resBody, &result); err != nil {
		return nil, fmt.Errorf("Failed to decode response body: %s", err)
	}
	log.Record("result", result).Tracef("Got result")
	return result.SharedLink, nil
}
