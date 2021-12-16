package box

import (
	"context"

	"github.com/gildas/go-errors"
	"github.com/gildas/go-request"
)

// Download download the content of a file (by its FileEntry)
func (module *Files) Download(ctx context.Context, entry *FileEntry) (*request.Content, error) {
	// query: version=string to get a specific version
	if entry == nil || len(entry.ID) == 0 {
		return nil, errors.ArgumentMissing.With("entry")
	}
	if !module.Client.IsAuthenticated() {
		return nil, errors.Unauthorized.WithStack()
	}

	downloadURL, _ := module.api.Parse(entry.ID + "/content")
	return module.Client.sendRequest(ctx, &request.Options{
		URL: downloadURL,
	}, nil)
	// TODO: if statuscode is 202 accepted, res should have a header "Retry-After" to tell when we can download the file
	// Seems like HTTP 200 happens on small files
}
