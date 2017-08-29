package db

import (
	"testing"

	"log"
	"github.com/xfort/RockCrawler/obj"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/net/context"
)

func TestArticleObjDB(t *testing.T) {

	testArticleObjDB_QueryArticlesByUser()
}
func testArticleObjDB_QueryArticlePublishStatus() {

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

func testArticleObjDB_QueryArticlesByUser() {
	articleDB := &ArticleObjDB{}
	err := articleDB.OpenDB("sqlite3", "/Users/xs/work/go/code/work/src/github.com/xfort/RockCrawler/data/QQKuaiBao.db")
	if err != nil {
		log.Fatalln(err)
	}

	articleList, user, err := articleDB.QueryArticlesByUser(context.TODO(), "5237828", "5237828")
	if err != nil {
		log.Fatalln(err)
	}

	log.Println(len(articleList))

	log.Println(user)
}
