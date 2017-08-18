package obj

import "encoding/json"

const (
	Site_MiaoPai      = "MiaoPai"
	Site_MiaoPai_Code = 1

	Site_DuoWan      = "DuoWan"
	Site_DuoWan_Code = 2

	Site_QQKuaibao      = "QQKuaiBao"
	Site_QQKuaibao_Code = 3

	Site_UC      = "uc"
	Site_UC_Code = 4

	Site_JinRiHL      = "JinRiTouTiao"
	Site_JinRiHL_Code = 5
)

const (
	Status_PublishDataErr = 20
	Status_Publishing     = 21
	Status_PublishSuccess = 22
	Status_PublishFail    = 23
)

type ArticleObj struct {
	DBId int64

	SourceId           string //平台文章ID
	Title              string
	Des                string
	ThumbnailsUrl      string
	SourceHtml         string //网页源码
	ContentHtml        string //经过解析后的文章内容
	SourceWebUrl       string //原文链接
	SourcePubtimestamp int64  //发布时间UTC秒数
	SourcePubtimestr   string //发布时间日期字符串

	SourceSiteTypeCode int    //网站标识
	SourceSiteName     string //网站标识

	ThumbnailsData []string //缩略图数据，从小到大，多个尺寸

	CreateTimestr string //入库日期时间，格式

	PubStatusCode int //发布状态，发布标记
	UserObj

	MediaData map[string]string //视频等富媒体数据

	VideoSrc string //视频播放地址，可以直接播放的地址

	SourceAuthor string //原文中的来源作者，格式 来源_作者，例如多玩网站内一些文章，玉面小白狐_

	TaskObj *TaskObj

	PubDBId         int64 //发布状态数据表的dbid
	SourceViewCount int   //阅读数
}

type UserObj struct {
	DbId int64

	SourceId       string //平台用户ID
	Nickname       string
	IconUrl        string
	HomeUrl        string
	SourceSiteCode int    //网站标识
	SourceSiteName string //网站名字

	ArticleNum int //文章数
}

func ObtainArticleObj() *ArticleObj {
	return &ArticleObj{}
}

func ObtainUserObj() *UserObj {
	return &UserObj{}
}

func (article *ArticleObj) GetThumbnailsData() string {
	if article.ThumbnailsData != nil && len(article.ThumbnailsData) > 0 {
		jsonByte, err := json.Marshal(article.ThumbnailsData)
		if err != nil {
			return "error" + err.Error()
		}
		return string(jsonByte)
	}
	return ""
}

func (article *ArticleObj) GetMediaData() string {

	if article.MediaData != nil && len(article.MediaData) > 0 {
		bytesStr, err := json.Marshal(article.MediaData)
		if err != nil {
			return "error" + err.Error()
		}
		return string(bytesStr)
	}
	return ""
}
