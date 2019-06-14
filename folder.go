package box

import (
	"strings"
	"fmt"
	"context"
	"github.com/gildas/go-core"
)

// Folders module
type Folders struct {
	*Client
}

// FolderCollection represents a collection of FolderEntry
type FolderCollection struct {
	Count   int           `json:"total_count"`
	Folders []FolderEntry `json:"entries"`
}

// FolderEntry represents a File Entry
type FolderEntry struct {
	Type               string         `json:"type"`
	ID                 string         `json:"id"`
	Name               string         `json:"name"`
	Description        string         `json:"description"`
	ETag               string         `json:"etag"`
	SequenceID         string         `json:"sequence_id"`
	Size               int64          `json:"size"`
	ItemStatus         string         `json:"item_status"`
	SharedLink         *SharedLink    `json:"shared_link,omitempty"`
	Checksum           string         `json:"sha1"`
	FileVersion        FileVersion    `json:"file_version"`
	Parent             PathEntry      `json:"parent"`
	Paths              PathCollection `json:"path_collection"`
	ItemCollection     PathCollection `json:"item_collection"`
	Tags               []string       `json:"tags"`
	SyncState          string         `json:"sync_state"`

	CreatedAt          core.Time      `json:"created_at,omitempty"`
	ModifiedAt         core.Time      `json:"modified_at,omitempty"`
	TrashedAt          core.Time      `json:"trashed_at,omitempty"`
	PurgedAt           core.Time      `json:"purged_at,omitempty"`
	ContentCreatedAt   core.Time      `json:"content_created_at,omitempty"`
	ContentModifiedAt  core.Time      `json:"content_modified_at,omitempty"`
	CreatedBy          UserEntry      `json:"created_by"`
	ModifiedBy         UserEntry      `json:"modified_by"`
	OwnedBy            UserEntry      `json:"owned_by"`

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
func (folder *FolderEntry) AsPathEntry() (*PathEntry) {
	return &PathEntry {
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
		return nil, fmt.Errorf("Missing folder name")
	}
	if !module.Client.IsAuthenticated() {
		return nil, fmt.Errorf("Not Authenticated")
	}

	parentID := "0"
	if len(entry.Parent.ID) > 0 {
		parentID = entry.Parent.ID
	}

	result := FolderEntry{}
	if err := module.Client.sendRequest(ctx, &requestOptions{
		Method: "POST",
		Path:   "https://api.box.com/2.0/folders",
		Payload: struct {
			Name   string    `json:"name"`
			Parent PathEntry `json:"parent"`
		}{entry.Name, PathEntry{ID: parentID}},
	}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete deletes a folder recursively
func (module *Folders) Delete(ctx context.Context, entry *FolderEntry) error {
	if entry == nil || len(entry.ID) == 0 {
		return fmt.Errorf("Missing folder ID")
	}
	if !module.Client.IsAuthenticated() {
		return fmt.Errorf("Not Authenticated")
	}
	return module.Client.sendRequest(ctx, &requestOptions{
		Method: "DELETE",
		Path:   "https://api.box.com/2.0/folders/"+entry.ID+"?recursive=true",
	}, nil)
}

// FindByID retrieves a folder by its id
func(module *Folders) FindByID(ctx context.Context, folderID string) (*FolderEntry, error) {
	// query: fields=comma-separated list of fields to include in the response
	if len(folderID) == 0 {
		return nil, fmt.Errorf("Missing folder ID")
	}
	if !module.Client.IsAuthenticated() {
		return nil, fmt.Errorf("Not Authenticated")
	}

	result := FolderEntry{}
	if err := module.Client.sendRequest(ctx, &requestOptions{
		Method: "GET",
		Path:   "https://api.box.com/2.0/folders/" + folderID,
	}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// FindByName retrieves a folder by its name
// For now, exact match and 1 level (no recursion)
func (module *Folders) FindByName(ctx context.Context, name string) (*FolderEntry, error) {
	if len(name) == 0 {
		return nil, fmt.Errorf("Missing folder name")
	}
	if !module.Client.IsAuthenticated() {
		return nil, fmt.Errorf("Not Authenticated")
	}

	// First get the root folder
	root := FolderEntry{}
	if err := module.Client.sendRequest(ctx, &requestOptions{
		Method: "GET",
		Path:   "https://api.box.com/2.0/folders/0",
	}, &root); err != nil {
		return nil, err
	}

	name = strings.ToLower(name)
	for _, item := range root.ItemCollection.Paths {
		if item.Type == "folder" && strings.ToLower(item.Name) == name { // Case insensitive search, option in the future?
			return module.FindByID(ctx, item.ID)
		}
	}
	return nil, NotFoundError
}