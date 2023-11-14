# fun-crawler
An exercise to scrape and rank url based on number of "likes" in its comments

## Quick start

1. Clone this repo. Build it with: `go build`
2. Install [tomnomnom/waybackurls](https://github.com/tomnomnom/waybackurls) and [hakrawker](https://github.com/hakluke/hakrawler)
3. Run `./start.sh tuoitre.vn` or `./start.sh vnexpress.net`

## How fun-crawler works

### Dependencies
- CLI tools:
    - [tomnomnom/waybackurls](https://github.com/tomnomnom/waybackurls)
    - [hakrawker](https://github.com/hakluke/hakrawler)
- [chromedp](https://github.com/chromedp/chromedp) to handle javascript and clicks.
- Any chromium browser or [chrome docker-headless-shell](https://github.com/chromedp/docker-headless-shell) ( I'm using google chrome on my Kali machine ).
- [goquery](https://github.com/PuerkitoBio/goquery) go navigate through pages and parse html output.
- See `go.mod` for more info about libraries. 
- Stable internet.
### The main logic
1. It reads url[s] from stdin
2. Extract info about publish date. Check if it's older than 1 week
3. use `chromedp` to launch chrome instance to load page. 
  - Filter unnecessary resources by [cdproto/network](https://pkg.go.dev/github.com/chromedp/cdproto/network)
  - Perform several clicks to show all comments in comment sections. 
  - Return page data. 
4. use `goquery` to parse html and alculate the total number of "likes" based on some predefined querySelector for the domains I current supported

### TODO
- Consider speed up process by goroutines. Though we can be blocked 
- Check if we can have smarter waiting mechanism while loading page. So far we wait for full body or static sleep because there are too many edge cases and I have quite a time constraint

### Generating input urls to check and score

Get input urls from wayback machine using [tomnomnom/waybackurls](https://github.com/tomnomnom/waybackurls)

```bash
waybackurls -dates -no-subs vnexpress.net | fgrep '2023-11-' | fgrep 'html$' | awk '{print $2}' | sort -n | uniq  > wayback.list
```
( Temporarily get data from Nov 2023 first. Then my crawler will parse date info in url and skip the old one later )

Crawl info directly from the site to get latest urls
Using [hakrawker](https://github.com/hakluke/hakrawler)

```shell
echo "https://vnexpress.net" | hakrawler  | fgrep 'https://vnexpress.net' | grep "html$"  | awk '{print $2}' | sort -n | uniq > hakrawler.list
```

Merge those 2 list:
```shell
cat wayback.list hakrawler.list | sort -n | uniq > ready_to_rank.list
```

### Use my fun-crawler to process the urls

#pipe through my fun-crawler to get url:number_of_likes data
```shell
cat ready_to_rank_list | ./fun_crawler > results.txt
```

#Sort top 10
```shell
cat results | sort -t ':' -k3 -n | tail -10
```


