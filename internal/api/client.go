package api

import (
	"fmt"
	"net/url"
	"sort"
	"strconv"
)

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

const threadFetchSize = 30

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
	if err := c.transport.getJSON(fmt.Sprintf("/t/%d/last.json", id), &t); err != nil {
		return nil, err
	}

	sortPosts(t.PostStream.Posts)

	if len(t.PostStream.Stream) > len(t.PostStream.Posts) {
		need := threadFetchSize - len(t.PostStream.Posts)
		if need < 0 {
			need = 0
		}
		start := len(t.PostStream.Stream) - len(t.PostStream.Posts) - need
		if start < 0 {
			start = 0
		}
		end := len(t.PostStream.Stream) - len(t.PostStream.Posts)
		postIDs := append([]int(nil), t.PostStream.Stream[start:end]...)
		posts, err := c.GetThreadPosts(id, postIDs)
		if err != nil {
			return nil, err
		}
		t.PostStream.Posts = append(posts, t.PostStream.Posts...)
	}

	return &t, nil
}

func (c *Client) GetThreadPosts(id int, postIDs []int) ([]Post, error) {
	if len(postIDs) == 0 {
		return nil, nil
	}

	var r Thread
	values := url.Values{}
	for _, postID := range postIDs {
		values.Add("post_ids[]", strconv.Itoa(postID))
	}
	if err := c.transport.getJSON(fmt.Sprintf("/t/%d/posts.json?%s", id, values.Encode()), &r); err != nil {
		return nil, err
	}
	sortPosts(r.PostStream.Posts)
	return r.PostStream.Posts, nil
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

func sortPosts(posts []Post) {
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].PostNumber < posts[j].PostNumber
	})
}
