package box_test

import "github.com/gildas/go-errors"

type BogusData struct {
	ID string `json:"id"`
}

func (data BogusData) MarshalJSON() ([]byte, error) {
	return nil, errors.JSONMarshalError.WithStack()
}
