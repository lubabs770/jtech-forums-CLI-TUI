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

func (c *Client) csrfToken() (string, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/session/csrf.json")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var result struct {
		Csrf string `json:"csrf"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.Csrf, nil
}

func (c *Client) Login(username, password string) (string, error) {
	csrf, err := c.csrfToken()
	if err != nil {
		return "", fmt.Errorf("failed to get CSRF token: %w", err)
	}

	body, _ := json.Marshal(map[string]string{"login": username, "password": password})
	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/session", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CSRF-Token", csrf)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	resp, err := c.httpClient.Do(req)
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

type ErrUnauthorized struct{}

func (e *ErrUnauthorized) Error() string { return "session expired or invalid" }

func (c *Client) get(path string, out any) error {
	resp, err := c.httpClient.Get(c.baseURL + path)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusUnauthorized {
		return &ErrUnauthorized{}
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d for %s", resp.StatusCode, path)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func (c *Client) GetFeed(feed string) ([]Topic, error) {
	var r topicListResponse
	if err := c.get("/"+feed+".json", &r); err != nil {
		return nil, err
	}
	return r.TopicList.Topics, nil
}

func (c *Client) GetCategories() ([]Category, error) {
	var r categoryListResponse
	if err := c.get("/categories.json", &r); err != nil {
		return nil, err
	}
	return r.CategoryList.Categories, nil
}

func (c *Client) GetCategoryTopics(slug string, id int) ([]Topic, error) {
	var r topicListResponse
	if err := c.get(fmt.Sprintf("/c/%s/%d.json", slug, id), &r); err != nil {
		return nil, err
	}
	return r.TopicList.Topics, nil
}

func (c *Client) GetThread(id int) (*Thread, error) {
	var t Thread
	if err := c.get(fmt.Sprintf("/t/%d.json", id), &t); err != nil {
		return nil, err
	}
	return &t, nil
}

func (c *Client) post(path string, payload, out any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusUnauthorized {
		return &ErrUnauthorized{}
	}
	if resp.StatusCode >= 400 {
		var e struct {
			Errors []string `json:"errors"`
		}
		json.NewDecoder(resp.Body).Decode(&e)
		if len(e.Errors) > 0 {
			return fmt.Errorf("post failed: %s", e.Errors[0])
		}
		return fmt.Errorf("post failed: status %d", resp.StatusCode)
	}
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}

func (c *Client) PostReply(topicID int, raw string) error {
	return c.post("/posts", map[string]any{
		"topic_id": topicID,
		"raw":      raw,
	}, nil)
}

func (c *Client) CreateTopic(title, raw string, categoryID int) error {
	return c.post("/posts", map[string]any{
		"title":    title,
		"raw":      raw,
		"category": categoryID,
	}, nil)
}
