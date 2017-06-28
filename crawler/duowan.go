package crawler

import (
	"github.com/xfort/RockCrawler/obj"
	"github.com/xfort/rockgo"
	"github.com/xfort/RockCrawler/db"
	"net/http"
	"github.com/PuerkitoBio/goquery"
	"bytes"
	"net/url"
	"strings"
)

type DuowanCrawler struct {
	CrawlerObj
}

func (dw *DuowanCrawler) Init(cohttp *rockgo.RockHttp, codb *db.ArticleObjDB) error {
	dw.LoadArticles = dw.LoadHomeArtiles
	return dw.CrawlerObj.Init(cohttp, codb)
}

func (dw *DuowanCrawler) LoadHomeArtiles(task *obj.TaskObj) ([]*obj.ArticleObj, error) {

	articleArray, err := dw.loadArticleList(task.TaskUrl)
	if err != nil {
		return nil, err
	}
	articleArray, err = dw.deleteDuplicateArticle(articleArray)
	if err != nil {
		dw.AddLog(rockgo.Log_Error, "文章去重错误", err.Error())
	}

	if articleArray == nil || len(articleArray) <= 0 {
		return nil, rockgo.NewError("文章数组<=0", task.Name)
	}
	for _, item := range articleArray {
		if item == nil {
			continue
		}
		item, err = dw.loadArticleDetail(item)
		if err != nil {
			dw.AddLog(rockgo.Log_Error, "读取文章详情错误", err.Error(), task.Name, item.SourceWebUrl)
		}
		err = dw.insertArticle(item)
		if err != nil {
			dw.AddLog(rockgo.Log_Error, "添加文章到数据库错误", err.Error(), task.Name, item.SourceWebUrl)
		}
		item.TaskObj = task
	}
	return articleArray, nil
}

//读取文章列表
func (dw *DuowanCrawler) loadArticleList(urlstr string) ([]*obj.ArticleObj, error) {
	taskURL, err := url.Parse(urlstr)
	if err != nil {
		return nil, rockgo.NewError("无法解析url为URL", err.Error())
	}

	header := http.Header{}
	header.Set("user-agent", rockgo.UserAgent_Chrome_Web)
	resByte, err, response := dw.CoHttp.GetBytes(urlstr, &header)
	if err != nil {
		return nil, rockgo.NewError("读取文章列表错误", err.Error(), urlstr)
	}
	if response.StatusCode != 200 {
		return nil, rockgo.NewError("读取文章列表错误,http状态码", response.Status, urlstr)
	}
	htmlDoc, err := goquery.NewDocumentFromReader(bytes.NewReader(resByte))
	if err != nil {
		return nil, rockgo.NewError("http结果转为html doc错误", err.Error(), urlstr)
	}
	itemsNode := htmlDoc.Find("ul.news-list.news-list-page").Children()
	itemsLen := itemsNode.Length()
	if itemsLen <= 0 {
		return nil, rockgo.NewError("解析html中文章长度<=0", urlstr)
	}

	articleArray := make([]*obj.ArticleObj, 0, itemsLen+1)

	for index := 0; index < itemsLen; index++ {

		itemNode := itemsNode.Eq(index)
		article := obj.ObtainArticleObj()
		article.SourceSiteName = obj.Site_DuoWan
		article.SourceSiteTypeCode = obj.Site_DuoWan_Code

		article, err = dw.parseItemNode(itemNode, article, taskURL.Scheme+"://"+taskURL.Host)
		if err != nil {
			dw.AddLog(rockgo.Log_Warn, "解析文章html元素错误", err.Error(), urlstr)
		}
		articleArray = append(articleArray, article)
	}
	return articleArray, nil
}

func (dw *DuowanCrawler) parseItemNode(node *goquery.Selection, article *obj.ArticleObj, hosturl string) (*obj.ArticleObj, error) {
	node = node.Children().First()
	article.Title = node.AttrOr("title", "")

	article.SourceWebUrl = node.AttrOr("href", "")
	if article.SourceWebUrl == "" {
		return article, rockgo.NewError("解析单条文章node错误，href为空", article.Title)
	}
	article.SourceWebUrl = hosturl + article.SourceWebUrl
	return article, nil
}

//文章去重
func (dw *DuowanCrawler) deleteDuplicateArticle(articleArray []*obj.ArticleObj) ([]*obj.ArticleObj, error) {
	if articleArray == nil || len(articleArray) <= 0 {
		return articleArray, rockgo.NewError("文章数组长度<=0")
	}
	var err error
	urlMap := make(map[string]int, len(articleArray))
	for index, item := range articleArray {
		if item == nil {
			continue
		}

		if item.SourceWebUrl == "" {
			articleArray[index] = nil
			continue
		}
		if urlMap[item.SourceWebUrl] == 1 {
			articleArray[index] = nil
			continue
		}
		urlMap[item.SourceWebUrl] = 1

		item.DBId, err = dw.CoDB.QueryExistedArticle(item)
		if err != nil && item.DBId <= 0 {
			dw.AddLog(rockgo.Log_Error, "数据库查询文章是否存在发生错误", err.Error(), item.Title, item.SourceWebUrl, item.DBId)
			continue
		}
	}
	urlMap = nil
	return articleArray, nil
}

//添加到数据库
func (dw *DuowanCrawler) insertArticle(article *obj.ArticleObj) error {
	var err error
	article.DBId, err = dw.CoDB.InsertArticlce(article)

	return err
}

func (dw *DuowanCrawler) loadArticleDetail(article *obj.ArticleObj) (*obj.ArticleObj, error) {
	if article == nil || article.SourceWebUrl == "" {
		return article, rockgo.NewError("读取文章详情错误,文章为nil或url为空", article.Title)
	}

	header := &http.Header{}
	header.Set("user-agent", rockgo.UserAgent_Chrome_Web)
	resByte, err, response := dw.CoHttp.GetBytes(article.SourceWebUrl, header)
	if err != nil {
		return article, rockgo.NewError("读取文章详情错误", err.Error(), article.Title, article.SourceWebUrl)
	}

	if response.StatusCode != 200 {
		return article, rockgo.NewError("读取文章详情错误,http状态码", err.Error(), article.Title, article.SourceWebUrl)
	}
	article.SourceHtml = string(resByte)

	htmlDoc, err := goquery.NewDocumentFromReader(bytes.NewReader(resByte))
	if err != nil {
		return article, rockgo.NewError("文章详情转为htmldoc错误", err.Error(), article.Title, article.SourceWebUrl)
	}

	if article.Title == "" {
		htmlNode := htmlDoc.Find("article")
		if htmlNode.Length() > 0 {
			article.Title = htmlNode.First().Children().First().Text()
		}
	}
	htmlNode := htmlDoc.Find("address")

	if htmlNode.Length() > 0 {
		htmlNode = htmlNode.First().Children()
		nodeLen := htmlNode.Length()

		for index := 0; index < nodeLen; index++ {
			switch index {
			case 0: //日期时间
				article.SourcePubtimestr = htmlNode.Eq(index).Text()
			case 1: //来源
				article.SourceAuthor = htmlNode.Eq(index).Text()
				article.SourceAuthor = strings.Replace(article.SourceAuthor, `来源：`, "", 1)
			case 2: //作者
				article.Nickname = strings.Replace(htmlNode.Eq(index).Text(), `作者：`, "", 1)
				article.SourceAuthor = article.SourceAuthor + "_" + article.Nickname
			}
		}
	}

	htmlNode = htmlDoc.Find("div#text")
	if htmlNode.Length() <= 0 {
		return article, rockgo.NewError("提取文章内容错误div#text", article.Title, article.SourceWebUrl)
	}

	article.ContentHtml, err = dw.fixHtml(htmlNode.First())

	if err != nil {
		return article, rockgo.NewError("提取文章内容错误dhtml()", err.Error(), article.Title, article.SourceWebUrl)
	}
	article.ContentHtml = strings.TrimSpace(article.ContentHtml)
	article.SourceHtml = ""
	return article, nil
}

//修整html，删除<a>,<script> <style>元素，class属性
func (dw *DuowanCrawler) fixHtml(htmlNode *goquery.Selection) (string, error) {
	allNodes := htmlNode.Find("*")
	nodeLen := allNodes.Length()
	for index := 0; index < nodeLen; index++ {
		item := allNodes.Eq(index)

		if item.Is("script") || item.Is("style") {
			item.Remove()
		} else if item.Is("a") {
			item.ReplaceWithHtml(item.Text())
		} else {
			item.RemoveAttr("class")
			item.RemoveAttr("style")
		}
	}
	return htmlNode.Html()
}
