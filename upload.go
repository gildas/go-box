package box

import (
	"context"
	"fmt"
)

// UploadOptions contains the options for uploading data
type UploadOptions struct {
	Parent      *PathEntry
	Filename    string
	ContentType string
	Content     []byte
}

// Upload uploads data to Box.com
func (module *Files) Upload(ctx context.Context, options *UploadOptions) (*FileCollection, error) {
	//log := module.Client.Logger.Scope("upload").Child()

	// TODO: Create real errors
	if options == nil {
		return nil, fmt.Errorf("Missing options")
	}
	if len(options.Content) == 0 {
		return nil, fmt.Errorf("Missing Content")
	}
	if len(options.Filename) == 0 {
		return nil, fmt.Errorf("Missing filename")
	}
	if !module.Client.IsAuthenticated() {
		return nil, fmt.Errorf("Not Authenticated")
	}

	if len(options.ContentType) == 0 {
		options.ContentType = "application/octet-stream"
	}

	parentID := "0"
	if options.Parent != nil && len(options.Parent.ID) > 0 {
		parentID = options.Parent.ID
	}

	results := FileCollection{}
	if err := module.Client.sendRequest(ctx, &requestOptions{
		Method:     "POST",
		Path:       "https://upload.box.com/api/2.0/files/content", 
		Parameters: map[string]string{
			"name":      options.Filename,
			"parent_id": parentID,
			">file":     options.Filename,
		},
		Content:     options.Content,
		ContentType: options.ContentType,
	}, &results); err != nil {
		return nil, err
	}
	return &results, nil
}