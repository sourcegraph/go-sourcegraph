package sourcegraph

import (
	"fmt"
	"text/template"
)

type RepoGoodiesService interface {
	// ListBadges lists the available badges for repo.
	ListBadges(repo RepoSpec) ([]*Badge, error)

	// ListCounters lists the available counters for repo.
	ListCounters(repo RepoSpec) ([]*Counter, error)
}

type Badge struct {
	Name              string
	Description       string
	ImageURL          string
	UncountedImageURL string
	Markdown          string
}

func (b *Badge) HTML() string {
	return fmt.Sprintf(`<img src="%s" alt="%s">`, template.HTMLEscapeString(b.ImageURL), template.HTMLEscapeString(b.Name))
}

type Counter struct {
	Name              string
	Description       string
	ImageURL          string
	UncountedImageURL string
	Markdown          string
}

func (c *Counter) HTML() string {
	return fmt.Sprintf(`<img src="%s" alt="%s">`, template.HTMLEscapeString(c.ImageURL), template.HTMLEscapeString(c.Name))
}
