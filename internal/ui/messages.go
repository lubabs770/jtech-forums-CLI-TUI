package ui

import "github.com/sam/jtech-tui/internal/api"

type loggedInMsg struct{ cookie string }
type popViewMsg struct{}
type unauthorizedMsg struct{}
type openTopicMsg struct {
	topic     api.Topic
	category  *api.Category
	parent    *api.Category
	feedIndex int
}
type openCategoryMsg struct {
	cat    api.Category
	parent *api.Category
}
type editorFinishedMsg struct {
	content string
	err     error
}
type catsForFormMsg struct {
	cats []api.Category
	err  error
}
type newTopicErrMsg struct{ err error }
type replyErrMsg struct{ err error }
