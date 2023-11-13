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
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

var result_file = "./vne_result.txt"
var checked_urls_file = "./checked_urls.txt"

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

	urls := bufio.NewReader(os.Stdin)

	// Read urls from input, check if we had result already to decide to get and score
	// So we can resume from the previous run
	for {
		url, err := urls.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}
		if len(strings.TrimSpace(url)) == 0 {
			break
		}
		url = strings.TrimSpace(url)

		//Log url to checked list
		fc.WriteString(url)

		if !bytes.Contains(checked_urls, []byte(url)) {
			log.Printf("Start checking %s\n", url)
			score_result, err := rank_vnexpress(url)
			if err != nil {
				fmt.Println(err)
				continue
			}
			for k, v := range score_result {
				f.WriteString(k + ":" + strconv.Itoa(v) + "\n")
			}
			log.Printf("Done: %s:%d", url, score_result[url])
		}
	}

}

// Open chrome headless to navigate to a url
// Perform a click action by a custom js file if needed
func click_n_get(url, js string) string {
	var comment string
	var empty_place_holder interface{}

	//Set browser options
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("blink-settings", "imagesEnabled=false"),
		chromedp.Flag("disable-gpu", true),
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
		}),
		chromedp.Navigate(url),

		//Wait for whole body to be ready
		//The original method to wait for a special block comment only
		//But there are too many edge case so I do it for sure
		chromedp.WaitReady("body", chromedp.ByQuery),

		// click show more comment . Don't know how to speed this up in js part yet
		// Also can't make a simple loop here. Need to check chromedp syntax a bit
		// Look silly but ok
		chromedp.Evaluate(js, empty_place_holder),
		chromedp.Evaluate(js, empty_place_holder),
		chromedp.Evaluate(js, empty_place_holder),
		chromedp.Evaluate(js, empty_place_holder),
		chromedp.Evaluate(js, empty_place_holder),
		chromedp.Sleep(time.Millisecond*50),

		chromedp.OuterHTML(`*`, &comment, chromedp.ByQuery),
	)
	if err != nil {
		log.Fatal(err)
	}
	return comment
}

func is_old_url(url, date_jqSelector string) bool {
	tmp_date := "/11/2023"
	resp, err := http.Get(url)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()
	cdoc, _ := goquery.NewDocumentFromReader(resp.Body)
	return !strings.Contains(cdoc.Find(date_jqSelector).Text(), tmp_date)

}

func rank_tuoitre(url string) (map[string]int, error) {
	var result = make(map[string]int)
	if is_old_url(url, ".detail-time") {
		return result, errors.New("skipped old page")
	}
	//Else, continue

	//Can't set var here because it will raise an error when we click multiple times
	//I don't know, js things
	js := `
		if (document.querySelector('.commentpopupall')) {
			document.querySelector('.commentpopupall').click();
		}
	`
	fullbody := click_n_get(url, js)
	fmt.Println(fullbody)

	return result, nil
}

func rank_vnexpress(url string) (map[string]int, error) {
	var result = make(map[string]int)
	var total_likes int

	if is_old_url(url, ".date") {
		return result, errors.New("skipped old page")
	}

	//Else, continue

	js := `
		if (document.querySelector('.txt_666')) {
			document.querySelector('.txt_666').click();
		}
	`
	comment := click_n_get(url, js)

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(comment))
	if err != nil {
		log.Fatal(err)
	}

	//Parse html. Very vnexpress specific

	doc.Find(".reactions-total").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the number
		number := s.Find(".number").Text()
		if number != "" {
			//fmt.Printf("Total like for this comment %s\n", number)
			num, err := strconv.Atoi(strings.ReplaceAll(number, ".", ""))
			if err != nil {
				log.Fatal(err)
			}
			total_likes += num
		}
	})

	result[url] = total_likes
	return result, nil
}
