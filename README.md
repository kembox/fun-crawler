# fun-crawler
An exercise to scrape and rank url based on number of "likes" in its comments


### Generating input urls to check and score

Get input urls from wayback machine using (tomnomnom/waybackurls)[https://github.com/tomnomnom/waybackurls]

```bash
waybackurls -dates -no-subs vnexpress.net | fgrep '2023-11-' | fgrep 'html$' | awk '{print $2}' | sort -n | uniq  > wayback.list
```
( Temporarily get data from Nov 2023 first. Then my crawler will parse date info in url and skip the old one later )

Crawl info directly from the site to get latest urls
Using (hakrawker)[https://github.com/hakluke/hakrawler]

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


