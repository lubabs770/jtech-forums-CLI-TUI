package api

type Topic struct {
	ID           int    `json:"id"`
	Title        string `json:"title"`
	Slug         string `json:"slug"`
	PostsCount   int    `json:"posts_count"`
	ReplyCount   int    `json:"reply_count"`
	CategoryID   int    `json:"category_id"`
	LastPostedAt string `json:"last_posted_at"`
	Pinned       bool   `json:"pinned"`
}

type Category struct {
	ID               int    `json:"id"`
	Name             string `json:"name"`
	Slug             string `json:"slug"`
	TopicCount       int    `json:"topic_count"`
	Description      string `json:"description"`
	Color            string `json:"color"`
	TextColor        string `json:"text_color"`
	StyleType        string `json:"style_type"`
	Icon             string `json:"icon"`
	Emoji            string `json:"emoji"`
	ParentCategoryID int    `json:"parent_category_id"`
}

type Post struct {
	ID         int    `json:"id"`
	PostNumber int    `json:"post_number"`
	Username   string `json:"username"`
	Raw        string `json:"raw"`
	Cooked     string `json:"cooked"`
	CreatedAt  string `json:"created_at"`
}

type Thread struct {
	ID         int    `json:"id"`
	Title      string `json:"title"`
	PostStream struct {
		Posts []Post `json:"posts"`
	} `json:"post_stream"`
}

// Discourse feed response wrapper
type topicListResponse struct {
	TopicList struct {
		Topics []Topic `json:"topics"`
	} `json:"topic_list"`
}

type categoryListResponse struct {
	CategoryList struct {
		Categories []Category `json:"categories"`
	} `json:"category_list"`
}
