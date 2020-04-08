package box

import (
	"context"

	"github.com/gildas/go-errors"
	"github.com/gildas/go-request"
)

// sendRequest sends an HTTP request to Box.com's API
func (client *Client) sendRequest(ctx context.Context, options *request.Options, results interface{}) (*request.ContentReader, error) {
	if options == nil {
		return nil, errors.ArgumentMissing.With("options").WithStack()
	}
	options.Context = ctx
	options.Logger = client.Logger
	options.UserAgent = "BOX Client " + VERSION
	if client.IsAuthenticated() {
		options.Authorization = request.BearerAuthorization(client.Auth.Token.AccessToken)
	}

	response, err := request.Send(options, results)

	// TODO: We need to get access to the response headers
	// boxRequestID := res.Header.Get("Box-Request-Id")

	if err != nil {
		var details *RequestError
		if jerr := response.UnmarshalContentJSON(&details); jerr == nil {
			var httperr *errors.Error
			if errors.As(err, &httperr) {
				details.StatusCode = httperr.Code
			}
			if errors.Is(err, errors.HTTPBadRequest) && errors.Is(details, InvalidGrant) {
				return nil, errors.Unauthorized.Wrap(details)
			}
			if errors.Is(err, errors.HTTPUnauthorized) {
				return nil, errors.Unauthorized.Wrap(details)
			}
			if errors.Is(err, errors.HTTPNotFound) {
				return nil, errors.NotFound.Wrap(details)
			}
			return nil, errors.WithStack(details)
		}
		if errors.Is(err, errors.HTTPUnauthorized) {
			return nil, errors.Unauthorized.Wrap(err)
		}
		if errors.Is(err, errors.HTTPNotFound) {
			return nil, errors.NotFound.Wrap(err)
		}
	}
	return response, err
}
