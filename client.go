package seafile

import (
	"io"
	"net/http"
	"net/url"
)

type Client struct {
	endpoint string
	token    string
}

func New(endpoint, token string) *Client {
	return &Client{endpoint: endpoint, token: token}
}

func (c *Client) makeURL(path string, pairs ...string) string {
	if len(pairs) == 0 {
		return c.endpoint + path
	}
	params := make(url.Values)
	for i := 0; i < len(pairs); i += 2 {
		params.Set(pairs[i], pairs[i+1])
	}
	return c.endpoint + path + "?" + params.Encode()
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
