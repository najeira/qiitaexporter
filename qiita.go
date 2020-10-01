package main

import (
	"strconv"
	"strings"
	"time"
)

type Tag struct {
	Name     string   `json:"name"`
	Versions []string `json:"versions"`
}

type Item struct {
	ID           string    `json:"id"`
	URL          string    `json:"url"`
	Title        string    `json:"title"`
	Body         string    `json:"body"`
	RenderedBody string    `json:"rendered_body"`
	CreatedAt    time.Time `json:"created_at"`
	Tags         []*Tag    `json:"tags"`
	Private      bool      `json:"private"`
}

func (item *Item) AllTags() string {
	tags := make([]string, len(item.Tags))
	for i := range item.Tags {
		tags[i] = strconv.Quote(item.Tags[i].Name)
	}
	return strings.Join(tags, ",")
}

func (item *Item) Date() string {
	return item.CreatedAt.Format("2006-01-02")
}

func (item *Item) ImageToLocal(dir, slug string) error {
	body, imgs := convertImages(slug, item.Body)

	for _, img := range imgs {
		img.download(dir)
	}

	item.Body = body

	return nil
}
