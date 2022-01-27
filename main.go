package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var oldlUrls = map[string]bool{}

func UrlsRead() []string {
	var retUrls []string
	buff, err := ioutil.ReadFile("./urls.txt")
	if err != nil {
		panic(err)
	}
	urls := strings.Split(string(buff), "\n")
	for _, url := range urls {
		if !oldlUrls[url] {
			oldlUrls[url] = true
			retUrls = append(retUrls, url)
		}
	}
	return retUrls
}

func dialTimeout(network, addr string) (net.Conn, error) {
	return net.DialTimeout(network, addr, time.Duration(5*time.Second))
}

func main() {
	t := 10
	transport := http.Transport{
		Dial: dialTimeout,
	}
	client := http.Client{
		Transport: &transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if via[0].URL.String() != req.URL.String() {
				req.Header["Location"] = []string{req.URL.String()}
			}
			return nil
		},
		Timeout: time.Duration(5 * time.Second),
	}
	var wg sync.WaitGroup
	for {
		fmt.Println("=====================================")
		f, _ := os.OpenFile("datalog.log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
		urls := UrlsRead()

		for i, urlSpace := range urls {
			urlTrim := strings.TrimSpace(urlSpace)
			if i != 0 && i%t == 0 {
				wg.Add(t)
				wg.Wait()
			}
			go func(_url string, _i int) {
				defer wg.Done()
				urlParse, _ := url.Parse(_url)
				address, _ := net.LookupIP(urlParse.Host)
				res, err := client.Get(_url)
				str := ""
				if err == nil {
					redirect := ""
					if len(res.Request.Header["Location"]) > 0 {
						redirect = res.Request.Header["Location"][0]
					}
					doc, _ := goquery.NewDocumentFromReader(res.Body)
					title := doc.Find("title").Text()
					str = fmt.Sprintf("%d|%s|%s|%d|%s|%s\n", _i, urlParse.Host, address[0], res.StatusCode, title, redirect)
					fmt.Println(str)
				} else {
					str = fmt.Sprintf("%d|%s|%d\n", _i, urlParse.Host, 500)
					fmt.Println(str)
				}
				f.WriteString(str)
			}(urlTrim, i)
		}
		f.Close()
	}
}
