package cursor

import (
	"io"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
	"github.com/wearefair/log-aggregator/types"
)

type Client struct {
	file    *os.File
	current types.Cursor
}

type DB interface {
	Cursor() types.Cursor
	Set(types.Cursor) error
}

func New(cursorFilePath string) (*Client, error) {
	file, err := os.OpenFile(cursorFilePath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to open cursor file at %s", cursorFilePath)
	}
	cursor, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to read cursor from file")
	}
	return &Client{
		file:    file,
		current: types.Cursor(cursor),
	}, nil
}

func (c *Client) Cursor() types.Cursor {
	return c.current
}

func (c *Client) Set(cursor types.Cursor) error {
	var err error
	_, err = c.file.Seek(0, io.SeekStart)
	if err != nil {
		return errors.Wrap(err, "Error seeking to beginning of cursor file")
	}

	_, err = c.file.WriteString(string(cursor))
	if err != nil {
		return errors.Wrap(err, "Error writing cursor to cursor file")
	}

	err = c.file.Truncate(int64(len(cursor)))
	if err != nil {
		return errors.Wrap(err, "Error truncating cursor file")
	}

	err = c.file.Sync()
	if err != nil {
		return errors.Wrap(err, "Error syncing cursor file")
	}
	c.current = cursor
	return nil
}
