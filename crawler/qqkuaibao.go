package crawler

import (
	"github.com/xfort/rockgo"
	"github.com/xfort/RockCrawler/obj"
	"net/http"
	"errors"
	"github.com/xfort/RockCrawler/db"
	"net/url"
	"github.com/bitly/go-simplejson"
	"fmt"
	"io"
	"context"
	"github.com/PuerkitoBio/goquery"
	"strings"
	"log"
)

const KB_API_MediaCardInfo = "http://kuaibao.qq.com/getMediaCardInfo?chlid="

type QQKuaibaoCrawler struct {
	CrawlerObj

	resArticlesChan chan *obj.ArticleObj
}

func (qqkb *QQKuaibaoCrawler) InitQQKB(rockhttp *rockgo.RockHttp, qqkbdb *db.ArticleObjDB) error {
	qqkb.LoadArticles = qqkb.loadArticlesByTasks
	qqkb.PublishArticles = qqkb.publishArticles

	qqkb.TypeName = obj.Site_QQKuaibao
	return qqkb.CrawlerObj.Init(rockhttp, qqkbdb)
}

func (qqkb *QQKuaibaoCrawler) loadArticlesByTasks(taskObj *obj.TaskObj) ([]*obj.ArticleObj, error) {

	if taskObj.TaskUrl == "" {
		return nil, errors.New("任务url为空")
	}
	articleList, err := qqkb.LoadHomeArticlesByUrl(taskObj.TaskUrl, nil)
	if err != nil {
		return nil, err
	}

	for _, item := range articleList {
		item, err := qqkb.LoadArticleDetail(item.SourceWebUrl, item)
		if err != nil {
			item.SourceHtml = err.Error()
		}
	}

	return articleList, nil
}

func (qqkb *QQKuaibaoCrawler) publishArticles(artilces []*obj.ArticleObj) error {
	//TODO
	return nil
}

//读取文章所有数据
func (qqkb *QQKuaibaoCrawler) LoadHomeArticlesByUrl(homeUrl string, header http.Header) ([]*obj.ArticleObj, error) {

	urlValue, err := url.Parse(homeUrl)
	if err != nil {
		return nil, err
	}
	chlId := urlValue.Query().Get("chlid")

	if chlId == "" {
		return nil, errors.New("解析url中的chlid为空_" + homeUrl)
	}
	resByte, err, response := qqkb.CoHttp.GetBytes(KB_API_MediaCardInfo+chlId, &header)

	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		return nil, errors.New("读取快报主页http失败_" + response.Status + "_" + chlId)
	}

	resJson, err := simplejson.NewJson(resByte)
	if err != nil {
		return nil, err
	}

	newsListArray := resJson.GetPath("info", "newsList")
	tmpArray, err := newsListArray.Array()
	if err != nil {
		return nil, errors.New(fmt.Sprintln("解析快报主页newsList失败_", err, string(resByte)))
	}
	newsLen := len(tmpArray)
	articleObjArray := make([]*obj.ArticleObj, 0, newsLen+2)

	for index := 0; index < newsLen; index++ {
		itemJson := newsListArray.GetIndex(index)
		articleObj := obj.ObtainArticleObj()
		articleObj.SourceSiteCode = obj.Site_QQKuaibao_Code
		articleObj.SourceSiteName = obj.Site_QQKuaibao

		articleObj.SourceId = itemJson.Get("id").MustString("")
		articleObj.Title = itemJson.Get("title").MustString("")
		articleObj.ThumbnailsUrl = itemJson.Get("thumbnails_qqnews").GetIndex(0).MustString("")
		articleObj.SourceWebUrl = itemJson.Get("url").MustString("")
		articleObj.UserObj.SourceId = itemJson.Get("chlid").MustString("")
		articleObj.UserObj.Nickname = itemJson.Get("chlname").MustString("")
		articleObj.SourcePubtimestr = itemJson.Get("time").MustString("")
		articleObj.SourcePubtimestamp = itemJson.Get("timestamp").MustInt64()
		articleObjArray = append(articleObjArray, articleObj)
	}

	return articleObjArray, nil
}

//读取文章内容详细数据,先读取数据库然后从源地址度去解析
func (qqkb *QQKuaibaoCrawler) LoadArticleDetail(articleUrl string, article *obj.ArticleObj) (*obj.ArticleObj, error) {
	var err error
	article, err = qqkb.loadArticleDetailFromDB(article)
	if err != nil || article.DBId <= 0 || article.ContentHtml == "" {
		article, err = qqkb.loadArticleDetailFromUrl(articleUrl, article)
		dbID, _, err := qqkb.CoDB.InsertArticleIfNotExistBySourceId(article)
		if err != nil {
			log.Println("添加文章到数据库失败_" + err.Error())
		}
		article.DBId = dbID
	}
	return article, err
}

//读取文章内容详细数据，从数据库
func (qqkb *QQKuaibaoCrawler) loadArticleDetailFromDB(article *obj.ArticleObj) (*obj.ArticleObj, error) {

	dbid, err := qqkb.CoDB.QueryArticleBySourceId(context.TODO(), article.SourceId, article)
	if err != nil {
		return article, err
	}

	if dbid > 0 {
		article.DBId = dbid
	}
	return article, nil
}

//读取文章内容详细数据
func (qqkb *QQKuaibaoCrawler) loadArticleDetailFromUrl(articleUrl string, article *obj.ArticleObj) (*obj.ArticleObj, error) {
	article, err := qqkb.loadArticleSourceHtml(articleUrl, article)
	if err != nil {
		return article, err
	}
	article, err = qqkb.parseArticleSourceHtml(strings.NewReader(article.SourceHtml), article)
	return article, err
}

//读取文章详情页内容
func (qqkb *QQKuaibaoCrawler) loadArticleSourceHtml(articleUrl string, article *obj.ArticleObj) (*obj.ArticleObj, error) {
	resByte, err, response := qqkb.CoHttp.GetBytes(articleUrl, nil)
	if err != nil {
		return article, err
	}
	if response.StatusCode != 200 {
		return article, errors.New("http结果状态码异常——" + response.Status)
	}
	article.SourceHtml = string(resByte)
	return article, nil
}

//解析文章详情页内容
func (qqkb *QQKuaibaoCrawler) parseArticleSourceHtml(r io.Reader, article *obj.ArticleObj) (*obj.ArticleObj, error) {

	htmlDoc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return article, nil
	}
	if article.Title == "" {
		article.Title = htmlDoc.Find("p.title").First().Text()
	}

	if article.UserObj.Nickname == "" {
		article.UserObj.Nickname = htmlDoc.Find("title").First().Text()
	}

	if article.SourcePubtimestr == "" {
		article.SourcePubtimestr = htmlDoc.Find("div.src").First().Text()
		article.SourcePubtimestr = strings.TrimSpace(article.SourcePubtimestr)
		article.SourcePubtimestr = strings.Replace(article.SourcePubtimestr, article.UserObj.Nickname, "", 2)
	}

	contentNode := htmlDoc.Find("div#content")
	if contentNode.Length() <= 0 {
		return article, errors.New("解析快报新闻详情页的div#content失败")
	}
	contentNode = contentNode.First()

	contentNode.Children().First().Remove()
	contentNode.Children().First().Remove()
	contentNode.Find("div#share").First().Remove()

	childLen := contentNode.Children().Length()
	for index := 0; index < childLen; index++ {
		itemNode := contentNode.Children().Eq(index)
		//fmt.Println("node_data", itemNode.Text())
		itemNode.RemoveAttr("class")
		if itemNode.Text() == "本文来自腾讯新闻客户端自媒体，不代表腾讯新闻的观点和立场" {
			itemNode.Remove()
		}
	}

	contentStr, err := contentNode.Html()
	if err != nil {
		return article, err
	}
	contentStr = strings.Replace(contentStr, `<!--Added by nonysun at 2014/03/06-->`, "", 1)
	article.ContentHtml = strings.TrimSpace(contentStr)

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
