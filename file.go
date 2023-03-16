package seafile

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type FileDetail struct {
	ID           string `json:"id"`
	Type         string `json:"type"`
	Name         string `json:"name"`
	Permission   string `json:"permission"`
	MTime        int    `json:"mtime"`
	LastModified int    `json:"last_modified"`
	Size         int    `json:"size"`
	CanEdit      bool   `json:"can_edit"`
}

func (c *Client) FileDetail(ctx context.Context, repoID, path string) (*FileDetail, error) {
	req, err := http.NewRequest(http.MethodGet, c.makeURL(fmt.Sprintf("/api2/repos/%s/file/detail/", repoID), "p", path), nil)
	if err != nil {
		return nil, err
	}
	status, b, err := c.request(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	if status == http.StatusNotFound {
		return nil, nil
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("get file detail: %d", status)
	}

	fd := &FileDetail{}
	err = json.Unmarshal(b, fd)

	return fd, nil
}
