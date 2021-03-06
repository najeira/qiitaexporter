package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
)

var (
	flagPostDir  = flag.String("postdir", "content/posts", "posts dir")
	flagTmplFile = flag.String("template", "", "template file")

	quoteReplacer = strings.NewReplacer("\"", "\\\"")
)

func main() {
	flag.Parse()

	if *flagTmplFile != "" {
		var err error
		tmpl, err = template.ParseFiles(*flagTmplFile)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
	}

	for i := 1; ; i++ {
		hasNext, err := download100(i)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}

		if !hasNext {
			break
		}
	}
}

func download100(page int) (hasNext bool, rerr error) {
	url := fmt.Sprintf("https://qiita.com/api/v2/authenticated_user/items?page=%d&per_page=20", page)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return false, err
	}

	resp, err := do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, errors.New(resp.Status)
	}

	var items []*Item
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return false, err
	}

	if err := os.MkdirAll(*flagPostDir, 0777); err != nil {
		return false, err
	}

	for i := range items {
		item := items[i]

		if item.Private {
			continue
		}

		// 関連画像をまとめるためにディレクトリを作っていく
		slug := fmt.Sprintf("qiita-%s", item.ID)
		dirName := fmt.Sprintf("%s-%s", item.Date(), slug)
		dirPath := filepath.Join(*flagPostDir, dirName)
		if err := os.MkdirAll(dirPath, 0777); err != nil {
			return false, err
		}

		if err := item.ImageToLocal(dirPath, slug); err != nil {
			return false, err
		}

		item.Title = quoteReplacer.Replace(item.Title)

		//fname := fmt.Sprintf("%s-qiita-%s.ja.md", item.Date(), item.ID)
		docPath := filepath.Join(*flagPostDir, dirName, "index.md")

		// start print
		fmt.Print(item.Title, "...")

		f, err := os.Create(docPath)
		if err != nil {
			return false, err
		}

		if err := tmpl.Execute(f, item); err != nil {
			return false, err
		}

		if err := f.Close(); err != nil {
			return false, err
		}

		// end print
		fmt.Println("done")
	}

	total, err := strconv.Atoi(resp.Header.Get("Total-Count"))
	if err != nil {
		return false, err
	}

	return page < total, nil
}

func do(req *http.Request) (*http.Response, error) {
	token := fmt.Sprintf("Bearer %s", os.Getenv("QIITA"))
	req.Header.Set("Authorization", token)
	return http.DefaultClient.Do(req)
}
