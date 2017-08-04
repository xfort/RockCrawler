package crawler

import (
	"github.com/xfort/RockCrawler/obj"
	"github.com/xfort/rockgo"
	"log"
	"net/url"
	"strings"
	"net/http"
	"github.com/PuerkitoBio/goquery"
	"bytes"
	"strconv"
	"regexp"
	"github.com/bitly/go-simplejson"
	"fmt"

	"github.com/xfort/RockCrawler/db"
	"os"
	"path"
	"time"
	"io/ioutil"
	"errors"

	_ "github.com/mattn/go-sqlite3"
)

type MiaopaiCrawler struct {
	mphttp   *rockgo.RockHttp
	MPDB     *db.ArticleObjDB
	execpath string

	sourceTaskChan chan string //任务源

	outResChan chan *obj.ArticleObj //任务结果
}

func (mp *MiaopaiCrawler) Init(rockhttp *rockgo.RockHttp, objdb *db.ArticleObjDB) error {
	var err error
	mp.execpath, err = os.Getwd()
	if err != nil {
		return err
	}
	mp.sourceTaskChan = make(chan string, 1024)
	mp.outResChan = make(chan *obj.ArticleObj, 20480)
	mp.mphttp = rockhttp
	if objdb == nil {
		err := mp.OpenDB()
		if err != nil {
			mp.AddLog(rockgo.Log_Error, "秒拍数据库初始化失败", err)
			return err
		}
	} else {
		mp.MPDB = objdb
	}
	return nil
}

func (mp *MiaopaiCrawler) OpenDB() error {

	mp.MPDB = &db.ArticleObjDB{}
	err := mp.MPDB.OpenDB("sqlite3", path.Join(mp.execpath, "data", "miaopai.db"))
	if err != nil {
		return err
	}
	err = mp.MPDB.CreateTables()
	if err != nil {
		return err
	}
	return nil
}

func (mp *MiaopaiCrawler) Start() {
	go mp.startHandlerTasks()
	mp.readConfig()

}
func (mp *MiaopaiCrawler) readConfig() {
	configByte, err := ioutil.ReadFile(path.Join(mp.execpath, "config_miaopai.json"))
	if err != nil {
		mp.AddLog(rockgo.Log_Error, "读取配置文件错误", path.Join(mp.execpath, "config_miaopai.json"))
		time.AfterFunc(1*time.Minute, mp.readConfig)
		return
	}
	configjson, err := simplejson.NewJson(configByte)
	if err != nil {
		mp.AddLog(rockgo.Log_Error, "解析配置文件为json错误", path.Join(mp.execpath, "config_miaopai.json"), err)
		time.AfterFunc(1*time.Minute, mp.readConfig)
		return
	}
	tasksJson := configjson.Get("tasks")

	array := tasksJson.MustArray()
	lenArray := len(array)
	if lenArray <= 0 {
		mp.AddLog(rockgo.Log_Error, "配置文件为json数组为0", path.Join(mp.execpath, "config_miaopai.json"))
		time.AfterFunc(1*time.Minute, mp.readConfig)
		return
	}

	for index := 0; index < lenArray; index++ {
		item := tasksJson.GetIndex(index)
		homeUrl := item.Get("home_url").MustString()
		if homeUrl == "" {
			mp.AddLog(rockgo.Log_Warn, "配置文件内home_url为空", path.Join(mp.execpath, "config_miaopai.json"))
			continue
		}
		mp.AddTaskHomeUrl(homeUrl)
	}

	if lenArray < 10 {
		lenArray = 10
	} else if lenArray > 60 {
		lenArray = 60
	}

	time.AfterFunc(time.Duration(lenArray)*time.Minute, mp.readConfig)
}

func (mp *MiaopaiCrawler) startHandlerTasks() {
	for {
		homeUrl, ok := <-mp.sourceTaskChan
		if !ok {
			break
		}
		articleArray, _, err := mp.LoadHomeArticles(homeUrl)
		if err != nil {
			mp.AddLog(rockgo.Log_Error, "解析主页错误", homeUrl, err)
		}
		go mp.sendRes(articleArray)
	}
}

func (mp *MiaopaiCrawler) sendRes(articleArray []*obj.ArticleObj) {
	if articleArray == nil || len(articleArray) <= 0 {
		return
	}
	var err error
	for _, item := range articleArray {
		if item == nil {
			continue
		}
		if item.DBId < 0 {
			item.DBId, err = mp.MPDB.QueryExistedArticle(item)
			if err != nil {
				mp.AddLog(rockgo.Log_Warn, "读取文章dbid错误", item.DBId, err.Error(), item.Title)
			}
		}
		mp.outResChan <- item
	}
	err = nil
}
func (mp *MiaopaiCrawler) AddTaskHomeUrl(homeUrl string) {
	mp.sourceTaskChan <- homeUrl
}

func (mp *MiaopaiCrawler) GetResArticle() (*obj.ArticleObj, bool) {
	res, ok := <-mp.outResChan
	return res, ok
}

func (mp *MiaopaiCrawler) AddLog(lv int, v ...interface{}) {
	log.Println(lv, v)
}

//读取主页所有文章数据
func (mp *MiaopaiCrawler) LoadHomeArticles(urlstr string) ([]*obj.ArticleObj, *obj.UserObj, error) {
	suid, user, err := mp.LoadHomeData(urlstr, nil)
	if err != nil || suid == "" {
		return nil, user, err
	}

	articleArray, err := mp.loadArticles(suid, user.ArticleNum)
	if err != nil {
		return articleArray, user, err
	}

	arrayLen := len(articleArray)
	//var exists bool
	if arrayLen <= 0 {
		mp.AddLog(rockgo.Log_Error, "文章列表为空", urlstr)
		return articleArray, user, rockgo.NewError("文章列表为空", urlstr)
	}

	for index := arrayLen - 1; index >= 0; index-- {
		item := articleArray[index]
		item.UserObj = *user

		item.DBId, err = mp.MPDB.QueryExistedArticle(item)
		if err != nil {
			mp.AddLog(rockgo.Log_Error, "查询文章是否存在错误", item.Title, item.Nickname, err.Error())
			err = nil
		}
		if item.DBId > 0 {
			//数据库存在此文
			mp.AddLog(rockgo.Log_Info, "数据库已存在此文", item.Title, item.Nickname)
			continue
		}
		mp.AddLog(rockgo.Log_Info, "数据库无此文", item.Title, item.Nickname, item.DBId)

		//err = mp.loadArticleDetail(item)
		//if err != nil {
		//	mp.AddLog(rockgo.Log_Error, "读取文章详情错误", item.Title, item.Nickname, err.Error())
		//	err = nil
		//}
		item.DBId, err = mp.MPDB.InsertArticlce(item)
		if err != nil {
			mp.AddLog(rockgo.Log_Error, "添加文章到数据库错误", item.Title, item.Nickname, err.Error())
			err = nil
		}

		//item.DBId, exists, err = mp.MPDB.InsertArticleIfNotExist(item)
		//if err != nil {
		//	mp.AddLog(rockgo.Log_Error, "添加文章到数据库错误", item.Title, item.Nickname, err)
		//} else {
		//	if exists {
		//		mp.AddLog(rockgo.Log_Info, "数据库已存在此文", item.Title, item.Nickname)
		//	} else {
		//		mp.AddLog(rockgo.Log_Info, "数据库无此文", item.Title, item.Nickname, item.DBId)
		//	}
		//}
	}
	return articleArray, user, err
}

func (mp *MiaopaiCrawler) LoadHomeData(urlstr string, user *obj.UserObj) (suid string, userobj *obj.UserObj, err error) {
	urlObj, err := url.Parse(urlstr)
	if err != nil {
		return "", user, err
	}

	if strings.EqualFold(urlObj.Host, "www.miaopai.com") {
		urlObj.Host = "m.miaopai.com"
	}

	header := http.Header{}
	header.Set("User-Agent", rockgo.UserAgent_Android)
	resBytes, err, response := mp.mphttp.GetBytes(urlObj.String(), &header)
	if err != nil {
		return "", user, err
	}

	if response.StatusCode != 200 {
		return "", user, rockgo.NewError("读取主页网页错误")
	}

	htmlDoc, err := goquery.NewDocumentFromReader(bytes.NewReader(resBytes))
	if err != nil {
		return "", user, err
	}

	if user == nil {
		user = obj.ObtainUserObj()
		user = obj.ObtainUserObj()
	}

	user.HomeUrl = urlstr

	user.Nickname = htmlDoc.Find("b.nick").First().Text()
	user.Nickname = strings.TrimSpace(user.Nickname)
	if user.Nickname == "" {
		mp.AddLog(rockgo.Log_Warn, "解析主页的nick错误", urlstr)
	}

	videoNum := htmlDoc.Find("div.box_count").First().Children().First().Text()
	videoNum = strings.TrimSpace(videoNum)
	if videoNum == "" {
		mp.AddLog(rockgo.Log_Warn, "解析主页的视频数据错误", urlstr, videoNum)
	} else {
		videoNum = strings.Replace(videoNum, `,`, "", -1)
		user.ArticleNum, err = strconv.Atoi(videoNum)
		if err != nil {
			mp.AddLog(rockgo.Log_Warn, "解析主页的视频数错误", videoNum, urlstr, err.Error())
			err = nil
		}
	}

	user.IconUrl, _ = htmlDoc.Find("div.head.WscaleH").First().Attr("data-url")

	suidArray := regexp.MustCompile(`var suid = '(\w+?)';`).FindSubmatch(resBytes)

	if len(suidArray) > 1 {
		user.SourceId = string(suidArray[1])
	}

	if user.SourceId == "" {
		mp.AddLog(rockgo.Log_Error, "解析主页的suid失败", user.Nickname, urlstr)
		err = rockgo.NewError("解析主页的suid失败", user.Nickname, urlstr)
	}
	return user.SourceId, user, err
}

//主页文章列表
func (mp *MiaopaiCrawler) loadArticles(suid string, articlenum int) ([]*obj.ArticleObj, error) {
	if articlenum <= 5 || articlenum > 100 {
		articlenum = 100
	}

	urlValue := url.Values{}
	urlValue.Set("suid", suid)
	urlValue.Set("page", "1")
	urlValue.Set("per", strconv.Itoa(articlenum))

	urlstr := "http://m.miaopai.com/show/getOwnerVideo?" + urlValue.Encode()

	header := &http.Header{}
	header.Set("user-agent", rockgo.UserAgent_Android)
	resBytes, err, response := mp.mphttp.GetBytes(urlstr, header)
	if err != nil {
		err = rockgo.NewError("读取主页文章列表数据失败", urlstr, err)
		mp.AddLog(rockgo.Log_Error, err.Error())
		return nil, err
	}
	if response.StatusCode != 200 {
		err = rockgo.NewError("读取主页文章列表数据失败，http结果码", urlstr, response.Status)
		mp.AddLog(rockgo.Log_Error, err.Error())
		return nil, err
	}

	dataJson, err := simplejson.NewJson(resBytes)
	if err != nil {
		err = rockgo.NewError("解析文章列表数据为json失败", urlstr, string(resBytes), err)
		mp.AddLog(rockgo.Log_Error, err.Error())
		return nil, err
	}
	htmlStr, err := dataJson.Get("msg").String()
	if err != nil {
		err = rockgo.NewError("解析文章列表数据json内的msg失败", urlstr, string(resBytes), err)
		mp.AddLog(rockgo.Log_Error, err.Error())
		return nil, err
	}

	htmldoc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlStr))
	if err != nil {
		err = rockgo.NewError("解析文章列表为doc失败", urlstr, err, string(resBytes))
		mp.AddLog(rockgo.Log_Error, err.Error())
		return nil, err
	}

	divArray := htmldoc.Find("div.card_wrapping")
	divLen := divArray.Length()
	if divLen <= 0 {
		err = rockgo.NewError("解析文章列表doc内文章数据为0,div.card_wrapping", urlstr, divLen)
		mp.AddLog(rockgo.Log_Error, err.Error())
		return nil, err
	}
	articleArray := make([]*obj.ArticleObj, 0, divLen+1)
	for index := 0; index < divLen; index++ {
		itemDiv := divArray.Eq(index)

		itemArticle, err := mp.parseItemArticleHtml(itemDiv, nil)
		if err != nil {
			mp.AddLog(rockgo.Log_Warn, "解析文章列表内的单个文章错误", err, urlstr)
		}
		articleArray = append(articleArray, itemArticle)
	}

	return articleArray, nil
}

//解析文章列表html内单个文章div
func (mp *MiaopaiCrawler) parseItemArticleHtml(itemdiv *goquery.Selection, article *obj.ArticleObj) (*obj.ArticleObj, error) {

	var errmsg string
	childrenNode := itemdiv.Children()

	if article == nil {
		article = obj.ObtainArticleObj()
		article.SourceSiteTypeCode = obj.Site_MiaoPai_Code
		article.SourceSiteName = obj.Site_MiaoPai
	}

	href := childrenNode.Eq(0).AttrOr("href", "")
	if href == "" {
		errmsg = fmt.Sprintln("解析单个文章div内href错误", itemdiv.Text())
	} else {
		article.SourceId = strings.Replace(href, `/show/channel/`, "", 1)
		article.SourceWebUrl = "http://m.miaopai.com" + href
		article.VideoSrc = "http://gslb.miaopai.com/stream/" + article.SourceId + ".mp4"
	}

	article.ThumbnailsUrl = childrenNode.Eq(0).Children().Eq(0).AttrOr("data-url", "")
	article.Title = childrenNode.Eq(1).Text()
	article.Title = strings.TrimSpace(article.Title)
	if errmsg != "" {
		return article, errors.New(errmsg + "_" + article.Title)
	}
	return article, nil
}

//读取文章详情页，读取日期时间
func (mp *MiaopaiCrawler) loadArticleDetail(article *obj.ArticleObj) error {
	if article.SourceWebUrl == "" {
		return rockgo.NewError("读取文章详情时，sourceWebUrl为空", article.Title)
	}
	header := http.Header{}
	header.Set("user-agent", rockgo.UserAgent_Android)

	resByte, err, response := mp.mphttp.GetBytes(article.SourceWebUrl, &header)
	if err != nil {
		return rockgo.NewError("读取文章详情时错误", err.Error(), article.SourceWebUrl, article.Title)
	}
	if response.StatusCode != 200 {
		return rockgo.NewError("读取文章详情时错误", response.Status, article.SourceWebUrl, article.Title)
	}

	htmlDoc, err := goquery.NewDocumentFromReader(bytes.NewReader(resByte))
	if err != nil {
		return rockgo.NewError("文章详情转为htmldoc错误", err.Error(), article.SourceWebUrl, article.Title)
	}
	htmlNode := htmlDoc.Find("div.left").First().Children()

	if htmlNode.Length() >= 2 {
		timeStr := htmlNode.Eq(1).Children().First().Text()
		article.SourcePubtimestr = timeStr
	}

	return nil
}
