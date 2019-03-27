package scape

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
)

type Clock struct {
	in  string
	out string
}

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

func readNames(table *html.Node) []string {
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
	names := readNames(tables[1])
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

func readHours(table *html.Node) []Clock {
	clocks := make([]Clock, 0)
	for row := table.FirstChild.NextSibling.FirstChild.NextSibling; row != nil; row = row.NextSibling {
		if row.Data != "tr" {
			continue
		}

		idx := 0
		var clockIn, clockOut string
		for col := row.FirstChild; col != nil; col = col.NextSibling {
			child := col.FirstChild
			if child == nil || child.Type != html.TextNode {
				continue
			}
			if idx == 2 {
				clockIn = strings.TrimSpace(child.Data)
			} else if idx == 3 {
				clockOut = strings.TrimSpace(child.Data)
				clocks = append(clocks, Clock{in: clockIn, out: clockOut})
			}
			idx++
		}
	}
	return clocks
}

const timeLayout = "15:04:05"

var eightAm, _ = time.Parse(timeLayout, "08:00:00")
var onePm, _ = time.Parse(timeLayout, "13:00:00")
var twoPm, _ = time.Parse(timeLayout, "14:00:00")
var sixPm, _ = time.Parse(timeLayout, "18:00:00")

func twoDigits(i int) string {
	return fmt.Sprintf("%02d", i)
}

func sumHours(clocks []Clock) time.Duration {
	hours := time.Duration(0)
	for _, clock := range clocks {
		in, _ := time.Parse(timeLayout, clock.in)
		out, err := time.Parse(timeLayout, clock.out)
		if err != nil { // Não fechou o ponto
			h, m, s := time.Now().Clock()
			out, _ = time.Parse(timeLayout, twoDigits(h)+":"+twoDigits(m)+":"+twoDigits(s))
		}
		if in.Before(onePm) {
			if in.Before(eightAm) {
				in = eightAm
			}
			if out.After(onePm) {
				out = onePm
			}
		} else {
			if in.Before(twoPm) {
				in = twoPm
			}
			if out.After(sixPm) {
				out = sixPm
			}
		}
		hours += out.Sub(in)
	}
	return hours
}

func (scape *Scape) HoursToday() time.Duration {
	module := "index.php?module=rel_horas"
	year, month, day := time.Now().Date()

	resp, err := scape.client.PostForm(
		baseAddr+module,
		url.Values{
			"dia": {strconv.Itoa(day)},
			"mes": {strconv.Itoa(int(month))},
			"ano": {strconv.Itoa(year)},
		})
	if err != nil {
		log.Fatal(err)
	}
	tables := findTables(resp.Body)
	// tables[0] is the calendar
	// tables[1] is the date selector
	// tables[2] is the clock in/out time
	// tables[3] is the total time

	clocks := readHours(tables[2])

	return sumHours(clocks)
}
