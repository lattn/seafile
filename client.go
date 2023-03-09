package seafile

import (
	"io"
	"net/http"
)

type Client struct {
	endpoint string
	token    string
}

func New(endpoint, token string) *Client {
	return &Client{endpoint: endpoint, token: token}
}

func (c *Client) makeURL(path string) string {
	return c.endpoint + path
}

func (c *Client) request(req *http.Request) (status int, body []byte, err error) {
	req.Header.Set("Authorization", "Token "+c.token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	status = resp.StatusCode
	body, err = io.ReadAll(resp.Body)
	return
}
