package box

import (
	"strings"
	"context"
	"fmt"
	"github.com/gildas/go-core"
)

// Files module
type Files struct {
	*Client
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
	Checksum   string `json:"sha1,omitepmty"`
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
func(module *Files) FindByID(ctx context.Context, fileID string) (*FileEntry, error) {
	// query: fields=comma-separated list of fields to include in the response
	if len(fileID) == 0 {
		return nil, fmt.Errorf("Missing file ID")
	}
	if !module.Client.IsAuthenticated() {
		return nil, fmt.Errorf("Not Authenticated")
	}

	result := FileEntry{}
	if _, err := module.Client.sendRequest(ctx, &requestOptions{
		Method: "GET",
		Path:   "https://api.box.com/2.0/files/" + fileID,
	}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// FindByName retrieves a file by its name
// For now, exact match and 1 level (no recursion)
func (module *Files) FindByName(ctx context.Context, name string, parent *PathEntry) (*FileEntry, error) {
	if len(name) == 0 {
		return nil, fmt.Errorf("Missing file name")
	}
	if parent == nil || len(parent.ID) == 0 {
		return nil, fmt.Errorf("Missing parent ID")
	}
	if !module.Client.IsAuthenticated() {
		return nil, fmt.Errorf("Not Authenticated")
	}

	// First get the parent folder
	folder, err := module.Client.Folders.FindByID(ctx, parent.ID)
	if err != nil {
		return nil, err
	}

	name = strings.ToLower(name)
	for _, item := range folder.ItemCollection.Paths {
		if item.Type == "file" && strings.ToLower(item.Name) == name { // Case insensitive search, option in the future?
			return module.FindByID(ctx, item.ID)
		}
	}
	return nil, NotFoundError
}