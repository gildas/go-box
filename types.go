package box

import (
	"github.com/gildas/go-core"
)

// Auth represents the Authentication information
type Auth struct {
	ClientID     string  `json:"clientID"`
	ClientSecret string  `json:"clientSecret"`
	AppAuth      AppAuth `json:"appAuth"`
	EnterpriseID string  `json:"enterpriseID"`
}

// AppAuth is used to authenticate an application
type AppAuth struct {
	PublicKeyID string `json:"publickeyID"`
	PrivateKey  string `json:"privateKey"`
	Passphrase  string `json:"passphrase"`
}

// Token is the token used to send requests to Box.com
type Token struct {
	TokenType    string   `json:"token_type"`
	AccessToken  string   `json:"access_token"`
	ExpiresIn    int64    `json:"expires_in"` // TODO: We should transcript this in ExpiresAt time.Time
	RestrictedTo []string `json:"restricted_to"`
}

// FileCollection represents a collection of FileEntry
type FileCollection struct {
	Count int         `json:"total_count"`
	Files []FileEntry `json:"entries"`
}

// FileEntry represents a File Entry
type FileEntry struct {
	Type              string         `json:"type"`
	ID                string         `json:"id"`
	Name              string         `json:"name"`
	Description       string         `json:"description"`
	ETag              string         `json:"etag"`
	SequenceID        string         `json:"sequence_id"`
	Size              int64          `json:"size"`
	ItemStatus        string         `json:"item_status"`
	SharedLink        *SharedLink    `json:"shared_link,omitempty"`
	Checksum          string         `json:"sha1"`
	FileVersion       FileVersion    `json:"file_version"`
	Parent            PathEntry      `json:"parent"`
	Paths             PathCollection `json:"path_collection"`
	CreatedAt         core.Time      `json:"created_at,omitempty"`
	ModifiedAt        core.Time      `json:"modified_at,omitempty"`
	TrashedAt         core.Time      `json:"trashed_at,omitempty"`
	PurgedAt          core.Time      `json:"purged_at,omitempty"`
	ContentCreatedAt  core.Time      `json:"content_created_at,omitempty"`
	ContentModifiedAt core.Time      `json:"content_modified_at,omitempty"`
	CreatedBy         UserEntry      `json:"created_by"`
	ModifiedBy        UserEntry      `json:"modified_by"`
	OwnedBy           UserEntry      `json:"owned_by"`
}

// FileVersion represents the version of a FileEntry
type FileVersion struct {
	Type     string `json:"type"`
	ID       string `json:"id"`
	Checksum string `json:"sha1"`
}

// PathCollection represents a collection of PathEntry
type PathCollection struct {
	Count int         `json:"total_count"`
	Paths []PathEntry `json:"entries"`
}

// PathEntry represents a Path Entry
type PathEntry struct {
	Type       string `json:"type"`
	ID         string `json:"id"`
	Name       string `json:"name"`
	ETag       string `json:"etag"`
	SequenceID string `json:"sequence_id"`
}

// UserEntry represents a User in a FileEntry
type UserEntry struct {
	Type  string `json:"type"`
	ID    string `json:"id"`
	Name  string `json:"name"`
	Login string `json:"login"`
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