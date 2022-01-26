package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func UrlsRead() []string {
	buff, err := ioutil.ReadFile("./urls.txt")
	if err != nil {
		panic(err)
	}
	return strings.Split(string(buff), "\n")
}

func main() {
	f, err := os.OpenFile("datalog.log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}

	defer f.Close()
	urls := UrlsRead()
	for _, urlSpace := range urls {
		url := strings.TrimSpace(urlSpace)
		res, _ := http.Get(url)
		urlNoHttp := strings.Split(url, "//")[1]
		conn, _ := net.Dial("tcp", fmt.Sprintf("%s:80", urlNoHttp))
		ip := strings.ReplaceAll(conn.RemoteAddr().String(), ":80", "")
		doc, err := goquery.NewDocumentFromReader(res.Body)
		if err != nil {
			panic(err)
		}
		title := doc.Find("title").Text()
		str := fmt.Sprintf("%s | %s | %d | %s\n", urlNoHttp, ip, res.StatusCode, title)
		f.WriteString(str)
	}
}
