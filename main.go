package main

import (
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

	//Nothing special, just to check how to manage defaults options
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

	js := `
		document.querySelector('.txt_666').click();
	`

	var result = make(map[string]int)
	var total_likes int
	var comment string
	var empty_place_holder interface{}
	var url string = `https://vnexpress.net/xuyen-dem-dau-gia-ba-mo-cat-o-ha-noi-4673746.html`

	err := chromedp.Run(ctx,

		chromedp.Navigate(url),

		// wait for comment box element is visible
		chromedp.WaitVisible(`.box_comment_vne`, chromedp.ByQuery),
		//chromedp.WaitReady(`.box_comment_vne`, chromedp.ByQuery),

		// click show more comment . Don't know how to speed this up in js part yet
		chromedp.Evaluate(js, empty_place_holder),
		chromedp.Sleep(time.Millisecond*50),
		chromedp.Evaluate(js, empty_place_holder),
		chromedp.Sleep(time.Millisecond*50),
		chromedp.Evaluate(js, empty_place_holder),
		chromedp.Sleep(time.Millisecond*50),

		chromedp.OuterHTML(`.box_comment_vne`, &comment, chromedp.ByQuery),
	)

	if err != nil {
		log.Fatal(err)
	}

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
	fmt.Println(result)

	f, _ := os.Create("./result.txt")
	for k, v := range result {
		f.WriteString(k + ":" + strconv.Itoa(v) + "\n")
	}
}
