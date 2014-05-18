package main

import (
	"code.google.com/p/goauth2/oauth"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/google/go-github/github"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {

	token := flag.String("access-token", "", "oauth access token to github")
	filename := flag.String("out", "", "filename to write output")
	stars := flag.Int("stars", 10, "minimum number of starts a repo must have to be considered")
	flag.Parse()

	if *filename == "" {
		log.Println("Need an output filename.")
		flag.PrintDefaults()
		os.Exit(2)
	}

	var httpClient *http.Client
	if *token != "" {
		httpClient = (&oauth.Transport{
			Token: &oauth.Token{AccessToken: *token},
		}).Client()
	}
	client := github.NewClient(httpClient)

	out, err := os.Create(*filename)
	if err != nil {
		log.Fatalf("[ERROR] Opening output file: %v.", err)
	}
	defer out.Close()
	enc := json.NewEncoder(out)

	var repos []github.Repository

	opts := &github.SearchOptions{}
	opts.Page = 1
	opts.PerPage = 100
	for opts.Page > 0 {
		repRes, resp, err := client.Search.Repositories(fmt.Sprintf("stars:>=%d fork:true language:go", *stars), opts)
		if err != nil {
			log.Fatalf("[ERROR] Searching repositories: %v.", err)
		}
		log.Printf("[INFO] page=%d/%d\tbefore-rate=%d/%d", opts.Page, resp.LastPage, client.Rate.Remaining, client.Rate.Limit)

		repos = append(repos, repRes.Repositories...)

		if client.Rate.Remaining == 0 {
			diff := client.Rate.Reset.Sub(time.Now())
			log.Printf("[INFO] Rate limited, sleeping for %v.", diff)
			time.Sleep(diff)
		}

		opts.Page = resp.NextPage
	}

	if err := enc.Encode(repos); err != nil {
		log.Fatalf("[ERROR] Encoding JSON to output: %v.", err)
	}

}
