package crawler

import (
	"testing"
	"github.com/xfort/rockgo"
	"log"
)

func TestDuowanCrawler(t *testing.T) {
	dwCrawler := &DuowanCrawler{}
	dwCrawler.TypeName = "duowan"
	err := dwCrawler.Init(rockgo.NewRockHttp(), nil)
	if err != nil {
		log.Fatalln(err)
	}
	dwCrawler.Start()
	for {
		item, ok := dwCrawler.GetOutArticle()
		if !ok {
			break
		}
		log.Println(item)
	}
}
