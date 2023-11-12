package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
)

func main() {

	urls := bufio.NewReader(os.Stdin)
	for {
		url, err := urls.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		if len(strings.TrimSpace(url)) == 0 {
			break
		}

		fmt.Println("Start ranking vnexpress url")
		rank_vnexpress(url)
	}

}

func click_n_get(url, js, comment_block_selector string) string {
	//Nothing special, just to check how to manage defaults options
	var comment string
	var empty_place_holder interface{}

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.NoDefaultBrowserCheck,
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
	ctx, cancel = context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	err := chromedp.Run(ctx,

		chromedp.Navigate(url),

		// wait for comment box element is visible
		chromedp.WaitVisible(comment_block_selector, chromedp.ByQuery),

		// click show more comment . Don't know how to speed this up in js part yet
		// Also can't make a simple loop here. Need to check chromedp syntax a bit
		// Look silly but ok
		chromedp.Evaluate(js, empty_place_holder),
		chromedp.Sleep(time.Millisecond*50),
		chromedp.Evaluate(js, empty_place_holder),
		chromedp.Sleep(time.Millisecond*50),
		chromedp.Evaluate(js, empty_place_holder),
		chromedp.Sleep(time.Millisecond*50),

		chromedp.OuterHTML(comment_block_selector, &comment, chromedp.ByQuery),
	)
	if err != nil {
		log.Fatal(err)
	}
	return comment
}

func rank_vnexpress(url string) {
	var result = make(map[string]int)
	var total_likes int

	js := `
		document.querySelector('.txt_666').click();
	`

	comment_block_selector := `.box_comment_vne`

	comment := click_n_get(url, js, comment_block_selector)

	fmt.Println("Finished chrome")

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(comment))
	if err != nil {
		log.Fatal(err)
	}

	//fmt.Println(comment)

	doc.Find(".reactions-total").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the number
		number := s.Find(".number").Text()
		if number != "" {
			//fmt.Printf("Total like for this comment %s\n", number)
			num, err := strconv.Atoi(number)
			if err != nil {
				log.Fatal(err)
			}
			total_likes += num
		}
	})

	result[url] = total_likes
	fmt.Printf("Result: %v\n", result)

	f, _ := os.Create("./result.txt")
	for k, v := range result {
		f.WriteString(k + ":" + strconv.Itoa(v) + "\n")
	}
}
