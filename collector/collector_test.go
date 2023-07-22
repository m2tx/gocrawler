package collector

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/html"
)

func TestCollector(t *testing.T) {
	collector := New(&HTTPClientMock{
		StatusCode: 200,
		Body:       "<html><head></head><body><p id=\"ID\">TEXT</p><div><div><ok>OK</ok></div></div></body></html>",
	})

	collector.OnRequest(func(req *http.Request) error {
		assert.Equal(t, "mock.com", req.URL.Host)
		return nil
	})

	collector.OnNode("p", func(req *http.Request, resp *http.Response, node *html.Node) error {
		assert.Equal(t, "p", node.Data)
		assert.Equal(t, "TEXT", node.FirstChild.Data)
		return nil
	})

	collector.OnNode("div div ok", func(req *http.Request, resp *http.Response, node *html.Node) error {
		assert.Equal(t, "ok", node.Data)
		assert.Equal(t, "OK", node.FirstChild.Data)
		return nil
	})

	collector.Visit("http://mock.com")
}
