package main

import (
	"github.com/xfort/RockCrawler/crawler"
	"github.com/xfort/rockgo"
	"log"
	"encoding/json"
)

func main() {
	startMiaoPai()
}

func startMiaoPai() {
	mpCrawler := &crawler.MiaopaiCrawler{}
	err := mpCrawler.Init(rockgo.NewRockHttp(), nil)
	if err != nil {
		log.Fatalln(err)
	}
	mpCrawler.Start()

	for {
		resArticle, ok := mpCrawler.GetResArticle()
		if !ok {
			break
		}
		resBytes, err := json.Marshal(resArticle)
		log.Println(err, string(resBytes))
	}
}
