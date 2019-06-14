package box

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gildas/go-core"
)

// SharedLinks module
type SharedLinks struct {
	*Client
}

// SharedLink represents a shared link
type SharedLink struct {
	URL               *core.URL   `json:"url"`
	DownloadURL       *core.URL   `json:"download_url"`
	VanityURL         *core.URL   `json:"vanity_url"`
	EffectiveAccess   string      `json:"effective_access"`
	IsPasswordEnabled bool        `json:"is_password_enabled"`
	UnSharedAt        *core.Time  `json:"unshared_at"`
	DownloadCount     int         `json:"download_count"`
	PreviewCount      int         `json:"preview_count"`
	Access            string      `json:"access"`
	Permissions       Permissions `json:"permissions"`
}

// Permissions exresses what is allowed on objects
type Permissions struct {
	CanDownload bool `json:"can_download,omitempty	"`
	CanPreview  bool `json:"can_preview,omitempty"`
}

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

// Create creates a shared link for a given File entry
func (module *SharedLinks) Create(ctx context.Context, entry *FileEntry, options *SharedLinkOptions) (*SharedLink, error) {
	//log := module.Client.Logger.Scope("createsharedlink").Child()
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
