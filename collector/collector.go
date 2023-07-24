package collector

import (
	"fmt"
	"net/http"

	"github.com/m2tx/gocrawler/selector"
	"golang.org/x/net/html"
)

type OnRequest func(req *http.Request) error
type OnNode func(req *http.Request, resp *http.Response, node *html.Node) error

type onNodeEntry struct {
	Query  selector.QueryString
	OnNode OnNode
}

type Collector interface {
	Visit(url string) error
	OnRequest(onRequest OnRequest)
	OnNode(query selector.QueryString, onNode OnNode)
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type collector struct {
	client             HTTPClient
	onRequestListeners []OnRequest
	onNodeListeners    []onNodeEntry
}

func NewWithDefault() Collector {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil
		},
	}

	return New(client)
}

func New(client HTTPClient) Collector {
	return &collector{
		client: client,
	}
}

func (c *collector) Visit(url string) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	for _, onRequest := range c.onRequestListeners {
		if err := onRequest(req); err != nil {
			return err
		}
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode >= http.StatusBadRequest && resp.StatusCode < http.StatusInternalServerError {
		return fmt.Errorf("status code %d for %s", resp.StatusCode, url)
	}

	doc, err := html.Parse(resp.Body)
	resp.Body.Close()
	if err != nil {
		return err
	}
	for _, onNode := range c.onNodeListeners {
		nodes := onNode.Query.Select(doc)
		for _, node := range nodes {
			if err := onNode.OnNode(req, resp, node); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *collector) OnRequest(onRequest OnRequest) {
	c.onRequestListeners = append(c.onRequestListeners, onRequest)
}

func (c *collector) OnNode(query selector.QueryString, onNode OnNode) {
	c.onNodeListeners = append(c.onNodeListeners, onNodeEntry{
		Query:  query,
		OnNode: onNode,
	})
}
