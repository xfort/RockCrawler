package publish

import (
	"github.com/xfort/RockCrawler/obj"
	"github.com/xfort/rockgo"
	"github.com/xfort/RockCrawler/db"

	"net/http"
	"net/url"
	"time"
	"errors"
)

const (
	XS_Name = "xs"
)

type XSPublish struct {
	PublishObj
}

func (xs *XSPublish) Init(httpobj *rockgo.RockHttp, dbobj *db.ArticleObjDB) error {
	xs.handleArticle = xs.handleArticleObj
	err := dbobj.CreatePublishTab(XS_Name)
	if err != nil {
		return err
	}
	return xs.PublishObj.Init(httpobj, dbobj)
}

func (xs *XSPublish) handleArticleObj(articleObj *obj.ArticleObj, taskObj *obj.TaskObj) error {
	ok, err := xs.queryInsertArticlePublish(articleObj)
	if err != nil {
		xs.AddLog(5, "添加文章到发布数据表错误", err.Error(), articleObj.Title)
		return err
	}

	if ok {
		xs.AddLog(3, "文章已发布", articleObj.Title)
		return nil
	}

	err = xs.postArticle(articleObj)
	xs.updateArticlePubStatus(articleObj)
	if err != nil {
		return err
	}
	return nil
}

//查询文章发布状态,若不存在则添加，并对PubDBId赋值,
//bool ,true=查询找到此文
func (xs *XSPublish) queryInsertArticlePublish(articleObj *obj.ArticleObj) (bool, error) {
	status, err := xs.DBObj.QueryArticlePublishStatus(XS_Name, articleObj)
	if err != nil {
		return false, err
	}

	if status >= 20 && status < 30 {
		return true, nil
	}

	if articleObj.Title == "" || articleObj.ContentHtml == "" {
		articleObj.PubStatusCode = obj.Status_PublishDataErr
	} else {
		articleObj.PubStatusCode = obj.Status_Publishing
	}

	articleObj.PubDBId, err = xs.DBObj.InsertPublishArticle(XS_Name, articleObj)
	if err != nil {
		return false, err
	}

	if articleObj.PubStatusCode == obj.Status_PublishDataErr {
		return false, errors.New("文章数据异常_无法发送_" + articleObj.Title + articleObj.SourceWebUrl + articleObj.Nickname)
	}

	return true, nil
}

//发布文章
func (xs *XSPublish) postArticle(articleObj *obj.ArticleObj) (error) {
	publisherArray := articleObj.TaskObj.Publisers

	for _, item := range publisherArray {
		header := http.Header{}

		for key, value := range item.HeaderObj {
			header.Set(key, value)
		}

		body := url.Values{}
		for key, value := range item.BodyObj {
			body.Set(key, value)
		}

		bodySource := body.Get("source")
		if bodySource == "" {
			body.Set("source", articleObj.SourceSiteName+"_"+articleObj.Nickname)
		} else {
			if articleObj.Nickname != "" {
				body.Set("source", bodySource+"_"+articleObj.Nickname)
			}
		}

		if articleObj.SourcePubtimestr == "" {
			body.Set("pubdate", time.Now().Format("2006-01-02 15:04:05"))
		} else {
			body.Set("pubdate", articleObj.SourcePubtimestr)
		}

		body.Set("title", articleObj.Title)
		body.Set("body", articleObj.ContentHtml)
		body.Set("picname", articleObj.ThumbnailsUrl)

		_, err, response := xs.HttpObj.PostForm(item.Url, &header, body)
		if err != nil {
			xs.AddLog(5, "发布文章错误", articleObj.Title, err.Error(), articleObj.SourceWebUrl)
			articleObj.PubStatusCode = obj.Status_PublishFail
		} else if response.StatusCode != 200 {
			xs.AddLog(5, "发布文章错误,http状态码!=200", articleObj.Title, response.Status, articleObj.SourceWebUrl)
			articleObj.PubStatusCode = obj.Status_PublishFail
		} else {
			xs.AddLog(3, "发布文章成功", articleObj.Title)
			articleObj.PubStatusCode = obj.Status_PublishSuccess
		}
		err = xs.updateArticlePubStatus(articleObj)
		if err != nil {
			xs.AddLog(4, "更新文章发布状态 错误", articleObj.Title, err.Error())
		}
	}
	return nil
}

func (xs *XSPublish) updateArticlePubStatus(articleObj *obj.ArticleObj) (error) {
	return xs.DBObj.UpdateArticlePublishStatus(XS_Name, articleObj, articleObj.PubStatusCode)
}
