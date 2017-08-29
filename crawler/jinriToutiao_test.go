package crawler

import (
	"testing"
	"github.com/xfort/rockgo"
	"log"
	"time"

	"github.com/xfort/RockCrawler/obj"
)

func TestJRTouTiao(t *testing.T) {
	jrTT := JinRiTouTiaoCrawler{}
	jrTT.DBDirPath = "/Users/xs/work/go/code/work/src/github.com/xfort/RockCrawler/data"
	jrTT.ConfigDirPath = "/Users/xs/work/go/code/work/src/github.com/xfort/RockCrawler"
	jrTT.PublishArticles = JRPublish

	err := jrTT.InitJR(rockgo.NewRockHttp(), nil)
	if err != nil {
		log.Fatalln("初始化失败", err)
	}
	jrTT.Start()

	time.Sleep(10 * time.Minute)
}

func JRPublish(articles []*obj.ArticleObj) error {
	for _, item := range articles {
		log.Println(item.Title,item.UserObj.SourceId)
	}
	return nil
}
