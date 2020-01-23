package box

import (
	"context"
	"encoding/json"
	"net/url"

	"github.com/gildas/go-errors"
	"github.com/gildas/go-request"
)

// UploadOptions contains the options for uploading data
type UploadOptions struct {
	Parent   *PathEntry
	Filename string
	Content  *request.Content
	Payload  interface{}
}

// Upload uploads data to Box.com
func (module *Files) Upload(ctx context.Context, options *UploadOptions) (*FileCollection, error) {
	//log := module.Client.Logger.Scope("upload")

	// TODO: Create real errors
	if options == nil {
		return nil, errors.ArgumentMissingError.With("options").WithStack()
	}

	if len(options.Filename) == 0 {
		return nil, errors.ArgumentMissingError.With("filename").WithStack()
	}
	if options.Payload != nil {
		payload, err := json.Marshal(options.Payload)
		if err != nil {
			return nil, errors.JSONMarshalError.Wrap(err)
		}
		options.Content = request.ContentWithData(payload, "application/json")
	}

	if options.Content == nil {
		return nil, errors.ArgumentMissingError.With("content").WithStack()
	}
	if !module.Client.IsAuthenticated() {
		return nil, errors.UnauthorizedError.WithStack()
	}

	parentID := "0"
	if options.Parent != nil && len(options.Parent.ID) > 0 {
		parentID = options.Parent.ID
	}

	uploadURL, _ := url.Parse("https://upload.box.com/api/2.0/files/content")
	results := FileCollection{}
	if _, err := module.Client.sendRequest(ctx, &request.Options{
		URL: uploadURL,
		Payload: map[string]string{
			"name":      options.Filename,
			"parent_id": parentID,
			">file":     options.Filename,
		},
		Attachment: options.Content.Reader(),
	}, &results); err != nil {
		return nil, err
	}
	return &results, nil
}
