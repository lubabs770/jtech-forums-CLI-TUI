package api

import "fmt"

type Client struct {
	transport *transport
}

func New(baseURL, sessionCookie string) (*Client, error) {
	transport, err := newTransport(baseURL, sessionCookie)
	if err != nil {
		return nil, err
	}
	return &Client{transport: transport}, nil
}

func (c *Client) Login(username, password string) (string, error) {
	return c.transport.login(username, password)
}

func (c *Client) SessionCookie() string {
	return c.transport.sessionCookie()
}

type ErrUnauthorized struct{}

func (e *ErrUnauthorized) Error() string { return "session expired or invalid" }

func (c *Client) GetFeed(feed string) ([]Topic, error) {
	var r topicListResponse
	if err := c.transport.getJSON("/"+feed+".json", &r); err != nil {
		return nil, err
	}
	return r.TopicList.Topics, nil
}

func (c *Client) GetCategories() ([]Category, error) {
	var r categoryListResponse
	if err := c.transport.getJSON("/categories.json", &r); err != nil {
		return nil, err
	}
	return r.CategoryList.Categories, nil
}

func (c *Client) GetCategoryTopics(slug string, id int) ([]Topic, error) {
	var r topicListResponse
	if err := c.transport.getJSON(fmt.Sprintf("/c/%s/%d.json", slug, id), &r); err != nil {
		return nil, err
	}
	return r.TopicList.Topics, nil
}

func (c *Client) GetThread(id int) (*Thread, error) {
	var t Thread
	if err := c.transport.getJSON(fmt.Sprintf("/t/%d.json", id), &t); err != nil {
		return nil, err
	}
	return &t, nil
}

func (c *Client) PostReply(topicID int, raw string) error {
	return c.transport.postJSON("/posts", map[string]any{
		"topic_id": topicID,
		"raw":      raw,
	}, nil)
}

func (c *Client) CreateTopic(title, raw string, categoryID int) error {
	return c.transport.postJSON("/posts", map[string]any{
		"title":    title,
		"raw":      raw,
		"category": categoryID,
	}, nil)
}
