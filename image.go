package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	imgRegexp = regexp.MustCompile(`https://(qiita-image-store\.s3\.amazonaws\.com/|qiita-user-contents\.imgix\.net/https%3A%2F%2Fqiita-image-store\.s3\.amazonaws\.com).+?["\])>]`)
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

func convertImages(slug, body string) (string, []Image) {
	imgs := make([]Image, 0)
	count := 0
	body = imgRegexp.ReplaceAllStringFunc(body, func(s string) string {
		count++

		//img := parseImageTag(s)
		src := strings.TrimRight(s, " ])>\"")
		cls := s[len(src):]
		img := Image{
			Src: src,
		}

		ext := path.Ext(img.Src)

		// 記事ごとに別のディレクトリなのでカウントだけでよい
		//img.FileName = fmt.Sprintf("qiita-%s-%d%s", p.itemID, count, ext)
		img.FileName = fmt.Sprintf("%d%s", count, ext)

		imgs = append(imgs, img)

		// 記事ごとに別のディレクトリなので相対パス表記する
		//imgPath := path.Join(p.imagePrefix, img.FileName)
		imgPath := fmt.Sprintf("/posts/%s/%s%s", slug, img.FileName, cls)

		//imgTag := fmt.Sprintf("![%s](%s)", img.Alt, imgPath)
		return imgPath
	})
	return body, imgs
}
