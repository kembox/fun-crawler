package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

// The `site_attributes` struct contains different jquery format strings
// To get to needed location in web pages by `chromedp` and `goquery`
type site_attributes struct {
	button_querySelector     string
	like_box_querySelector   string
	like_count_querySelector string
	date_querySelector       string
	extra_wait_milisec       int
}

var sites map[string]site_attributes = map[string]site_attributes{
	"vnexpress.net": {
		button_querySelector:     ".txt_666",
		like_box_querySelector:   ".reactions-total",
		like_count_querySelector: ".number",
		date_querySelector:       ".date",
		extra_wait_milisec:       1,
	},
	"tuoitre.vn": {
		button_querySelector:     ".viewmore-comment",
		like_box_querySelector:   ".totalreact",
		like_count_querySelector: ".total",
		date_querySelector:       ".detail-time",
		extra_wait_milisec:       2000,
	},
}

/*
var result_file = "./vne_result.txt"
var checked_urls_file = "./checked_urls.txt"
*/
var result_file = "./test_vne_result.txt"
var checked_urls_file = "./test_checked_urls.txt"

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {

	f, err := os.OpenFile(result_file, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	fc, err := os.OpenFile(checked_urls_file, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	defer fc.Close()

	checked_urls, cerr := os.ReadFile(checked_urls_file)
	check(cerr)

	myurls := bufio.NewReader(os.Stdin)

	// Read urls from input, check if we had result already to decide to get and score
	// So we can resume from the previous run
	for {
		myurl, err := myurls.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}
		if len(strings.TrimSpace(myurl)) == 0 {
			break
		}
		myurl = strings.TrimSpace(myurl)

		hostname := get_hostname(myurl)

		//Log url to checked list
		fmt.Println(myurl)
		fc.WriteString(myurl)

		if !bytes.Contains(checked_urls, []byte(myurl)) {
			log.Printf("Start checking %s\n", myurl)
			score_result, err := like_collector(myurl, sites[hostname])
			if err != nil {
				fmt.Println(err)
				continue
			}
			for k, v := range score_result {
				f.WriteString(k + ":" + strconv.Itoa(v) + "\n")
			}
			log.Printf("Done: %s:%d", myurl, score_result[myurl])
		}
	}

}

// Open chrome headless to navigate to a url
// Perform a click action by a custom js file if needed
func click_n_get(url, js string, extra_wait_milisec int) string {
	var body string

	//Set browser options
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("blink-settings", "imagesEnabled=false"),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-web-security", true),
	)
	// new browser, first tab
	browserCtx, browserCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer browserCancel()

	// create chrome instance
	ctx, cancel := chromedp.NewContext(
		browserCtx,
		//chromedp.WithDebugf(log.Printf),
	)
	defer cancel()

	// create a timeout
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Start browsing
	err := chromedp.Run(ctx,
		network.Enable(),
		//Filter some unnecessary traffic
		network.SetBlockedURLS([]string{
			"https://*google*",
			"https://www.google*",
			"https://*pubmatic.com*",
			"https://*adnxs.com*",
			"https://*doubleclick.net*",
			"*eclick.vn*",
			"https://*.vnecdn.net/*like.svg",
			"https://vnexpress.net/microservice/*",
			"https://my.vnexpress.net/*",
			"https://*.admicro.vn/*",
			"https://*.yomedia.vn/*",
			"https://sb.scorecardresearch.com/*",
			"https://*sohatv.vn/*",
		}),
		chromedp.Navigate(url),

		//Wait for whole body to be ready
		//The original method to wait for a special block comment only
		//But there are too many edge case so I do it for sure
		chromedp.WaitReady("body", chromedp.ByQuery),

		// click show more comment . Don't know how to speed this up in js part yet
		// Also can't make a simple loop here. Need to check chromedp syntax a bit
		// Look silly but ok
		//chromedp.Evaluate(js, empty_place_holder),
		chromedp.Evaluate(js, nil),
		chromedp.Evaluate(js, nil),
		chromedp.Evaluate(js, nil),
		chromedp.Evaluate(js, nil),
		chromedp.Sleep(time.Millisecond*time.Duration(extra_wait_milisec)),
		chromedp.WaitVisible(`body`, chromedp.ByQuery),

		chromedp.OuterHTML(`body`, &body, chromedp.ByQuery),
	)
	if err != nil {
		log.Fatal(err)
	}
	return body
}

func is_old_url(myurl string, date_jqSelector string) bool {
	tmp_date := "/11/2023"
	resp, err := http.Get(myurl)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()
	cdoc, _ := goquery.NewDocumentFromReader(resp.Body)
	return !strings.Contains(cdoc.Find(date_jqSelector).Text(), tmp_date)
}

// Navigate to an url by chrome headless
// Check if the date is relevant
// Click "more comments" button to show enough data then fetch it
// Parse html to get total number of likes
// The `site_attributes` struct contains different jquery format string to get to needed location
func like_collector(myurl string, s site_attributes) (map[string]int, error) {
	var result = make(map[string]int)
	button_querySelector := s.button_querySelector
	like_box_querySelector := s.like_box_querySelector
	like_count_querySelector := s.like_count_querySelector
	extra_wait_milisec := s.extra_wait_milisec
	date_querySelector := s.date_querySelector

	if is_old_url(myurl, date_querySelector) {
		return result, errors.New("skipped old page")
	}

	button_in_js := fmt.Sprintf("if (document.querySelector('%s')) { document.querySelector('%s').click();}", button_querySelector, button_querySelector)
	body := click_n_get(myurl, button_in_js, extra_wait_milisec)
	//For some reasons I can't load full tuoitre's content with chromedp properly. So need to put in some extra sleep

	result[myurl] = count_likes(body, like_box_querySelector, like_count_querySelector)

	return result, nil
}

func count_likes(body_html, like_box_querySelector, like_count_querySelector string) (total_likes int) {

	total_likes = 0

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body_html))
	if err != nil {
		log.Fatal(err)
	}

	//select 2 class to make sure it's the correct place to check
	doc.Find(like_box_querySelector).Each(func(i int, s *goquery.Selection) {
		// For each item found, get the number
		number := s.Find(like_count_querySelector).Text()
		if number != "" {
			num, err := strconv.Atoi(strings.ReplaceAll(number, ".", ""))
			if err != nil {
				log.Fatal(err)
			}
			total_likes += num
		}
	})

	return total_likes
}

func get_hostname(myurl string) (domain string) {
	xurl, err := url.Parse(myurl)
	if err != nil {
		log.Fatal(err)
	}
	hostname := strings.TrimPrefix(xurl.Hostname(), "www.")
	return hostname
}
