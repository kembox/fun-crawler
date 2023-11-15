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
4. use `goquery` to parse html and calculate the total number of "likes" based on some predefined querySelector for the domains I current supported. See [this](https://www.w3schools.com/cssref/css_selectors.php) for more info about querySelector if you're not familiar with it ( Me too )

### TODO
- Consider speed up process by goroutines. Though we can be blocked 
- Check if we can have smarter waiting mechanism while loading page. So far we wait for full body or static sleep because there are too many edge cases and I have quite a time constraint

## How to generate input urls to check and score

- Get input urls from wayback machine using [tomnomnom/waybackurls](https://github.com/tomnomnom/waybackurls)

```bash
waybackurls -dates -no-subs vnexpress.net | fgrep '2023-11-' | egrep "html?$" | awk '{print $2}' | sort -n | uniq  > wayback.list
```
( Temporarily get data from Nov 2023 first. Then my crawler will parse date info in url and skip the old one later )

- Crawl info directly from the site to get latest urls
Using [hakrawker](https://github.com/hakluke/hakrawler)

```shell
echo "https://vnexpress.net" | hakrawler  | fgrep 'https://vnexpress.net' | egrep "html?$"  | awk '{print $2}' | sort -n | uniq > hakrawler.list
```

- Merge those 2 list:
```shell
cat wayback.list hakrawler.list | sort -n | uniq > ready_to_rank.list
```

### Use my fun-crawler to process the urls

- pipe through my fun-crawler to get url:number_of_likes data
```shell
cat ready_to_rank_list | ./fun_crawler -resume -outfile "file_location_of_your_choice"
```

- Sort top 10
```shell
cat results | sort -t ':' -k3 -n | tail -10
```
See my quickstart [start.sh](https://github.com/kembox/fun-crawler/blob/main/start.sh) script for more info

### Sample output

```
#vnexpress.net
https://vnexpress.net/nguoi-phu-nu-mo-3-quan-com-2-000-dong-de-tra-on-doi-4675595.html 6174
https://vnexpress.net/de-xuat-can-nhac-quy-dinh-cam-nguoi-co-nong-do-con-lai-xe-4675223.html 9299
https://vnexpress.net/de-xuat-can-nhac-quy-dinh-da-uong-ruou-bia-khong-lai-xe-4675223.html 9299
https://vnexpress.net/can-lang-nghe-du-luan-ve-phim-dat-rung-phuong-nam-thay-vi-doi-xu-ly-4674242-tong-thuat.html 9666
https://vnexpress.net/can-lang-nghe-du-luan-ve-phim-dat-rung-phuong-nam-thay-vi-doi-xu-ly-4674242.html 9666
https://vnexpress.net/quoc-hoi-tiep-tuc-chat-van-cac-bo-truong-y-te-giao-duc-van-hoa-4674242.html 9666
https://vnexpress.net/co-nen-ly-hon-khi-vo-cai-me-chong-4674713.html 9889
https://vnexpress.net/bo-truong-nguyen-van-hung-can-xu-ly-nguoi-boi-xau-phim-dat-rung-phuong-nam-4674165.html 10173
https://vnexpress.net/ong-pham-nhat-vuong-gap-ty-phu-giau-thu-hai-an-do-4674636.html 11443
https://vnexpress.net/tai-xe-oto-tong-lien-hoan-o-sai-gon-khai-vua-nhau-xong-4676224.html 13668

#tuoitre.vn
https://tuoitre.vn/bi-mat-nao-giup-hang-trung-quoc-giao-cho-khach-viet-nam-qua-nhanh-qua-re-20231107225659139.htm 224
https://tuoitre.vn/bat-tam-giam-tai-xe-gay-tai-nan-lien-hoan-tong-chet-co-gai-o-thu-duc-20231113131641939.htm 226
https://tuoitre.vn/bi-to-mua-bai-bao-khoa-hoc-truong-dh-ton-duc-thang-len-tieng-20231110231101767.htm 265
https://tuoitre.vn/hai-bo-truong-len-tieng-viec-cong-dong-mang-bat-nat-y-nhi-dat-rung-phuong-nam-20231107170646089.htm 373
https://tuoitre.vn/ngan-ngam-voi-nan-xa-rac-bua-bai-o-pho-di-bo-nguyen-hue-20231113080303766.htm 407
https://tuoitre.vn/cam-tai-xe-tuyet-doi-khong-co-nong-do-con-trong-hoi-tho-lieu-qua-nghiem-khac-20231110084643579.htm 467
https://tuoitre.vn/cam-tuyet-doi-nguoi-co-nong-do-con-lai-xe-lieu-qua-nghiem-khac-20231110084643579.htm 467
https://tuoitre.vn/hang-trieu-giay-phep-lai-xe-dang-su-dung-binh-thuong-vi-sao-phai-doi--20231110142933651.htm 500
https://tuoitre.vn/cuc-csgt-doi-quy-dinh-dan-choi-bien-so-dep-noi-bi-sup-ham-20231110062012106.htm 527
https://tuoitre.vn/chua-co-bao-gio-cho-e-nhu-hien-nay-20231112080633584.htm 919
```