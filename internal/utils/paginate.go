package utils

import (
	"encoding/base64"

	"github.com/powertoolsdev/api/internal/models"
	paginator "github.com/raphaelvigee/go-paginate"
	"github.com/raphaelvigee/go-paginate/cursor"
	gDriver "github.com/raphaelvigee/go-paginate/driver/gorm"
)

// Paginator is a reference to wrap the third party type
type Paginator = paginator.Paginator

// Cursor is a reference to wrap the third party type
type Cursor = cursor.Cursor

// Type is a reference to wrap the third party type
type Type = cursor.Type

type Page = paginator.Page

// NewPaginator wraps a thirdparty library for pagination
func newPaginator(desc bool) *Paginator {
	// Define the pagination criteria

	driverOpts := gDriver.Options{
		Columns: []gDriver.Column{
			{
				Name: "created_at",
				Desc: desc,
			},
		},
	}

	pg := paginator.New(paginator.Options{
		Driver: gDriver.New(driverOpts),
	})

	return pg
}

type cursorParams struct {
	cursor     string
	cursorType Type
	limit      int
}

func NewPaginator(opts *models.ConnectionOptions) (*Paginator, Cursor, error) {
	var (
		c            Cursor
		pg           *Paginator
		err          error
		curParams    cursorParams
		defaultLimit = 25
	)

	curParams.limit = defaultLimit
	var decoded string

	if opts.After != nil {
		decoded = *opts.After
		curParams.cursorType = cursor.After
	}

	if opts.Before != nil {
		decoded = *opts.Before
		curParams.cursorType = cursor.Before
	}
	if opts.Limit != nil {
		curParams.limit = *opts.Limit
	}
	reverse := false
	if opts.Reverse != nil {
		reverse = *opts.Reverse
	}

	curParams.cursor = decoded
	pg = newPaginator(reverse)
	c, err = pg.Cursor(curParams.cursor, curParams.cursorType, curParams.limit)
	return pg, c, err
}

func EncodeCursor(decoded interface{}) (string, error) {
	cur := cursor.Chain(cursor.MsgPack(), cursor.Base64(base64.StdEncoding))
	val := []interface{}{decoded}
	byt, err := cur.Marshal(val)
	if err != nil {
		return "", err
	}
	return string(byt), nil
}
