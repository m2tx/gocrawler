package main

import (
	"bytes"
	"context"
	"fmt"
	"image/jpeg"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/chai2010/webp"
	"github.com/m2tx/gocrawler/collector"
	"github.com/m2tx/gocrawler/selector"
	"github.com/m2tx/gocrawler/worker"
	"golang.org/x/net/html"
)

var (
	workerDownload *worker.WorkerPool[string]
	workerVisit    *worker.WorkerPool[string]

	visited    = map[string]bool{}
	downloaded = map[string]bool{}

	attrSrc  = selector.Attribute("src")
	attrHref = selector.Attribute("href")
)

func main() {
	ctx := context.Background()

	workerDownload = worker.NewWorkerPool[string](5, download)
	workerDownload.Start(ctx)

	workerVisit = worker.NewWorkerPool[string](10, visit)
	workerVisit.Start(ctx)

	workerVisit.Add("https://www.istockphoto.com")

	workerVisit.Wait()
	workerDownload.Wait()
}

func download(ctx context.Context, url string) {
	fmt.Printf("download %s \n", url)
	if strings.HasSuffix(url, ".html") {
		return
	}

	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	fileName := getFileNameByURL(url)
	if fileName == "" {
		return
	}

	out, err := os.Create(fileName)
	if err != nil {
		return
	}
	defer out.Close()

	var reader io.Reader
	if strings.HasSuffix(url, ".webp") {
		img, err := webp.Decode(resp.Body)
		if err != nil {
			return
		}
		var buff bytes.Buffer
		err = jpeg.Encode(&buff, img, &jpeg.Options{Quality: 100})
		if err != nil {
			return
		}
		reader = &buff
	} else {
		reader = resp.Body
	}
	_, err = io.Copy(out, reader)
	if err != nil {
		return
	}
}

func sanitizedURL(url string) string {
	index := strings.IndexRune(url, '?')
	if index > -1 {
		return url[:index]
	}
	return url
}

func getFileNameByURL(url string) string {
	url = sanitizedURL(url)
	args := strings.Split(url, "/")
	fileName := "./tmp/" + args[len(args)-1]
	return fileName
}

func visit(ctx context.Context, url string) {
	fmt.Printf("visit %s \n", url)
	c := collector.NewWithDefault()

	c.OnNode("img", func(req *http.Request, resp *http.Response, node *html.Node) error {
		src := attrSrc.Val(node)
		if src == "" {
			return nil
		}

		url, err := resp.Request.URL.Parse(src)
		if err != nil {
			return err
		}

		addDownload(url.String())
		return nil
	})

	c.OnNode("a", func(req *http.Request, resp *http.Response, node *html.Node) error {
		href := attrHref.Val(node)
		if href == "" {
			return nil
		}

		url, err := resp.Request.URL.Parse(href)
		if err != nil {
			return err
		}

		addVisit(url.String())
		return nil
	})

	err := c.Visit(url)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func addDownload(url string) {
	if downloaded[url] {
		return
	}

	downloaded[url] = true
	fmt.Printf("add download %s \n", url)
	workerDownload.Add(url)
}

func addVisit(url string) {
	if visited[url] {
		return
	}

	visited[url] = true
	fmt.Printf("add visit %s \n", url)
	workerVisit.Add(url)
}
