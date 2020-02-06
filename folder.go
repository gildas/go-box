package box

import (
	"context"
	"encoding/json"
	"net/url"
	"net/http"
	"strings"
	"time"

	"github.com/gildas/go-core"
	"github.com/gildas/go-errors"
	"github.com/gildas/go-request"
)

// Folders module
type Folders struct {
	*Client
	api *url.URL
}

// FolderCollection represents a collection of FolderEntry
type FolderCollection struct {
	Count   int           `json:"total_count"`
	Folders []FolderEntry `json:"entries"`
}

// FolderEntry represents a File Entry
type FolderEntry struct {
	Type           string         `json:"type"`
	ID             string         `json:"id"`
	Name           string         `json:"name"`
	Description    string         `json:"description"`
	ETag           string         `json:"etag"`
	SequenceID     string         `json:"sequence_id"`
	Size           int64          `json:"size"`
	ItemStatus     string         `json:"item_status"`
	SharedLink     *SharedLink    `json:"shared_link,omitempty"`
	Checksum       string         `json:"sha1"`
	FileVersion    FileVersion    `json:"file_version"`
	Parent         *PathEntry     `json:"parent"`
	Paths          PathCollection `json:"path_collection"`
	ItemCollection PathCollection `json:"item_collection"`
	Tags           []string       `json:"tags"`
	SyncState      string         `json:"sync_state"`

	CreatedAt         time.Time `json:"-"`
	ModifiedAt        time.Time `json:"-"`
	TrashedAt         time.Time `json:"-"`
	PurgedAt          time.Time `json:"-"`
	ContentCreatedAt  time.Time `json:"-"`
	ContentModifiedAt time.Time `json:"-"`
	CreatedBy         UserEntry `json:"created_by"`
	ModifiedBy        UserEntry `json:"modified_by"`
	OwnedBy           UserEntry `json:"owned_by"`

	AllowedSharedLinkAccessLevels         []string `json:"allowed_shared_link_access_levels"`
	AllowedInviteeRoles                   []string `json:"allowed_invitee_roles"`
	HasCollaborations                     bool     `json:"has_collaborations"`
	CanNonOwnersInvite                    bool     `json:"can_non_owners_invite"`
	IsExternallyOwned                     bool     `json:"is_externally_owned"`
	IsCollaborationRestrictedToEnterprise bool     `json:"is_collaboration_restricted_to_enterprise"`
	//UploadEmail        nil || {access, email}   `json:"folder_upload_email,omitempty"`
	//watermark_info interface{}
	//metadata       interface{}
}

// AsPathEntry gets a PathEntry from the current FolderEntry
func (folder *FolderEntry) AsPathEntry() *PathEntry {
	return &PathEntry{
		Type:       "folder",
		ID:         folder.ID,
		Name:       folder.Name,
		ETag:       folder.ETag,
		SequenceID: folder.SequenceID,
		Checksum:   folder.Checksum,
	}
}

// Create creates a new folder
// entry.Name is mandatory, if entry.Parent.ID is not set the root folder is chosen
func (module *Folders) Create(ctx context.Context, entry *FolderEntry) (*FolderEntry, error) {
	// query: fields=comma-separated list of fields to include in the response
	if entry == nil || len(entry.Name) == 0 {
		return nil, errors.ArgumentMissing.With("name").WithStack()
	}
	if !module.Client.IsAuthenticated() {
		return nil, errors.Unauthorized.WithStack()
	}

	parentID := "0"
	if entry.Parent != nil && len(entry.Parent.ID) > 0 {
		parentID = entry.Parent.ID
	}

	result := FolderEntry{}
	_, err := module.Client.sendRequest(ctx, &request.Options{
		URL:     module.api,
		Payload: struct {
			Name   string    `json:"name"`
			Parent PathEntry `json:"parent"`
		}{entry.Name, PathEntry{ID: parentID}},
	}, &result)
	return &result, err
}

// Delete deletes a folder recursively
func (module *Folders) Delete(ctx context.Context, entry *FolderEntry) error {
	if entry == nil || len(entry.ID) == 0 {
		return errors.ArgumentMissing.With("ID").WithStack()
	}
	if !module.Client.IsAuthenticated() {
		return errors.Unauthorized.WithStack()
	}
	deleteURL, _ := module.api.Parse(entry.ID)
	_, err := module.Client.sendRequest(ctx, &request.Options{
		Method:     http.MethodDelete,
		URL:        deleteURL,
		Parameters: map[string]string{"recursive": "true"},
	}, nil)
	return err
}

// FindByID retrieves a folder by its id
func (module *Folders) FindByID(ctx context.Context, folderID string) (*FolderEntry, error) {
	// query: fields=comma-separated list of fields to include in the response
	if len(folderID) == 0 {
		return nil, errors.ArgumentMissing.With("ID").WithStack()
	}
	if !module.Client.IsAuthenticated() {
		return nil, errors.Unauthorized.WithStack()
	}

	findURL, _ := module.api.Parse(folderID)
	result := FolderEntry{}
	_, err := module.Client.sendRequest(ctx, &request.Options{
		URL: findURL,
	}, &result)
	return &result, err
}

// FindByName retrieves a folder by its name
// For now, exact match and 1 level (no recursion)
func (module *Folders) FindByName(ctx context.Context, name string) (*FolderEntry, error) {
	if len(name) == 0 {
		return nil, errors.ArgumentMissing.With("name").WithStack()
	}
	if !module.Client.IsAuthenticated() {
		return nil, errors.Unauthorized.WithStack()
	}

	// First get the root folder
	findURL, _ := module.api.Parse("0")
	root := FolderEntry{}
	if _, err := module.Client.sendRequest(ctx, &request.Options{
		URL: findURL,
	}, &root); err != nil {
		return nil, err
	}

	name = strings.ToLower(name)
	for _, item := range root.ItemCollection.Paths {
		if item.Type == "folder" && strings.ToLower(item.Name) == name { // Case insensitive search, option in the future?
			return module.FindByID(ctx, item.ID)
		}
	}
	return nil, errors.NotFound.With("folder", name).WithStack()
}

// MarshalJSON marshals this into JSON
func (folder FolderEntry) MarshalJSON() ([]byte, error) {
	type surrogate FolderEntry
	data, err := json.Marshal(struct {
		surrogate
		CA  core.Time `json:"created_at,omitempty"`
		MA  core.Time `json:"modified_at,omitempty"`
		TA  core.Time `json:"trashed_at,omitempty"`
		PA  core.Time `json:"purged_at,omitempty"`
		CCA core.Time `json:"content_created_at,omitempty"`
		CMA core.Time `json:"content_modified_at,omitempty"`
	}{
		surrogate: surrogate(folder),
		CA:        (core.Time)(folder.CreatedAt),
		MA:        (core.Time)(folder.ModifiedAt),
		TA:        (core.Time)(folder.TrashedAt),
		PA:        (core.Time)(folder.PurgedAt),
		CCA:       (core.Time)(folder.ContentCreatedAt),
		CMA:       (core.Time)(folder.ContentModifiedAt),
	})
	return data, errors.JSONMarshalError.Wrap(err)
}

// UnmarshalJSON decodes JSON
func (folder *FolderEntry) UnmarshalJSON(payload []byte) (err error) {
	type surrogate FolderEntry
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
	*folder = FolderEntry(inner.surrogate)
	folder.CreatedAt = (time.Time)(inner.CA)
	folder.ModifiedAt = (time.Time)(inner.MA)
	folder.TrashedAt = (time.Time)(inner.TA)
	folder.PurgedAt = (time.Time)(inner.PA)
	folder.ContentCreatedAt = (time.Time)(inner.CCA)
	folder.ContentModifiedAt = (time.Time)(inner.CMA)
	return
}
