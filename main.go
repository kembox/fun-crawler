// Command click is a chromedp comment demonstrating how to use a selector to
// click on an element.
package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
)

func main() {

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.NoDefaultBrowserCheck,
	)

	// new browser, first tab
	browserCtx, browserCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer browserCancel()

	// create chrome instance
	ctx, cancel := chromedp.NewContext(browserCtx)
	defer cancel()

	// create a timeout
	ctx, cancel = context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	js := `
		document.querySelector('.txt_666').click();
	`
	// navigate to a page, wait for an element, click
	var comment string
	var empty_place_holder interface{}
	err := chromedp.Run(ctx,
		//chromedp.Navigate(`https://pkg.go.dev/time`),
		//chromedp.Navigate(`https://vnexpress.net/du-hoc-binh-dan-4673938.html`),
		chromedp.Navigate(`https://vnexpress.net/xuyen-dem-dau-gia-ba-mo-cat-o-ha-noi-4673746.html`),

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

		//chromedp.Text(`.box_comment_vne.width_common`, &comment, chromedp.ByQuery),
		//chromedp.Text(`.box_comment_vne`, &comment, chromedp.ByQuery),
		chromedp.OuterHTML(`.box_comment_vne`, &comment, chromedp.ByQuery),
	)
	if err != nil {
		log.Fatal(err)
	}
	//log.Printf("%s", comment)

	/*
		    doc.Find(".left-content article .post-title").Each(func(i int, s *goquery.Selection) {
				// For each item found, get the title
				title := s.Find("a").Text()
				fmt.Printf("Review %d: %s\n", i, title)
			})
	*/
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(comment))
	if err != nil {
		log.Fatal(err)
	}

	doc.Find(".reactions-total").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the number
		number := s.Find(".number").Text()
		if number != "" {
			fmt.Printf("Total like for this comment %s\n", number)
		}
	})
	//fmt.Printf("%T\n", doc)
}
