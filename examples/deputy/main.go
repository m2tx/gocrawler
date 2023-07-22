package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/m2tx/gocrawler/collector"
	"github.com/m2tx/gocrawler/selector"
	"github.com/m2tx/gocrawler/worker"
	"golang.org/x/net/html"
)

const (
	bitSize int = 64

	legislatury int = 56
	year        int = 2023
)

var (
	deputyRegex *regexp.Regexp
	realRegex   *regexp.Regexp
)

func init() {
	deputyRegex = regexp.MustCompile(`(?P<Name>[\w\W\s]*) \((?P<PoliticalParty>[\w\W\s]*)-(?P<State>[\w\W\s]*)\)`)
	realRegex = regexp.MustCompile(`R\$\s(?P<VALUE>[0-9.]*,[0-9]{2})`)
}

type CostDetail struct {
	Description string  `json:"description"`
	Value       float64 `json:"value"`
}

type Deputy struct {
	ID                        string       `json:"id"`
	Name                      string       `json:"name"`
	PoliticalParty            string       `json:"politicalParty"`
	State                     string       `json:"state"`
	Salary                    float64      `json:"salary"`
	OfficeBudget              float64      `json:"officeBudget"`
	ParliamentaryQuota        float64      `json:"parliamentaryQuota"`
	ParliamentaryQuotaDetails []CostDetail `json:"parliamentaryQuotaDetails"`
	Total                     float64      `json:"total"`
}

func main() {
	ctx := context.Background()

	workerDeputy := worker.NewWorkerPool[*Deputy](5, setDeputyDetails)
	workerDeputy.Start(ctx)

	getDeputiesCost(ctx, workerDeputy)

	workerDeputy.Wait()
}

func getDeputiesCost(ctx context.Context, worker *worker.WorkerPool[*Deputy]) {
	attrValue := selector.Attribute("value")

	c := collector.NewWithDefault()

	c.OnNode("select#deputado option", func(req *http.Request, resp *http.Response, node *html.Node) error {
		if node.FirstChild.Type == html.TextNode {
			data := node.FirstChild.Data
			if deputyRegex.Match([]byte(data)) {
				strs := deputyRegex.FindStringSubmatch(data)

				deputy := &Deputy{
					ID:             attrValue.Val(node),
					Name:           strs[1],
					PoliticalParty: strs[2],
					State:          strs[3],
				}

				worker.Add(deputy)
			}
		}

		return nil
	})

	err := c.Visit(fmt.Sprintf("https://www.camara.leg.br/transparencia/gastos-parlamentares?legislatura=%d&ano=%d&mes=&por=deputado&deputado=&uf=&partido=", legislatury, year))
	if err != nil {
		fmt.Println(err)
		return
	}
}

func setDeputyDetails(ctx context.Context, deputy *Deputy) {
	c := collector.NewWithDefault()

	c.OnRequest(func(req *http.Request) error {
		fmt.Println(req.URL)

		return nil
	})

	c.OnNode("section#verba div.container div.gastos__resumo p.gastos__resumo-texto--destaque", func(req *http.Request, resp *http.Response, node *html.Node) error {
		data := node.FirstChild.Data

		strs := realRegex.FindStringSubmatch(data)

		officeBudget, err := parseFloat(strs[0])
		if err != nil {
			return err
		}

		deputy.OfficeBudget = officeBudget

		return nil
	})

	c.OnNode("div.remuneracao-viagens div#remuneracao p.remuneracao-viagens__desc", func(req *http.Request, resp *http.Response, node *html.Node) error {
		data := node.FirstChild.Data

		strs := realRegex.FindStringSubmatch(data)

		salary, err := parseFloat(strs[0])
		if err != nil {
			return err
		}

		deputy.Salary = salary

		return nil
	})

	c.OnNode("section#cota table#js-tipo-despesa.js-chart--pie tbody tr", func(req *http.Request, resp *http.Response, node *html.Node) error {
		query := selector.QueryString("td")
		nodes := query.Select(node)

		value, err := parseFloat(nodes[1].FirstChild.Data)
		if err != nil {
			return fmt.Errorf("error.cost.details: %v", err)
		}

		costDetails := CostDetail{
			Description: nodes[0].FirstChild.Data,
			Value:       value,
		}
		deputy.ParliamentaryQuotaDetails = append(deputy.ParliamentaryQuotaDetails, costDetails)

		return nil
	})

	c.OnNode("div.gastos__resumo div.card-body section p.gastos__resumo-texto--destaque span", func(req *http.Request, resp *http.Response, node *html.Node) error {
		parliamentaryQuota, err := parseFloat(node.FirstChild.Data)
		if err != nil {
			return fmt.Errorf("error.cost.total: %v", err)
		}

		deputy.ParliamentaryQuota = parliamentaryQuota

		return nil
	})

	if err := c.Visit(fmt.Sprintf("https://www.camara.leg.br/transparencia/gastos-parlamentares?legislatura=%d&ano=%d&mes=&por=deputado&deputado=%s&uf=&partido=", legislatury, year, deputy.ID)); err != nil {
		return
	}

	deputy.Total = deputy.Salary + deputy.OfficeBudget + deputy.ParliamentaryQuota

	bytes, err := json.MarshalIndent(deputy, "", " ")
	if err != nil {
		fmt.Println(err)
	}

	err = os.WriteFile(fmt.Sprintf("./tmp/deputy_%s.json", deputy.ID), bytes, 0644)
	if err != nil {
		fmt.Println(err)
	}
}

func parseFloat(v string) (value float64, err error) {
	v = strings.Replace(v, "R$", "", 1)
	v = strings.Trim(v, " ")

	containsComman, containsDot := strings.ContainsRune(v, ','), strings.ContainsRune(v, '.')
	if containsComman && containsDot {
		v = strings.ReplaceAll(v, ".", "")
		v = strings.Replace(v, ",", ".", 1)
	} else if containsComman {
		v = strings.Replace(v, ",", ".", 1)
	} else if containsDot {
		v = strings.ReplaceAll(v, ".", "")
	}

	value, err = strconv.ParseFloat(v, bitSize)
	return
}
