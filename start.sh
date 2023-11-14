#!/bin/bash

#Script to automate steps for:
# - generating url input from wayback machine and crawling directly from the site

domain="$1"

ts=$(date +%Y%m%d%H%M)
base_dir="/tmp"
data_dir=${base_dir}/$ts

mkdir -p $data_dir

function generate_input(
    domain="$1"
    waybackurls -dates -no-subs ${domain} | fgrep '2023-11-' | fgrep 'html$' | awk '{print $2}' | sort -n | uniq  > $data_dir/wayback.list
    echo "https://${domain}" | hakrawler  | fgrep 'https://${domain}' | grep "html$"  | awk '{print $2}' | sort -n | uniq > $data_dir/hakrawler.list

    cat $data_dir/* | sort -n | uniq  > $data_dir/ready_to_rank.list
)

generate_input()

#Start scraping url and collect likes
cat $data_dir/ready_to_rank  | ./fun_crawler  

