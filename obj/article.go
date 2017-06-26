package obj

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
