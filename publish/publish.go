package publish

import (
	"github.com/xfort/rockgo"
	"github.com/xfort/RockCrawler/db"

	"os"
	"github.com/xfort/RockCrawler/obj"
)

//发布基础类
type PublishObj struct {
	HttpObj *rockgo.RockHttp
	DBObj   *db.ArticleObjDB

	sourceArticleChan chan *obj.ArticleObj
	execpath          string

	handleArticle func(*obj.ArticleObj, *obj.TaskObj) error
}

func (pub *PublishObj) Init(httpobj *rockgo.RockHttp, dbobj *db.ArticleObjDB) error {
	var err error
	pub.execpath, err = os.Getwd()

	if err != nil {
		return err
	}

	pub.sourceArticleChan = make(chan *obj.ArticleObj, 2048)

	if dbobj == nil {

	} else {
		pub.DBObj = dbobj
	}
	return nil
}

func (pub *PublishObj) Start() {
	go pub.handleArticles()
}

func (pub *PublishObj) AddArticle(article *obj.ArticleObj) {
	pub.sourceArticleChan <- article
}

func (pub *PublishObj) handleArticles() {
	for {
		item, ok := <-pub.sourceArticleChan
		if !ok {
			break
		}
		pub.handleArticle(item, item.TaskObj)
		//if err != nil {
		//
		//}
	}
}
