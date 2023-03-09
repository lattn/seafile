package seafile

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func GenerateToken(endpoint, username, password string) (token string, err error) {
	f := url.Values{}
	f.Set("username", username)
	f.Set("password", password)
	resp, err := http.PostForm(endpoint+"/api2/auth-token/", f)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("query api: %d", resp.StatusCode)
		return
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	result := struct {
		Token string `json:"token"`
	}{}

	err = json.Unmarshal(b, &result)
	if err != nil {
		return
	}

	token = result.Token
	return
}

func (c *Client) ValidateToken(ctx context.Context) (bool, error) {
	req, err := http.NewRequest(http.MethodGet, c.makeURL("/api2/auth/ping"), nil)
	if err != nil {
		return false, err
	}
	_, body, err := c.request(req.WithContext(ctx))
	if err != nil {
		return false, err
	}
	return string(body) == `"pong"`, nil
}
