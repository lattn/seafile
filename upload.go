package seafile

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
)

type UploadLink struct {
	c    *Client
	link string
}

func (link UploadLink) UploadFile(ctx context.Context, filename string, fileReader io.Reader) (res string, err error) {
	r, w := io.Pipe()

	mw := multipart.NewWriter(w)
	go func() {
		defer w.Close()
		defer mw.Close()

		err = mw.WriteField("parent_dir", "/")
		if err != nil {
			return
		}
		err = mw.WriteField("relative_path", strings.TrimPrefix(filepath.Dir(filename), "/"))
		if err != nil {
			return
		}
		part, err := mw.CreateFormFile("file", filepath.Base(filename))
		if err != nil {
			return
		}
		_, err = io.Copy(part, fileReader)
		if err != nil {
			return
		}
	}()

	req, err := http.NewRequest(http.MethodPost, link.link, r)
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())

	status, body, err := link.c.request(req.WithContext(ctx))
	if err != nil {
		return
	}
	if status != http.StatusOK {
		err = fmt.Errorf("upload file: %d", status)
		return
	}
	res = string(body)
	return
}

func (c *Client) CreateUploadLink(ctx context.Context, repoID string) (link UploadLink, err error) {
	req, err := http.NewRequest(http.MethodGet, c.makeURL(fmt.Sprintf("/api2/repos/%s/upload-link/", repoID)), nil)
	if err != nil {
		return
	}
	status, body, err := c.request(req.WithContext(ctx))
	if err != nil {
		return
	}
	if status != http.StatusOK {
		err = fmt.Errorf("get upload link: %d", status)
		return
	}
	link = UploadLink{c, strings.Trim(string(body), `"`)}
	return
}
