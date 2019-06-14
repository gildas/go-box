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
	if err := module.Client.sendRequest(ctx, &requestOptions{
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

// Download download the content of a file (by its FileEntry)
func (module *Files) Download(ctx context.Context, entry *FileEntry) ([]byte, error) {
	// query: version=string to get a specific version
	if entry == nil || len(entry.ID) == 0 {
		return nil, fmt.Errorf("Missing file ID")
	}
	if !module.Client.IsAuthenticated() {
		return nil, fmt.Errorf("Not Authenticated")
	}

	err := module.Client.sendRequest(ctx, &requestOptions{
		Method: "GET",
		Path:   "https://api.box.com/2.0/files/"+entry.ID+"/content",
	}, nil)

	if boxerr, ok := err.(RequestError); ok && boxerr.StatusCode == 302 {
		// Now we can download
/*
curl -v https://api.box.com/2.0/files/475520294617/content -H "Authorization: Bearer 5tjlEa98nBI6kyN0rPJUCD1g6d95kzs5"
> GET /2.0/files/475520294617/content HTTP/1.1
> Host: api.box.com
> User-Agent: curl/7.54.0
> Authorization: Bearer 5tjlEa98nBI6kyN0rPJUCD1g6d95kzs5
>
< HTTP/1.1 302 Found
< Date: Fri, 14 Jun 2019 09:23:30 GMT
< Transfer-Encoding: chunked
< Connection: keep-alive
< Strict-Transport-Security: max-age=31536000
< Cache-Control: no-cache, no-store
< Vary: Accept-Encoding
< BOX-REQUEST-ID: 0s9vosihvdq6e6pjrkbek8l8vbc
< Location: https://public.boxcloud.com/d/1/b1!SSzrPW6cHGnDSw_XWpsml0wbpBzTIoGopb8XjBoL9prMGyDoknZHfOe8PKD_A0QS98-oF6WVsXkZdvVrI1gaVi-1SAqJ57d3_C5WXwAGl-AwpniD8DDq447Mp8M2Lr-xkLSi6z1zqwGhghXHNS8vqZ0NZotaYrPV-LlIFubxlUQ7f_1ckSz36UY8aSViPV2ItWu6yT9_EZE2jNlFWtN1EvMrSxgu5ES1gtShKoJgUz00_jeWp4PJVfuCm8DVqzwJZirDYR0yVV9hsd9k1NRkgLTo3A5VSB50sPrU4TsLJY4-Ks6vk1MPZ-h6wlB3rNaREdsSGdiVCJRntf6QkALaL4xhdFkjv42aajJ7TDioD-LK7Z7Nv3GP8uT933GSnh5r3a2rpDy48l1Ucr4GeE6luSrjX0flQR0MEJjyabsEN52rwnjOFF9rPgi9wSo0m0I959NLtziQ0XxAs_AWAl3O1k26xJHxiI_XWXjH5ObLaZ5sJSd8rcFEC1qRhNpOhfGFwCO5WUyAfoBxGgtA454WPtLyWwTAXz2Sqw3BHYJ2yNjxXZ1VkbzQoNXSll_p3w3x9h4I74TMoUzWi2dcpJpTzsnnLQxtNt7y00v-I52BhEQm6-exxWBWyKk1xs4aXYLoxACyFIutFpExXPMtN1x0zlauyvbrHaPXS7-TWanQa8h5CFq6AyDWHD76-DbyqY-4cA7KJkE69Kk9OEh18p6Xnmw72GZtElEM0GtynfHgR1ZL51XjqzV-q8pSn9z1OYhG7u6NdU6YBUZuZ0vRnrt4Ddoq4s7mg8MReIjBT3gMHgVDWqW5qRD7w4zJTRV3K9_EUMJPLXTimcXlzKw6GVCi7v6PWVZAzAoU5iofdhUfgKePquqs70TjPe7RrOvFxs57GPHA_gx8rvUozjlAWSxoYxe8tNf00LhH4jgDPfDWA16qemQsGklJop4etkToarUJ5WpidKHIuqOetdhxTo0a5g9IXOqoU1aycYsGVc4Fw2xqyJobUsLeLoFZidMg-vDj5GztLVU9Y2j_I8xyGcxZD8TRIqXJVXo6hc9PxHdnq43DpNaqVreQeP3Ljj0DNkHmHH93KEst3VCWujFgh5RJlZAqf9RLnT6QWmf2lj8b6b1SDJbJPFt1coFCZQ2FBM7UYfCyBRf5UYUsqUbM0kvMnf0A7cGyoTd6JcecBaHVLQuwuRhnRj7JfeFpreWdlqFahVWdR1rqicZRdBz86LG2RfMAlL7zAfa5qlE2e8ubF6ZHoyv8Jx43YB2ScPdf_P6SOX8ghTde0V4SasyPOB0Q07TJkgA_KKHPW0-HHxw9tA../download
*/
		return []byte{}, nil
	}
	// TODO: if statuscode is 202 accepted, res should have a header "Retry-After" to tell when we can download the file
	if err != nil {
		return nil, err
	}
	// Seems like HTTP 200 is not possible to receive
	return nil, nil
}