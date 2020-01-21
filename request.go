package box

import (
	"context"

	"github.com/gildas/go-errors"
	"github.com/gildas/go-request"
)

// sendRequest sends an HTTP request to Box.com's API
func (client *Client) sendRequest(ctx context.Context, options *request.Options, results interface{}) (*request.ContentReader, error) {
	if options == nil {
		options = &request.Options{}
	}
	options.Context   = ctx
	options.Logger    = client.Logger
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
			if errors.Is(err, errors.HTTPBadRequestError) && errors.Is(details, InvalidGrantError) {
				return nil, errors.UnauthorizedError.Wrap(details)
			}
			if errors.Is(err, errors.HTTPUnauthorizedError) {
				return nil, errors.UnauthorizedError.Wrap(details)
			}
			if errors.Is(err, errors.HTTPNotFoundError) {
				return nil, errors.NotFoundError.Wrap(details)
			}
			return nil, errors.WithStack(details)
		}
		if errors.Is(err, errors.HTTPUnauthorizedError) {
			return nil, errors.UnauthorizedError.Wrap(err)
		}
		if errors.Is(err, errors.HTTPNotFoundError) {
			return nil, errors.NotFoundError.Wrap(err)
		}
		return nil, err
	}

	return response, err
}
