package box

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/gildas/go-core"
)

// SharedLinks module
type SharedLinks struct {
	*Client
}

// SharedLink represents a shared link
type SharedLink struct {
	URL               *url.URL    `json:"-"`
	DownloadURL       *url.URL    `json:"-"`
	VanityURL         *url.URL    `json:"-"`
	EffectiveAccess   string      `json:"effective_access"`
	IsPasswordEnabled bool        `json:"is_password_enabled"`
	UnsharedAt        *time.Time  `json:"-"`
	DownloadCount     int         `json:"download_count"`
	PreviewCount      int         `json:"preview_count"`
	Access            string      `json:"access"`
	Permissions       Permissions `json:"permissions"`
}

// Permissions exresses what is allowed on objects
type Permissions struct {
	CanDownload bool `json:"can_download,omitempty"`
	CanPreview  bool `json:"can_preview,omitempty"`
}

// SharedLinkOptions contains the shared link options
type SharedLinkOptions struct {
	Access      string      `json:"access"`
	UnsharedAt  *time.Time  `json:"-"`
	Password    string      `json:"password,omitempty"`
	Permissions Permissions `json:"permissions,omitempty"`
}

// Create creates a shared link for a given File entry
func (module *SharedLinks) Create(ctx context.Context, entry *FileEntry, options *SharedLinkOptions) (*SharedLink, error) {
	//log := module.Client.Logger.Scope("createsharedlink")
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
	if !module.Client.IsAuthenticated() {
		return nil, fmt.Errorf("Not Authenticated")
	}

	// TODO: Validate Access (open, company, collaborators)

	result := FileEntry{}
	if _, err := module.Client.sendRequest(ctx, &requestOptions{
		Method:  "PUT",
		Path:    fmt.Sprintf("https://api.box.com/2.0/files/%s?fields=shared_link", entry.ID),
		Payload: *options,
	}, &result); err != nil {
		return nil ,err
	}
	return result.SharedLink, nil
}

// MarshalJSON marshals this into JSON
func (slo SharedLinkOptions) MarshalJSON() ([]byte, error) {
	type surrogate SharedLinkOptions
	type sharedLink struct {
		surrogate
		UA *core.Time `json:"unshared_at,omitempty"`
	}
	return json.Marshal(struct {
		SL sharedLink `json:"shared_link"`
	}{
		sharedLink{
			surrogate(slo),
			(*core.Time)(slo.UnsharedAt),
		},
	})
}

// MarshalJSON marshals this into JSON
func (link SharedLink) MarshalJSON() ([]byte, error) {
	type surrogate SharedLink
	return json.Marshal(struct {
		surrogate
		U  *core.URL  `json:"url"`
		DU *core.URL  `json:"download_url"`
		VU *core.URL  `json:"vanity_url"`
		UA *core.Time `json:"unshared_at"`
	}{
		surrogate: surrogate(link),
		U:  (*core.URL)(link.URL),
		DU: (*core.URL)(link.DownloadURL),
		VU: (*core.URL)(link.VanityURL),
		UA: (*core.Time)(link.UnsharedAt),
	})
}

// UnmarshalJSON decodes JSON
func (link *SharedLink) UnmarshalJSON(payload []byte) (err error) {
	type surrogate SharedLink
	var inner struct {
		surrogate
		U  *core.URL  `json:"url"`
		DU *core.URL  `json:"download_url"`
		VU *core.URL  `json:"vanity_url"`
		UA *core.Time `json:"unshared_at"`
	}
	if err = json.Unmarshal(payload, &inner); err != nil {
		return err
	}
	*link = SharedLink(inner.surrogate)
	link.URL         = (*url.URL)(inner.U)
	link.DownloadURL = (*url.URL)(inner.DU)
	link.VanityURL   = (*url.URL)(inner.VU)
	link.UnsharedAt  = (*time.Time)(inner.UA)
	return
}