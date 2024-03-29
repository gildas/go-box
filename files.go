package box

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"
	"time"

	"github.com/gildas/go-core"
	"github.com/gildas/go-errors"
	"github.com/gildas/go-request"
)

// Files module
type Files struct {
	*Client
	api *url.URL
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
	CreatedAt         time.Time      `json:"-"`
	ModifiedAt        time.Time      `json:"-"`
	TrashedAt         time.Time      `json:"-"`
	PurgedAt          time.Time      `json:"-"`
	ContentCreatedAt  time.Time      `json:"-"`
	ContentModifiedAt time.Time      `json:"-"`
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
	Count  int         `json:"total_count"`
	Offset int         `json:"offset,omitempty"`
	Limit  int         `json:"limit,omitempty"`
	Paths  []PathEntry `json:"entries"`
}

// PathEntry represents a Path Entry
type PathEntry struct {
	Type       string `json:"type"`
	ID         string `json:"id"`
	Name       string `json:"name"`
	ETag       string `json:"etag"`
	SequenceID string `json:"sequence_id"`
	Checksum   string `json:"sha1,omitempty"`
}

// UserEntry represents a User in a FileEntry
type UserEntry struct {
	Type  string `json:"type"`
	ID    string `json:"id"`
	Name  string `json:"name"`
	Login string `json:"login"`
}

// DownloadOptions contains the options for downloading data
type DownloadOptions struct {
	Parent      *PathEntry
	Filename    string
	ContentType string
	Content     []byte
	Payload     interface{}
}

// FindByID retrieves a file by its id
func (module *Files) FindByID(ctx context.Context, fileID string) (*FileEntry, error) {
	// query: fields=comma-separated list of fields to include in the response
	if len(fileID) == 0 {
		return nil, errors.ArgumentMissing.With("id")
	}
	if !module.Client.IsAuthenticated() {
		return nil, errors.Unauthorized.WithStack()
	}

	findURL, _ := module.api.Parse(fileID)
	result := FileEntry{}
	_, err := module.Client.sendRequest(ctx, &request.Options{URL: findURL}, &result)
	return &result, err
}

// FindByName retrieves a file by its name
// For now, exact match and 1 level (no recursion)
func (module *Files) FindByName(ctx context.Context, name string, parent *PathEntry) (*FileEntry, error) {
	if len(name) == 0 {
		return nil, errors.ArgumentMissing.With("filename")
	}
	if parent == nil || len(parent.ID) == 0 {
		return nil, errors.ArgumentMissing.With("parent")
	}
	if !module.Client.IsAuthenticated() {
		return nil, errors.Unauthorized.WithStack()
	}

	// First get the parent folder
	folder, err := module.Client.Folders.FindByID(ctx, parent.ID)
	if err != nil {
		return nil, errors.ArgumentInvalid.With("parent", parent.ID).(errors.Error).Wrap(err)
	}

	name = strings.ToLower(name)
	for _, item := range folder.ItemCollection.Paths {
		if item.Type == "file" && strings.ToLower(item.Name) == name { // Case insensitive search, option in the future?
			return module.FindByID(ctx, item.ID)
		}
	}
	return nil, errors.NotFound.With("filename", name)
}

// MarshalJSON marshals this into JSON
func (file FileEntry) MarshalJSON() ([]byte, error) {
	type surrogate FileEntry
	data, err := json.Marshal(struct {
		surrogate
		CA  core.Time `json:"created_at,omitempty"`
		MA  core.Time `json:"modified_at,omitempty"`
		TA  core.Time `json:"trashed_at,omitempty"`
		PA  core.Time `json:"purged_at,omitempty"`
		CCA core.Time `json:"content_created_at,omitempty"`
		CMA core.Time `json:"content_modified_at,omitempty"`
	}{
		surrogate: surrogate(file),
		CA:        (core.Time)(file.CreatedAt),
		MA:        (core.Time)(file.ModifiedAt),
		TA:        (core.Time)(file.TrashedAt),
		PA:        (core.Time)(file.PurgedAt),
		CCA:       (core.Time)(file.ContentCreatedAt),
		CMA:       (core.Time)(file.ContentModifiedAt),
	})
	return data, errors.JSONMarshalError.Wrap(err)
}

// UnmarshalJSON decodes JSON
func (file *FileEntry) UnmarshalJSON(payload []byte) (err error) {
	type surrogate FileEntry
	var inner struct {
		surrogate
		CA  core.Time `json:"created_at,omitempty"`
		MA  core.Time `json:"modified_at,omitempty"`
		TA  core.Time `json:"trashed_at,omitempty"`
		PA  core.Time `json:"purged_at,omitempty"`
		CCA core.Time `json:"content_created_at,omitempty"`
		CMA core.Time `json:"content_modified_at,omitempty"`
	}
	if err = json.Unmarshal(payload, &inner); err != nil {
		return errors.JSONUnmarshalError.Wrap(err)
	}
	*file = FileEntry(inner.surrogate)
	file.CreatedAt = (time.Time)(inner.CA)
	file.ModifiedAt = (time.Time)(inner.MA)
	file.TrashedAt = (time.Time)(inner.TA)
	file.PurgedAt = (time.Time)(inner.PA)
	file.ContentCreatedAt = (time.Time)(inner.CCA)
	file.ContentModifiedAt = (time.Time)(inner.CMA)
	return
}
