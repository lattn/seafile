package seafile

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Repo struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Permission string `json:"permission"`
	MTime      int    `json:"mtime"`
}

func (c *Client) Libraries(ctx context.Context) (repos []Repo, err error) {
	req, err := http.NewRequest(http.MethodGet, c.makeURL("/api2/repos"), nil)
	if err != nil {
		return
	}
	status, body, err := c.request(req.WithContext(ctx))
	if err != nil {
		return
	}
	if status != http.StatusOK {
		err = fmt.Errorf("upload file: %d", status)
		return
	}
	err = json.Unmarshal(body, &repos)
	return
}

func (c *Client) LibraryByName(ctx context.Context, name string) (*Repo, error) {
	libs, err := c.Libraries(ctx)
	if err != nil {
		return nil, err
	}
	for _, lib := range libs {
		if lib.Name == name {
			lib := lib
			return &lib, nil
		}
	}
	return nil, nil
}
