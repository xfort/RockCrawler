package crawler

import (
	"github.com/xfort/RockCrawler/obj"
	"github.com/xfort/rockgo"
	"github.com/xfort/RockCrawler/db"
	"github.com/pingcap/tidb/_vendor/src/github.com/juju/errors"
	"strings"
	"github.com/robertkrimen/otto"
	"github.com/PuerkitoBio/goquery"
	"bytes"
	"html"
	"net/url"
	"strconv"
	"time"
	"fmt"
	"net/http"
	"github.com/bitly/go-simplejson"
)

type UCCrawler struct {
	CrawlerObj
}

func (uc *UCCrawler) InitUC(rockhttp *rockgo.RockHttp, articleDB *db.ArticleObjDB) error {
	uc.LoadArticles = uc.LoadUCArticles
	uc.TypeName = obj.Site_UC
	return uc.Init(rockhttp, articleDB)
}

//采集文章列表
func (uc *UCCrawler) LoadUCArticles(taskObj *obj.TaskObj) ([]*obj.ArticleObj, error) {
	if taskObj.CollectCode <= 0 {
		return nil, errors.New("任务配置collect不为1，不采集")
	}

	if taskObj.TaskUrl == "" {
		return nil, errors.New("任务配置url为空")
	}
	homeUrl, err := url.Parse(taskObj.TaskUrl)
	if err != nil {
		return nil, err
	}

	mid := homeUrl.Query().Get("mid")
	if mid == "" {
		return nil, NewError("解析主页的mid错误", taskObj.TaskUrl, err, homeUrl)
	}

	appid, ucparams, err := uc.loadHomeUserData(taskObj.TaskUrl)
	if err != nil {
		return nil, err
	}
	articleAarray, err := uc.loadArticleList(mid, appid, ucparams, nil)
	if err != nil {
		return nil, err
	}

	articleAarray = uc.loadSaveArticleListDetail(appid, articleAarray)

	return articleAarray, nil
}

//读取主页关键数据
func (uc *UCCrawler) loadHomeUserData(homeurl string) (appid string, ucparamStr string, err error) {

	resBytes, err, _ := uc.CoHttp.GetBytes(homeurl, nil)
	if err != nil {
		return "", "", err
	}

	htmlDoc, err := goquery.NewDocumentFromReader(bytes.NewReader(resBytes))
	if err != nil {
		return "", "", err
	}

	scriptArray := htmlDoc.Find("script")
	scriptLen := scriptArray.Length()

	if scriptLen <= 0 {
		return "", "", rockgo.NewError("未找到script元素", homeurl)
	}

	var configScript string = "0"
	for index := 0; index < scriptLen; index++ {
		itemScript := scriptArray.Eq(index)

		scriptStr, err := itemScript.Html()
		if err != nil {
			continue
		}
		scriptStr = strings.TrimSpace(scriptStr)
		if strings.HasPrefix(scriptStr, `var globalConfig`) {
			configScript = scriptStr
			break
		}
	}

	if configScript == "0" {
		return "", "", errors.New("未发现globalConfig的script元素")
	}

	configScript = html.UnescapeString(configScript)
	jsVM, _, err := otto.Run(configScript)
	if err != nil {
		return "", "", NewError("执行globalConfig的script失败", err, homeurl, configScript)
	}
	var jsValue otto.Value
	if jsValue, err = jsVM.Run(`globalConfig['NAPI_APPID']`); err != nil {
		return "", "", NewError("获取globalConfig['NAPI_APPID']错误", err, homeurl)
	}
	if appid, err = jsValue.ToString(); err != nil || appid == "" {
		return "", "", NewError("获取globalConfig['NAPI_APPID']错误", err, homeurl)
	}
	if jsValue, err = jsVM.Run(`globalConfig['uc_param_str']`); err != nil {
		return "", "", NewError("获取globalConfig['uc_param_str']错误", err, homeurl)
	}
	if ucparamStr, err = jsValue.ToString(); err != nil || ucparamStr == "" {
		return "", "", NewError("获取globalConfig['uc_param_str']错误", err, homeurl)
	}
	return appid, ucparamStr, nil
}

//读取文章列表，不包含详情页
func (uc *UCCrawler) loadArticleList(mid, appid, ucparam string, header *http.Header) ([]*obj.ArticleObj, error) {

	urlvalue := url.Values{}
	urlvalue.Set("_app_id", appid)
	urlvalue.Set("uc_param_str", ucparam)

	urlvalue.Set("_fetch", "1")
	urlvalue.Set("_fetch_incrs", "1")
	urlvalue.Set("_size", "50")
	urlvalue.Set("_max_pos", "")
	urlvalue.Set("_select", "xss_item_id,title,read_times,publish_at,thumbnail_url,cover_url,article_category,wm_id")
	urlvalue.Set("_", strconv.Itoa(time.Now().UTC().Nanosecond()/1000))

	urlStr := "http://napi.uc.cn/3/classes/article/categories/wemedia/lists/" + mid + "?" + urlvalue.Encode()

	resBytes, err, response := uc.CoHttp.GetBytes(urlStr, header)
	if err != nil || response.StatusCode != 200 {
		return nil, NewError("读取主页文章列表接口错误", urlStr, err)
	}

	rootjson, err := simplejson.NewJson(resBytes)

	if err != nil {
		return nil, NewError("解析文章列表json失败", err, urlStr)
	}

	rootjson = rootjson.Get("data")
	dataLen := len(rootjson.MustArray())
	if dataLen <= 0 {
		return nil, NewError("文章列表json中data文章长度<=0", urlStr)
	}

	articleArray := make([]*obj.ArticleObj, 0, dataLen+1)
	for index := 0; index < dataLen; index++ {
		itemJson := rootjson.GetIndex(index)

		ucArticle := obj.ObtainArticleObj()
		ucArticle.SourceSiteTypeCode = obj.Site_UC_Code
		ucArticle.SourceSiteName = obj.Site_UC
		ucArticle.Title = itemJson.Get("title").MustString()
		ucArticle.ThumbnailsUrl = itemJson.Get("cover_url").MustString()
		ucArticle.SourceId = itemJson.Get("_id").MustString()
		ucArticle.SourcePubtimestamp = itemJson.Get("_pos").MustInt64() / 1000
		ucArticle.SourcePubtimestr = itemJson.Get("publish_at").MustString()

		wmId := itemJson.Get("wm_id").MustString()
		ucArticle.SourceWebUrl = fmt.Sprintf("http://a.mp.uc.cn/article.html?uc_param_str=%s&from=media#!wm_aid=%s!!wm_id=%s", ucparam, ucArticle.SourceId, wmId)

		articleArray = append(articleArray, ucArticle)
	}
	return articleArray, nil
}

//读取解析新闻内容
func (uc *UCCrawler) loadParseArticleDetail(sourceId, appid string, article *obj.ArticleObj) (*obj.ArticleObj,
	error) {
	contentUrl := "http://napi.uc.cn/3/classes/article/objects/" + sourceId + "?"
	urlValue := url.Values{}
	urlValue.Set("_app_id", appid)
	urlValue.Set("_fetch", "1")
	urlValue.Set("_fetch_incrs", "1")
	urlValue.Set("_max_age", "60")
	urlValue.Set("_ch", "article")
	contentUrl = contentUrl + urlValue.Encode()

	resByte, err, response := uc.CoHttp.GetBytes(contentUrl, nil)

	if err != nil {
		return article, NewError("读取文章内容错误,", err, contentUrl)
	}

	if article == nil {
		article = obj.ObtainArticleObj()
		article.SourceSiteName = obj.Site_UC
		article.SourceSiteCode = obj.Site_UC_Code
	}
	article.SourceHtml = string(resByte)

	if response.StatusCode != 200 {
		return article, NewError("读取文章内容错误,http状态码异常", response.Status, contentUrl)
	}

	return uc.parseArticleDetail(resByte, article)
}
func (uc *UCCrawler) parseArticleDetail(resByte []byte, article *obj.ArticleObj) (*obj.ArticleObj, error) {

	dataJson, err := simplejson.NewJson(resByte)
	if err != nil {
		return article, NewError("解析文章内容为json失败,", article.SourceWebUrl, string(resByte))
	}
	dataJson = dataJson.Get("data")

	if article.SourcePubtimestr == "" {
		article.SourcePubtimestr = dataJson.Get("publish_at").MustString()
	}

	if article.ThumbnailsUrl == "" {
		article.ThumbnailsUrl = dataJson.Get("cover_url").MustString()
	}

	if article.Nickname == "" {
		article.Nickname = dataJson.Get("wm_name").MustString()
	}

	if article.SourceId == "" {
		article.SourceId = dataJson.Get("_id").MustString()
	}
	//if article.UserObj ==nil {
	//	article.UserObj = obj.UserObj{}
	//}
	article.UserObj.SourceId = dataJson.Get("wm_id").MustString()
	article.UserObj.Nickname = dataJson.Get("wm_name").MustString()

	article.SourceViewCount = dataJson.GetPath("_incrs", "readtimes").MustInt()

	if otherInfoJson, ok := dataJson.CheckGet("other_info"); ok {
		article.VideoSrc = otherInfoJson.Get("video_playurl").MustString()
	}

	article.Title = dataJson.Get("title").MustString(article.Title)
	article.ContentHtml = dataJson.Get("content").MustString("0")

	if article.ContentHtml == "0" {
		return article, NewError("读取文章json的内容数据错误", article.SourceWebUrl, article.Title)
	}
	return article, nil
}

//保存新闻数据到数据库
func (uc *UCCrawler) saveArticle(article *obj.ArticleObj) error {
	var err error
	article.DBId, _, err = uc.CoDB.InsertArticleIfNotExist(article)
	if err != nil {
		return err
	}
	return nil
}

//读取保存文章内容
func (uc *UCCrawler) loadSaveArticleListDetail(appid string, articleList []*obj.ArticleObj) ([]*obj.ArticleObj) {

	resArray := make([]*obj.ArticleObj, 0, len(articleList))
	var err error
	for _, item := range articleList {
		if item.SourceId == "" {
			uc.AddLog(rockgo.Log_Warn, "文章sourceId为空无法采集", item.Title, item.SourceSiteName)
			continue
		}
		item.DBId, err = uc.CoDB.QueryExistedArticle(item)
		if err != nil || item.ContentHtml == "" || item.DBId < 0 {
			uc.AddLog(rockgo.Log_Warn, "文章不存在，准备采集", err, item.Title, item.SourceSiteName)
		}

		if item.DBId > 0 && len(item.SourceHtml) > 0 {
			item, err = uc.parseArticleDetail([]byte(item.SourceHtml), item)
			if err == nil {
				resArray = append(resArray, item)
				continue
			}
		}

		item, err = uc.loadParseArticleDetail(item.SourceId, appid, item)
		if err != nil {
			uc.AddLog(rockgo.Log_Warn, "读取解析文章内容出现错误", item.Title, item.SourceSiteName, item.SourceWebUrl, err)
			continue
		}

		err = uc.saveArticle(item)
		if err != nil {
			uc.AddLog(rockgo.Log_Warn, "保存文章数据时出现错误", item.Title, item.SourceSiteName, item.SourceWebUrl, err)
		}
		resArray = append(resArray, item)
	}

	return articleList
}
