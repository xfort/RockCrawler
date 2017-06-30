package main

import (
	"github.com/xfort/RockCrawler/crawler"
	"github.com/xfort/rockgo"
	"log"
	"encoding/json"
	"github.com/xfort/RockCrawler/publish"
)

func main() {
	//startMiaoPai()
	startDuoWan()
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

func startDuoWan() {
	dwCrawler := &crawler.DuowanCrawler{}
	dwCrawler.TypeName = "duowan"

	rockhttp := rockgo.NewRockHttp()
	xsPublisher := &publish.XSPublish{}
	xsPublisher.HttpObj = rockhttp

	dwCrawler.XSPublish = xsPublisher

	err := dwCrawler.Init(rockhttp, nil)

	if err != nil {
		log.Fatalln(err)
	}
	err = xsPublisher.Init(rockhttp, dwCrawler.CoDB)
	if err != nil {
		log.Fatalln(err)
	}

	xsPublisher.Start()

	dwCrawler.Start()
	for {
		item, ok := dwCrawler.GetOutArticle()
		if !ok {
			break
		}
		log.Println(item.Title)

	}
}
