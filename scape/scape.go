package scape

import (
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

const baseAddr = "https://scape.lasseufpa.org/"

// Scape TODO comments.
type Scape struct {
	client http.Client
}

func New() *Scape {
	scape := new(Scape)
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatal(err)
	}
	scape.client = http.Client{Jar: jar}
	return scape
}

func (scape *Scape) Login(user, pass string) string {
	module := "index.php?module=autentication"
	resp, err := scape.client.PostForm(
		baseAddr+module,
		url.Values{
			"login":  {user},
			"passwd": {pass},
		})
	if err != nil {
		log.Fatal(err)
	}
	return extractName(resp.Body)
}

func findTables(body io.ReadCloser) []*html.Node {
	doc, err := html.Parse(body)
	if err != nil {
		log.Fatal(err)
	}
	tables := make([]*html.Node, 0)
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "table" {
			tables = append(tables, n)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	return tables
}

func readTable(table *html.Node) []string {
	names := make([]string, 0)
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.TextNode {
			data := strings.TrimSpace(n.Data)
			if len(data) != 0 {
				names = append(names, n.Data)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(table)
	return names
}

func extractName(body io.ReadCloser) string {
	doc, err := html.Parse(body)
	if err != nil {
		log.Fatal(err)
	}
	var name string
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "h1" {
			if c := n.FirstChild; c.Type == html.TextNode {
				name = strings.TrimSpace(string([]rune(c.Data)[3:]))
				return
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	return name
}

func extractUsers(body io.ReadCloser) []string {
	tables := findTables(body)
	names := readTable(tables[1])
	names = names[1:] //removes title
	return names
}

func (scape *Scape) WorkingUsers() []string {
	module := "index.php?module=working"
	resp, err := scape.client.Get(baseAddr + module)
	if err != nil {
		log.Fatal(err)
	}
	return extractUsers(resp.Body)
}

func (scape *Scape) IsWorking(name string) bool {
	names := scape.WorkingUsers()
	for _, n := range names {
		if n == name {
			return true
		}
	}
	return false
}
