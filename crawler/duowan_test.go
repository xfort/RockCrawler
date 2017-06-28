package crawler

import (
	"testing"
	"github.com/xfort/rockgo"
	"log"
	"github.com/xfort/RockCrawler/obj"
)

//func TestDuowanCrawler(t *testing.T) {
//	dwCrawler := &DuowanCrawler{}
//	dwCrawler.TypeName = "duowan"
//	err := dwCrawler.Init(rockgo.NewRockHttp(), nil)
//	if err != nil {
//		log.Fatalln(err)
//	}
//	dwCrawler.Start()
//	for {
//		item, ok := dwCrawler.GetOutArticle()
//		if !ok {
//			break
//		}
//		log.Println(item)
//	}
//}

func TestDuowanCrawler_LoadArticleDetail(t *testing.T) {

	dwCrawler := &DuowanCrawler{}
	dwCrawler.TypeName = "duowan"
	dwCrawler.CoHttp = rockgo.NewRockHttp()
	var err error
	article := &obj.ArticleObj{}
	article.SourceWebUrl = "http://wzry.duowan.com/1706/362573191281.html"

	article, err = dwCrawler.loadArticleDetail(article)
	log.Println(err)
	log.Println(article.ContentHtml)
}
