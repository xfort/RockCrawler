package crawler

import (
	"github.com/xfort/rockgo"
	"github.com/xfort/RockCrawler/obj"
	"github.com/xfort/RockCrawler/db"
	"regexp"
	"net/url"
	"net/http"
	"errors"
	"log"
	"github.com/bitly/go-simplejson"
	"fmt"
	"time"
	"strconv"
)

const (
	JinRiTouTiao_Home_Url          = "http://www.toutiao.com/c/user/article/?"
	JinRiTouTiao_ArticleDetail_Url = "https://m.toutiao.com/i%s/info/"
)

//今日头条
type JinRiTouTiaoCrawler struct {
	CrawlerObj
}

func (jr *JinRiTouTiaoCrawler) InitJR(rockhttp *rockgo.RockHttp, articledb *db.ArticleObjDB) error {
	jr.LoadArticles = jr.loadJRArticles
	jr.TypeName = obj.Site_JinRiHL
	return jr.Init(rockhttp, articledb)
}

//采集主页文章
func (jr *JinRiTouTiaoCrawler) loadJRArticles(taskObj *obj.TaskObj) ([]*obj.ArticleObj, error) {
	userId, resJson, err := jr.loadArticlesData(taskObj.TaskUrl, "")
	if err != nil {
		return nil, err
	}
	dataJson := resJson.Get("data")
	artilceList, err := jr.parseArticlesData(userId, dataJson)
	if err != nil {
		return nil, err
	}
	nextMaxBehotTime := resJson.GetPath("next", "max_behot_time").MustInt64(0)

	if nextMaxBehotTime > 0 {
		userId, resJson, err := jr.loadArticlesData(taskObj.TaskUrl, strconv.FormatInt(nextMaxBehotTime, 10))
		if err != nil {
			return nil, err
		}
		dataJson := resJson.Get("data")
		artilceListTmp, err := jr.parseArticlesData(userId, dataJson)
		if err != nil {
			return nil, err
		}
		if len(artilceListTmp) > 0 {
			artilceList = append(artilceList, artilceListTmp...)
		}
	}

	artilceList, err = jr.loadArticleListDetail(userId, artilceList)
	return artilceList, err
}

//读取主页一组文章数据
func (jr *JinRiTouTiaoCrawler) loadArticlesData(homeUrl string, maxbehotTime string) (string, *simplejson.Json, error) {
	if homeUrl == "" {
		return "", nil, errors.New("任务url为空")
	}
	userIdArray := regexp.MustCompile(`user/(.+?)/`).FindStringSubmatch(homeUrl)

	userId := ""
	if len(userIdArray) < 2 {
		log.Println(userIdArray)
		return "", nil, errors.New("提取userid失败," + homeUrl)
	}

	userId = userIdArray[1]
	if maxbehotTime == "" {
		maxbehotTime = "0"
	}

	urlValue := url.Values{}
	urlValue.Set("page_type", "1")
	urlValue.Set("user_id", userId)
	urlValue.Set("max_behot_time", maxbehotTime)
	urlValue.Set("count", "30")

	apiUrl := JinRiTouTiao_Home_Url + urlValue.Encode()
	header := http.Header{}
	header.Set("user-agent", rockgo.UserAgent_Chrome_Web)

	resBytes, err, response := jr.CoHttp.GetBytes(apiUrl, &header)

	if err != nil {
		return userId, nil, errors.New("读取主页文章列表接口数据错误," + err.Error() + "_" + homeUrl)
	}

	if response.StatusCode != 200 {
		return userId, nil, errors.New("读取主页文章列表接口数据异常,http状态码异常" + response.Status + "_" + homeUrl)
	}

	resJson, err := simplejson.NewJson(resBytes)
	if err != nil {
		return userId, nil, errors.New("解析文章列表接口json数据错误，" + err.Error() + "_" + homeUrl)
	}

	dataJson := resJson.Get("data")
	dataLen := len(dataJson.MustArray())

	if dataLen <= 0 {
		return userId, nil, errors.New("文章列表接口json内data的长度为0," + string(resBytes) + "_" + apiUrl)
	}

	return userId, resJson, nil
}

//解析文章列表数据
func (jr *JinRiTouTiaoCrawler) parseArticlesData(userid string, dataJson *simplejson.Json) ([]*obj.ArticleObj, error) {
	dataLen := len(dataJson.MustArray())
	articleArray := make([]*obj.ArticleObj, 0, dataLen+1)
	for index := 0; index < dataLen; index++ {
		item := dataJson.GetIndex(index)
		article := obj.ObtainArticleObj()
		article.SourceSiteName = obj.Site_JinRiHL
		article.SourceSiteTypeCode = obj.Site_JinRiHL_Code
		article.UserObj.SourceId = userid

		article.Title = item.Get("title").MustString()
		article.ThumbnailsUrl = item.Get("image_url").MustString()
		article.Nickname = item.Get("source").MustString()
		article.SourceId = item.Get("item_id").MustString()
		article.SourceWebUrl = "http://www.toutiao.com" + item.Get("source_url").MustString()
		article.SourcePubtimestamp = item.Get("behot_time").MustInt64()
		if article.SourcePubtimestamp > 0 {
			article.SourcePubtimestr = time.Unix(article.SourcePubtimestamp, 0).String()
		}
		if item.Get("has_video").MustBool() {
			article.VideoSrc = item.Get("display_url").MustString()
		}
		articleArray = append(articleArray, article)
	}
	return articleArray, nil
}

//读取文章内容
func (jr *JinRiTouTiaoCrawler) loadArticleDetail(userid, sourceid string, article *obj.ArticleObj) (*obj.ArticleObj, error) {
	detailUrl := fmt.Sprintf(JinRiTouTiao_ArticleDetail_Url, sourceid)

	header := http.Header{}
	header.Set("user-agent", rockgo.UserAgent_Android)
	resBytes, err, _ := jr.CoHttp.GetBytes(detailUrl, &header)
	if err != nil {
		return article, err
	}
	return jr.parseArticleDetail(userid, resBytes, article)
}

//解析文章内容
func (jr *JinRiTouTiaoCrawler) parseArticleDetail(userid string, data []byte, article *obj.ArticleObj) (*obj.ArticleObj, error) {
	if article == nil {
		article = obj.ObtainArticleObj()
		article.SourceSiteName = obj.Site_JinRiHL
		article.SourceSiteTypeCode = obj.Site_JinRiHL_Code
	}
	if article.UserObj.SourceId == "" {
		article.UserObj.SourceId = userid
	}
	article.SourceHtml = string(data)
	dataJson, err := simplejson.NewJson(data)
	if err != nil {
		return article, err
	}

	if !dataJson.Get("success").MustBool(true) {
		return article, err
	}

	dataJson = dataJson.Get("data")
	if article.SourceAuthor == "" {
		article.SourceAuthor = dataJson.Get("detail_source").MustString()
	}
	if article.SourcePubtimestamp <= 0 {
		article.SourcePubtimestamp = dataJson.Get("publish_time").MustInt64(0)
	}

	if article.Title == "" {
		article.Title = dataJson.Get("title").MustString()
	}

	article.ContentHtml = dataJson.Get("content").MustString()

	return article, nil
}

//读取一组文章内容
func (jr *JinRiTouTiaoCrawler) loadArticleListDetail(userid string, articleList []*obj.ArticleObj) ([]*obj.ArticleObj, error) {

	if len(articleList) <= 0 {
		return nil, errors.New("文章数量<=0")
	}
	var err error
	resArticles := make([]*obj.ArticleObj, 0, len(articleList)+1)
	for _, item := range articleList {

		item.DBId, err = jr.CoDB.QueryExistedArticle(item)
		if err != nil {
			jr.AddLog(rockgo.Log_Warn, "查询文章采集状态失败", err, item.Title)
		}
		if item.DBId > 0 && item.SourceHtml != "" {
			item, err = jr.parseArticleDetail(userid, []byte(item.SourceHtml), item)
			if err != nil {
				jr.AddLog(rockgo.Log_Warn, "解析文章内容失败", item.Title, err)
			}
		}

		if item.ContentHtml == "" {
			item, err := jr.loadArticleDetail(userid, item.SourceId, item)
			if err != nil {
				jr.AddLog(rockgo.Log_Warn, "读取解析文章内容失败", item.Title, err)
			}
		}
		if item.Title != "" && item.ContentHtml != "" {
			resArticles = append(resArticles, item)

			if item.DBId <= 0 {
				item.DBId, _, err = jr.CoDB.InsertArticleIfNotExist(item)
				if err != nil {
					jr.AddLog(rockgo.Log_Warn, "保存头条文章到数据库失败", item.Title)
				}
			}
		}
	}
	return resArticles, nil
}
