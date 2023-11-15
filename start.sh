#!/bin/bash

#Script to automate steps for:
# - generating url input from wayback machine and crawling directly from the site

domain="$1"

ts=$(date +%Y%m%d%H)
base_dir="/var/tmp/fun-crawler/${domain}/${ts}"
log_file=${base_dir}/logs.txt
input_dir=${base_dir}/input
results_output=${base_dir}/results.txt
mkdir -p $base_dir
mkdir -p $input_dir


function log() {
    echo $(date) "$@" | tee ${log_file}
}

function wayback() {
    domain="$1"
    waybackurls -dates -no-subs ${domain} | fgrep '2023-11-' | egrep "html?$" | awk '{print $2}' | sort -n | uniq  > $input_dir/${domain}_wayback.list
}
function by_hakrawler() {
    domain="$1"
    echo "https://${domain}" | hakrawler  | grep "https://${domain}" | egrep "html?$"  | awk '{print $2}' | sort -n | uniq > $input_dir/${domain}_hakrawler.list

}

log "Start processing: base dir will be in ${base_dir}"

log "Grabbing ${domain} data from wayback machine"
wayback $domain 
log "Crawl ${domain} by hakrawler"
by_hakrawler $domain

log "Merging input"
cat $input_dir/* | sort -n | uniq  > $base_dir/ready_to_rank.list

log "Start scraping url and collect likes"
cat $base_dir/ready_to_rank.list  | ./fun-crawler -resume -outfile "${results_output}" 

log "Sorting result"
sort -t " " -k2 -n ${results_output} -r | head -10
