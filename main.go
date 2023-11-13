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

/*
var result_file = "./vne_result.txt"
var checked_urls_file = "./checked_urls.txt"
*/
var result_file = "./test_vne_result.txt"
var checked_urls_file = "./test_checked_urls.txt"

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
			//score_result, err := rank_vnexpress(url)
			score_result, err := rank_tuoitre(url)
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
func click_n_get(url, js string, extra_wait_milisec int) string {
	var comment string
	var empty_place_holder interface{}

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
		chromedp.Evaluate(js, empty_place_holder),
		chromedp.Evaluate(js, empty_place_holder),
		chromedp.Evaluate(js, empty_place_holder),
		chromedp.Evaluate(js, empty_place_holder),
		chromedp.Sleep(time.Millisecond*time.Duration(extra_wait_milisec)),
		chromedp.WaitVisible(`body`, chromedp.ByQuery),

		chromedp.OuterHTML(`body`, &comment, chromedp.ByQuery),
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

func rank_vnexpress(url string) (map[string]int, error) {
	var result = make(map[string]int)

	if is_old_url(url, ".date") {
		return result, errors.New("skipped old page")
	}

	//Else, continue

	js := `
		if (document.querySelector('.txt_666')) {
			document.querySelector('.txt_666').click();
		}
	`
	comment := click_n_get(url, js, 1)
	//vnexpress doesn't need to have any extra wait.
	//I put it 1 milliseconds here to satisfy function definition :|

	//The selector that we use to select the needed content in console
	//for example: document.querySelector(".number")
	like_box_selector := ".reactions-total"
	like_count_selector := ".number"
	result[url] = count_likes(comment, like_box_selector, like_count_selector)

	//result[url] = total_likes
	return result, nil
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
		if (document.querySelector('.viewmore-comment')) {
			document.querySelector('.viewmore-comment').click();
		}
	`
	fullbody := click_n_get(url, js, 2000)
	//Tuoitre has some lazy load magic that chromedp can't just simply wait by matching selector
	//I have to use this stupid trick :(
	//Yes - 2 seconds is pretty safe

	//The selector that we use to select the needed content in console
	//for example: document.querySelector(".number")
	like_box_selector := ".totalreact"
	like_count_selector := ".total"
	result[url] = count_likes(fullbody, like_box_selector, like_count_selector)

	return result, nil
}

func count_likes(body_html, like_box_selector, like_count_selector string) (total_likes int) {

	total_likes = 0

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body_html))
	if err != nil {
		log.Fatal(err)
	}

	//select 2 class to make sure it's the correct place to check
	doc.Find(like_box_selector).Each(func(i int, s *goquery.Selection) {
		// For each item found, get the number
		number := s.Find(like_count_selector).Text()
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
