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

func main() {
	ctx := context.Background()

	workerDownload := worker.NewWorkerPool[string](5, download)
	workerDownload.Start(ctx)

	visit("http://www.globo.com", workerDownload)

	workerDownload.Wait()
}

func download(ctx context.Context, url string) {
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

func getFileNameByURL(url string) string {
	args := strings.Split(url, "/")
	fileName := "./tmp/" + args[len(args)-1]
	return fileName
}

func visit(url string, workerDownload *worker.WorkerPool[string]) {
	attrSrc := selector.Attribute("src")

	c := collector.NewWithDefault()

	c.OnNode("img", func(req *http.Request, resp *http.Response, node *html.Node) error {
		src := attrSrc.Val(node)

		url, err := resp.Request.URL.Parse(src)
		if err != nil {
			return err
		}

		workerDownload.Add(url.String())
		return nil
	})

	err := c.Visit(url)
	if err != nil {
		fmt.Println(err)
		return
	}
}
