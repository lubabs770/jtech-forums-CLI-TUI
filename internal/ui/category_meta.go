package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/sam/jtech-tui/internal/api"
)

func categoryMap(categories []api.Category) map[int]api.Category {
	out := make(map[int]api.Category, len(categories))
	for _, cat := range categories {
		out[cat.ID] = cat
	}
	return out
}

func categoryBadge(cat api.Category) string {
	label := strings.TrimSpace(cat.Name)
	if label == "" {
		return ""
	}

	style := lipgloss.NewStyle().Bold(true).Padding(0, 1)
	bg := normalizeHexColor(cat.Color)
	fg := normalizeHexColor(cat.TextColor)
	if bg != "" {
		style = style.Background(lipgloss.Color(bg))
	}
	if fg == "" && bg != "" {
		fg = contrastTextColor(bg)
	}
	if fg == bg {
		fg = contrastTextColor(bg)
	}
	if fg != "" {
		style = style.Foreground(lipgloss.Color(fg))
	} else {
		style = style.Foreground(lipgloss.Color("69"))
	}
	return style.Render(label)
}

func normalizeHexColor(hex string) string {
	hex = strings.TrimSpace(strings.TrimPrefix(hex, "#"))
	if len(hex) != 6 {
		return ""
	}
	for _, r := range hex {
		if !strings.ContainsRune("0123456789abcdefABCDEF", r) {
			return ""
		}
	}
	return "#" + strings.ToUpper(hex)
}

func categoryMetaLine(cat api.Category, details ...string) string {
	badge := categoryBadge(cat)
	parts := make([]string, 0, len(details)+1)
	if badge != "" {
		parts = append(parts, badge)
	}
	for _, detail := range details {
		detail = strings.TrimSpace(detail)
		if detail != "" {
			parts = append(parts, metaStyle.Render(detail))
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Left, joinWithSeparator(parts)...)
}

func joinWithSeparator(parts []string) []string {
	if len(parts) < 2 {
		return parts
	}
	out := make([]string, 0, len(parts)*2-1)
	for i, part := range parts {
		if i > 0 {
			out = append(out, sepStyle.Render(" • "))
		}
		out = append(out, part)
	}
	return out
}

func categoryDescription(cat api.Category) string {
	parts := []string{metaStyle.Render(fmt.Sprintf("%d topics", cat.TopicCount))}
	if cat.Description != "" {
		parts = append(parts, metaStyle.Render(stripHTML(cat.Description)))
	}
	return lipgloss.JoinHorizontal(lipgloss.Left, joinWithSeparator(parts)...)
}

func categorySectionHeader(cat api.Category) string {
	return categorySectionHeaderWithParent(nil, cat)
}

func categorySectionHeaderWithParent(parent *api.Category, cat api.Category) string {
	return categoryContextBar(parent, cat)
}

func categoryContextBar(parent *api.Category, cat api.Category, details ...string) string {
	parts := []string{}
	if parent != nil {
		parts = append(parts, categoryBadge(*parent), categoryBadge(cat))
	} else {
		left := categoryBadge(cat)
		if left == "" {
			left = titleChipStyle.Render(cat.Name)
		}
		parts = append(parts, left)
	}
	for _, detail := range details {
		detail = strings.TrimSpace(detail)
		if detail != "" {
			parts = append(parts, metaChipStyle.Render(detail))
		}
	}
	return renderContextSegments(parts)
}

func topicContextBar(parent *api.Category, cat api.Category, topicTitle string, timeLabel string) string {
	parts := []string{}
	if parent != nil {
		parts = append(parts, categoryBadge(*parent))
	}
	parts = append(parts,
		categoryBadge(cat),
		titleChipStyle.Render(strings.TrimSpace(topicTitle)),
		metaChipStyle.Render(strings.TrimSpace(timeLabel)),
	)
	return renderContextSegments(parts)
}

func topicHeaderBar(topicTitle string, timeLabel string) string {
	return renderContextSegments([]string{
		titleChipStyle.Render(strings.TrimSpace(topicTitle)),
		metaChipStyle.Render(strings.TrimSpace(timeLabel)),
	})
}

func renderContextSegments(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	joined := make([]string, 0, len(parts)*2-1)
	for i, part := range parts {
		if i > 0 {
			joined = append(joined, contextSepStyle.Render(">"))
		}
		joined = append(joined, part)
	}
	return contextBarStyle.Render(lipgloss.JoinHorizontal(lipgloss.Left, joined...))
}

func contrastTextColor(bg string) string {
	if len(bg) != 7 || bg[0] != '#' {
		return "#FFFFFF"
	}

	r, errR := strconv.ParseInt(bg[1:3], 16, 64)
	g, errG := strconv.ParseInt(bg[3:5], 16, 64)
	b, errB := strconv.ParseInt(bg[5:7], 16, 64)
	if errR != nil || errG != nil || errB != nil {
		return "#FFFFFF"
	}

	brightness := (r*299 + g*587 + b*114) / 1000
	if brightness >= 160 {
		return "#000000"
	}
	return "#FFFFFF"
}
