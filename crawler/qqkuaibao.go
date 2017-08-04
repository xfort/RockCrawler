package crawler

import (
	"github.com/xfort/rockgo"
	"github.com/xfort/RockCrawler/obj"
	"net/http"
	"errors"
	"github.com/xfort/RockCrawler/db"
)

type QQKuaibaoCrawler struct {
	CrawlerObj

	resArticlesChan chan *obj.ArticleObj
}

func (qqkb *QQKuaibaoCrawler) Init(rockhttp *rockgo.RockHttp, qqkbdb *db.ArticleObjDB) error {
	qqkb.LoadArticles = qqkb.loadArticlesByTasks
	qqkb.PublishArticles = qqkb.publishArticles

	qqkb.TypeName = obj.Site_QQKuaibao
	return qqkb.CrawlerObj.Init(rockhttp, qqkbdb)
}

func (qqkb *QQKuaibaoCrawler) loadArticlesByTasks(taskObj *obj.TaskObj) ([]*obj.ArticleObj, error) {

	return nil, nil
}

func (qqkb *QQKuaibaoCrawler) publishArticles(artilces []*obj.ArticleObj) error {
	return nil
}

//读取文章所有数据
func (qqkb *QQKuaibaoCrawler) LoadHomeArticlesByUrl(homeUrl string, header http.Header) ([]*obj.ArticleObj, error) {
	return nil, nil
}

//读取文章列表数据
//不包含 文章内容数据
func (qqkb *QQKuaibaoCrawler) LoadArticlesByUrl(homeurl string, header http.Header) ([]*obj.ArticleObj, error) {

	return nil, nil
}

//读取文章内容详细数据,从源地址
func (qqkb *QQKuaibaoCrawler) LoadArticleDetail(articleUrl string, article *obj.ArticleObj) (*obj.ArticleObj, error) {

	var err error
	article, err = qqkb.loadArticleDetailFromDB(article)
	if err != nil {
		article, err = qqkb.loadArticleDetailFromUrl(articleUrl, article)
	}

	return article, err
}

//读取文章内容详细数据，从数据库
func (qqkb *QQKuaibaoCrawler) loadArticleDetailFromDB(article *obj.ArticleObj) (*obj.ArticleObj, error) {

	return article, nil
}

//读取文章内容详细数据，从数据库
func (qqkb *QQKuaibaoCrawler) loadArticleDetailFromUrl(articleUrl string, article *obj.ArticleObj) (*obj.ArticleObj, error) {

	return article, nil
}

//读取结果数据，用于从其它协成读取结果数据
func (qqkb *QQKuaibaoCrawler) GetResultArticle() (*obj.ArticleObj, error) {
	article, ok := <-qqkb.resArticlesChan
	if !ok {
		return nil, errors.New("文章结果chan已被关闭")
	}
	return article, nil
}

func (qqkb *QQKuaibaoCrawler) addResArticles(artilces []*obj.ArticleObj) {

}
