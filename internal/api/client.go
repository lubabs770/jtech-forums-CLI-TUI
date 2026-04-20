package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
	jar        *cookiejar.Jar
}

func New(baseURL, sessionCookie string) (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	c := &Client{
		baseURL:    baseURL,
		jar:        jar,
		httpClient: &http.Client{Jar: jar},
	}
	if sessionCookie != "" {
		u, err := url.Parse(baseURL)
		if err != nil {
			return nil, err
		}
		jar.SetCookies(u, []*http.Cookie{{Name: "_t", Value: sessionCookie}})
	}
	return c, nil
}

func (c *Client) Login(username, password string) (string, error) {
	body, _ := json.Marshal(map[string]string{"login": username, "password": password})
	resp, err := c.httpClient.Post(c.baseURL+"/session", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		var e struct{ Error string `json:"error"` }
		json.NewDecoder(resp.Body).Decode(&e)
		if e.Error != "" {
			return "", fmt.Errorf("login failed: %s", e.Error)
		}
		return "", fmt.Errorf("login failed: status %d", resp.StatusCode)
	}
	for _, ck := range resp.Cookies() {
		if ck.Name == "_t" {
			return ck.Value, nil
		}
	}
	return "", fmt.Errorf("no session cookie in response")
}

func (c *Client) SessionCookie() string {
	u, _ := url.Parse(c.baseURL)
	for _, ck := range c.jar.Cookies(u) {
		if ck.Name == "_t" {
			return ck.Value
		}
	}
	return ""
}

// GetFeed stub — real implementation added in Task 5
func (c *Client) GetFeed(feed string) ([]Topic, error) {
	return nil, nil
}
