package selector

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/html"
)

func TestAttribute_Equals(t *testing.T) {
	attrValue := Attribute("value")
	assert.True(t, attrValue.Equals("value"))
}

func TestAttribute_Val(t *testing.T) {
	node, err := html.Parse(strings.NewReader("<p class=\"attribute\">Some Text</p>"))
	assert.NoError(t, err)

	queryString := QueryString("p")
	nodes := queryString.Select(node)

	attrValue := Attribute("class")
	assert.Equal(t, "attribute", attrValue.Val(nodes[0]))
}

func TestQueryString_Select(t *testing.T) {
	testcase := []struct {
		Html           string
		QueryString    QueryString
		ExpectedLength int
	}{
		{
			Html:           "<p><span>Text Span</span></p>",
			QueryString:    "p span",
			ExpectedLength: 1,
		},
		{
			Html:           "<p><span>Text Span</span></p>",
			QueryString:    "p",
			ExpectedLength: 1,
		},
		{
			Html:           "<p><span>Text Span 1</span><span>Text Span 2</span></p>",
			QueryString:    "p span",
			ExpectedLength: 2,
		},
		{
			Html:           "<p><span>Text Span 1</span></p><p><span>Text Span 2</span></p>",
			QueryString:    "p span",
			ExpectedLength: 2,
		},
		{
			Html:           "<p><span id=\"id\">Text Span</span></p>",
			QueryString:    "span#id",
			ExpectedLength: 1,
		},
		{
			Html:           "<p><span class=\"class\">Text Span</span></p>",
			QueryString:    "span.class",
			ExpectedLength: 1,
		},
		{
			Html:           "<p><span class=\"class-separator\">Text Span</span></p>",
			QueryString:    "span.class-separator",
			ExpectedLength: 1,
		},
		{
			Html:           "<p><span class=\"class class-separator\">Text Span</span></p>",
			QueryString:    "span.class.class-separator",
			ExpectedLength: 1,
		},
		{
			Html:           "<p><span><span class=\"class class-separator\">Text Span</span></span></p>",
			QueryString:    "span span.class.class-separator",
			ExpectedLength: 1,
		},
		{
			Html:           "<p><span>Text Span</span></p>",
			QueryString:    "p, span",
			ExpectedLength: 2,
		},
		{
			Html:           "<p><span>Text Span 1</span></p><p><span>Text Span 2</span></p>",
			QueryString:    "p, span",
			ExpectedLength: 4,
		},
		{
			Html:           "<div><div id=\"container\"><div><div><div class=\"container\"></div></div></div></div></div>",
			QueryString:    "div#container div.container",
			ExpectedLength: 1,
		},
	}
	for _, test := range testcase {
		t.Run(string(test.QueryString), func(t *testing.T) {
			node, err := html.Parse(strings.NewReader(test.Html))
			assert.NoError(t, err)

			nodes := test.QueryString.Select(node)
			assert.Len(t, nodes, test.ExpectedLength)
		})
	}
}
