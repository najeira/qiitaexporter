package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var (
	// <img width="156" alt="スクリーンショット 2019-03-15 12.04.43.png" src="https://qiita-image-store.s3.amazonaws.com/0/21341/5dbb4305-716f-f49b-0870-b560027c5c28.png">
	imgRegexp = regexp.MustCompile("<img .+?>")
)

type Image struct {
	Src    string
	Width  int
	Height int
	Alt    string

	FileName string
}

func (img *Image) download(dir string) {
	f, err := os.Create(filepath.Join(dir, img.FileName))
	if err != nil {
		panic(err)
	}

	resp, err := http.Get(img.Src)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		panic(err)
	}

	if err := f.Close(); err != nil {
		panic(err)
	}
}

type convertImagesParam struct {
	itemID string
	imagePrefix string
	body string
}

func convertImages(p convertImagesParam) (string, []Image) {
	imgs := make([]Image, 0)
	count := 0
	body := imgRegexp.ReplaceAllStringFunc(p.body, func(s string) string {
		count++

		img := parseImageTag(s)

		ext := path.Ext(img.Src)
		img.FileName = fmt.Sprintf("qiita-%s-%d%s", p.itemID, count, ext)

		imgs = append(imgs, img)

		imgPath := path.Join(p.imagePrefix, img.FileName)
		imgTag := fmt.Sprintf("![%s](%s)", img.Alt, imgPath)
		return imgTag
	})
	return body, imgs
}

func parseImageTag(s string) (ret Image) {
	r := strings.NewReader(s)

	var started bool
	for {
		ch, _, err := r.ReadRune()
		if err != nil {
			return
		} else if ch == '>' {
			return
		}

		// 最初のスペースまではタグ名
		if !started {
			if ch == ' ' {
				started = true
			}
			continue
		}

		name := drainAttributeName(r)

		var value string
		for {
			ch, _, err := r.ReadRune()
			if err != nil {
				panic(err)
			}
			if ch == ' ' {
				continue
			} else if ch == '"' {
				// "で値が始まったの
				value = drainAttributeValue(r, '"')
				break
			} else {
				// "なしに値が始まった
				if err := r.UnreadRune(); err != nil {
					panic(err)
				}
				value = drainAttributeValue(r, ' ')
				break
			}
		}

		name = strings.ToLower(name)
		switch name {
		case "src":
			ret.Src = value
		case "alt":
			ret.Alt = value
		case "width":
			ret.Width, _ = strconv.Atoi(value)
		case "height":
			ret.Height, _ = strconv.Atoi(value)
		}
	}
}

func drainAttributeName(r *strings.Reader) string {
	var value strings.Builder
	for {
		ch, _, err := r.ReadRune()
		if err != nil {
			return value.String()
		} else if ch == ' ' {
			continue
		} else if ch == '=' {
			return value.String()
		}
		value.WriteRune(ch)
	}
}

func drainAttributeValue(r *strings.Reader, end rune) string {
	var value strings.Builder
	for {
		ch, _, err := r.ReadRune()
		if err != nil {
			return value.String()
		} else if ch == end || ch == '>' {
			return value.String()
		}
		value.WriteRune(ch)
	}
}

type imageNameParam struct {
	itemID string
	dir string
	count int
	name string
}

func imageName(p imageNameParam) (string, error) {
	ext := path.Ext(p.name)
	fname := fmt.Sprintf("qiita-%s-%d%s", p.itemID, p.count, ext)
	return path.Join(*flagImgPathPrefix, fname), nil
}
