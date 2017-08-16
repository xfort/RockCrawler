package db

import (
	"testing"

	"log"
	"github.com/xfort/RockCrawler/obj"

	_ "github.com/mattn/go-sqlite3"
)

func TestArticleObjDB_QueryArticlePublishStatus(t *testing.T) {

	articleDB := &ArticleObjDB{}
	err := articleDB.OpenDB("sqlite3", "/Users/xs/work/go/code/work/src/github.com/xfort/RockCrawler/data/wechat_game.db")
	if err != nil {
		log.Fatalln(err)
	}
	article := obj.ObtainArticleObj()
	article.Title = "王者荣耀 单体控制最强的四大杀手 被末位那位控住天美都解不开"

	status, err := articleDB.QueryArticlePublishStatus("weixin_game_wzry", article)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("结果", status)
}
