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
	AddLog        func(int, ...interface{})
}

func (pub *PublishObj) Init(httpobj *rockgo.RockHttp, dbobj *db.ArticleObjDB) error {
	var err error
	pub.execpath, err = os.Getwd()

	if err != nil {
		return err
	}

	if pub.DBObj == nil && dbobj == nil {
		return rockgo.NewError("发布器初始化失败，数据库不能为空")
	}

	pub.sourceArticleChan = make(chan *obj.ArticleObj, 20480)
	pub.DBObj = dbobj
	return nil
}

func (pub *PublishObj) Start() {
	go pub.handleArticles()
}

func (pub *PublishObj) AddArticle(article *obj.ArticleObj) {
	pub.sourceArticleChan <- article
}

func (pub *PublishObj) AddArticles(articles []*obj.ArticleObj) {

	for _, item := range articles {
		if item != nil {
			pub.sourceArticleChan <- item
		}
	}
}

func (pub *PublishObj) handleArticles() {
	for {
		item, ok := <-pub.sourceArticleChan

		if !ok {
			break
		}
		if item.TaskObj.PublishCode == 0 {
			pub.AddLog(rockgo.Log_Info, "文章配置为不发布", item.Title, item.SourceSiteName, item.SourceWebUrl)
			continue
		}
		pub.handleArticle(item, item.TaskObj)
	}
}
