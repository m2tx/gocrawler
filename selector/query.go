package selector

import (
	"strings"

	"github.com/m2tx/gocrawler/utils"
	"golang.org/x/net/html"
)

type QueryString string
type Attribute string

type selectorNode struct {
	Data  string
	ID    string
	Class []string
}

func (a Attribute) Equals(value string) bool {
	return strings.EqualFold(string(a), value)
}

func (a Attribute) Val(n *html.Node) string {
	for _, attr := range n.Attr {
		if a.Equals(attr.Key) {
			return attr.Val
		}
	}

	return ""
}

func (query *QueryString) Select(n *html.Node) []*html.Node {
	selectedNodes := []*html.Node{}
	selectors := query.parseQueryToSelectors()

	query.forEachNode(n, func(n *html.Node) {
		if isNode(selectors, n) {
			selectedNodes = append(selectedNodes, n)
		}
	}, nil)

	return selectedNodes
}

func (query *QueryString) forEachNode(n *html.Node, pre, post func(*html.Node)) {
	if pre != nil {
		pre(n)
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		query.forEachNode(c, pre, post)
	}

	if post != nil {
		post(n)
	}
}

func (query *QueryString) parseQueryToSelectors() [][]selectorNode {
	var selectors [][]selectorNode
	orQuery := strings.Split(string(*query), ",")

	for j := 0; j < len(orQuery); j++ {
		var actualSelectors []selectorNode
		strQueriesNode := strings.Split(string(orQuery[j]), " ")
		initial := len(strQueriesNode) - 1

		for i := initial; i >= 0; i-- {
			actualSelectors = append(actualSelectors, splitSelector(strQueriesNode[i]))
		}

		selectors = append(selectors, actualSelectors)
	}

	return selectors
}

func splitSelector(rule string) selectorNode {
	if strings.Contains(rule, "#") && strings.Contains(rule, ".") {
		ar := strings.Split(rule, "#")
		cl := strings.Split(ar[1], ".")
		return selectorNode{
			Data:  ar[0],
			ID:    cl[0],
			Class: cl[1:],
		}
	} else if strings.Contains(rule, "#") {
		ar := strings.Split(rule, "#")
		return selectorNode{
			Data: ar[0],
			ID:   ar[1],
		}
	} else if strings.Contains(rule, ".") {
		ar := strings.Split(rule, ".")
		return selectorNode{
			Data:  ar[0],
			Class: ar[1:],
		}
	}
	return selectorNode{
		Data: rule,
	}
}

func isNode(selectors [][]selectorNode, node *html.Node) bool {
	is := true
ACTUAL:
	for _, actualSelectors := range selectors {
		n := node
		is = true
	QUERY:
		for index, query := range actualSelectors {
			if ok := query.isNode(n); !ok {
				if index == 0 {
					is = false
					continue ACTUAL
				}
				for {
					n = n.Parent
					if n == nil {
						is = false
						continue ACTUAL
					}
					if ok := query.isNode(n); ok {
						continue QUERY
					}
				}
			}
			n = n.Parent
		}
		if is {
			break
		}
	}
	return is
}

func (query *selectorNode) isNode(n *html.Node) bool {
	if n == nil {
		return false
	}

	if n.Type != html.ElementNode {
		return false
	}

	if n.Data == "" {
		return false
	}

	if n.Data != query.Data && query.Data != "" {
		return false
	}

	isQueryClass := len(query.Class) > 0
	isQueryID := query.ID != ""

	if isQueryClass || isQueryID {
		if len(n.Attr) == 0 {
			return false
		}

		checkedClassAttr := false
		checkedIDAttr := false

		for _, a := range n.Attr {
			switch a.Key {
			case "class":
				if isQueryClass {
					if !utils.SliceContainsSlice(strings.Split(a.Val, " "), query.Class) {
						return false
					}
					checkedClassAttr = true
				}
			case "id":
				if isQueryID {
					if query.ID != a.Val {
						return false
					}
					checkedIDAttr = true
				}
			default:
			}
		}

		if isQueryClass && !checkedClassAttr {
			return false
		}

		if isQueryID && !checkedIDAttr {
			return false
		}
	}

	return true
}
