package db

import (
	"database/sql"
	"bytes"
	"github.com/xfort/rockgo"
	"fmt"

	"github.com/xfort/RockCrawler/obj"
	"errors"
	"github.com/xfort/RockCrawler/proto"
	"golang.org/x/net/context"
	"log"
	"time"
)

const (
	Article_Tab = "article"

	Article_DBID     = "dbid"
	Article_SourceID = "source_id"
	Article_Title    = "title"

	User_SourceId = "user_sourceid"
	User_Nickname = "nickname"

	Article_ThumbnailsUrl      = "thumbnails"
	Article_SourceHtml         = "source_html"
	Article_ContentHtml        = "content_html"
	Article_Des                = "des"
	Article_SourceWebUrl       = "source_web_url"
	Article_SourcePubtimestamp = "source_pub_timestamp"
	Article_SourcePubtimestr   = "source_pub_timestr"
	Article_SourceAuthor       = "source_author"
	Article_SourceSiteTypeCode = "source_site_typecode"
	Article_SourceSiteName     = "source_site_name"
	Article_ThumbnailsData     = "thumbnails_data"
	Article_CreateTimestr      = "create_timestamp"
	Article_PubStatusCode      = "pub_status"

	User_Tab  = "user"
	User_DBID = "user_dbid"

	User_IconUrl        = "icon_url"
	User_HomeUrl        = "home_url"
	User_SourceSiteCode = "user_source_site_code"
	User_SourceSiteName = "user_source_site_name"
	User_ArticleNum     = "article_num"

	Article_MediaData = "media_data"
	Article_VideoSrc  = "video_src"
)

const (
	Publish_Tab_Suffix          = "_publish"
	Publish_DBId                = "id"
	Publish_ArticleDBId         = "article_dbid"
	Publish_ArticleSourceId     = "source_id"
	Publish_ArticleTitle        = "title"
	Publish_ArticleSourceWebUrl = "source_web_url"
	Publish_Status              = "status"
	Publish_CreateTime          = "create_time"
	Publish_LastUpdateTime      = "last_update_time"
)

type ArticleObjDB struct {
	objDB    *sql.DB
	dataname string
}

func (objdb *ArticleObjDB) OpenDB(driverName, dataName string) error {
	var err error
	objdb.objDB, err = sql.Open(driverName, dataName)
	if err != nil {
		return err
	}
	objdb.dataname = dataName

	return nil
}

func (objdb *ArticleObjDB) CreateTables() error {

	err := objdb.createArticleTab()
	if err != nil {
		return err
	}
	//err = objdb.createUserTab()

	return err
}

//文章数据表
func (objdb *ArticleObjDB) createArticleTab() error {
	sqlBuf := bytes.NewBufferString("CREATE TABLE IF NOT EXISTS " + Article_Tab + "(")
	sqlBuf.WriteString(Article_DBID + " INTEGER PRIMARY KEY AUTOINCREMENT,")
	sqlBuf.WriteString(Article_SourceID + " NCHAR(128),")
	sqlBuf.WriteString(Article_Title + " NCHAR(128),")
	sqlBuf.WriteString(User_SourceId + " NCHAR(128),")
	sqlBuf.WriteString(User_Nickname + " NVARCHAR(128),")
	sqlBuf.WriteString(Article_ThumbnailsUrl + " NVARCHAR(1024),")
	sqlBuf.WriteString(Article_SourceHtml + " NTEXT,")
	sqlBuf.WriteString(Article_ContentHtml + " NTEXT,")
	sqlBuf.WriteString(Article_Des + " NVARCHAR(512),")
	sqlBuf.WriteString(Article_SourceWebUrl + " NVARCHAR(1024),")
	sqlBuf.WriteString(Article_SourcePubtimestamp + " INTEGER,")
	sqlBuf.WriteString(Article_SourcePubtimestr + " DATETIME,")
	sqlBuf.WriteString(Article_SourceAuthor + " NCHAR(128),")
	sqlBuf.WriteString(Article_SourceSiteTypeCode + " INT,")
	sqlBuf.WriteString(Article_SourceSiteName + " NCHAR(64),")
	sqlBuf.WriteString(Article_ThumbnailsData + " NTEXT,")
	sqlBuf.WriteString(Article_CreateTimestr + " DATETIME DEFAULT (datetime('now','localtime')),")
	sqlBuf.WriteString(Article_PubStatusCode + " INT,")
	sqlBuf.WriteString(Article_MediaData + " NTEXT,")

	sqlBuf.WriteString(Article_VideoSrc + " NVARCHAR(1024)")
	//sqlBuf.WriteString(",")
	//sqlBuf.WriteString(",")
	sqlBuf.WriteString(");")

	_, err := objdb.objDB.Exec(sqlBuf.String())
	if err != nil {
		err = rockgo.NewError("创建文章数据表失败,", objdb.dataname, err.Error(), sqlBuf.String())
	}
	sqlBuf.Reset()
	return err
}

//用户数据表
func (objdb *ArticleObjDB) createUserTab() error {
	sqlBUf := bytes.NewBufferString("CREATE TABLE IF NOT EXISTS " + User_Tab + "(")
	sqlBUf.WriteString(User_DBID + " INTEGER PRIMARY KEY AUTOINCREMENT,")
	sqlBUf.WriteString(User_SourceId + " NCHAR(128) UNIQUE,")
	sqlBUf.WriteString(User_Nickname + " NCHAR(128),")
	sqlBUf.WriteString(User_IconUrl + " NVARCHAR(512),")
	sqlBUf.WriteString(User_ArticleNum + " INT,")
	sqlBUf.WriteString(User_HomeUrl + " NVARCHAR(1024),")
	sqlBUf.WriteString(User_SourceSiteCode + " INT,")
	sqlBUf.WriteString(User_SourceSiteName + " NCHAR(64),")
	sqlBUf.WriteString(");")
	_, err := objdb.objDB.Exec(sqlBUf.String())
	if err != nil {
		err = rockgo.NewError("创建用户数据表失败,", objdb.dataname, err.Error(), sqlBUf.String())
	}
	sqlBUf.Reset()
	return err
}

//添加文章,如果文章不存在就添加,根据文章的SourceWebUrl判断文章是否相同
//dbid,false=不存在
func (objdb *ArticleObjDB) InsertArticleIfNotExist(article *obj.ArticleObj) (int64, bool, error) {

	sqlStr := "INSERT INTO " + Article_Tab + "("
	sqlStr = sqlStr + Article_SourceID + ","
	sqlStr = sqlStr + Article_Title + ","
	sqlStr = sqlStr + User_SourceId + ","
	sqlStr = sqlStr + User_Nickname + ","
	sqlStr = sqlStr + Article_ThumbnailsUrl + ","
	sqlStr = sqlStr + Article_SourceHtml + ","
	sqlStr = sqlStr + Article_ContentHtml + ","
	sqlStr = sqlStr + Article_Des + ","
	sqlStr = sqlStr + Article_SourceWebUrl + ","
	sqlStr = sqlStr + Article_SourcePubtimestamp + ","
	sqlStr = sqlStr + Article_SourcePubtimestr + ","
	sqlStr = sqlStr + Article_SourceAuthor + ","
	sqlStr = sqlStr + Article_SourceSiteTypeCode + ","
	sqlStr = sqlStr + Article_SourceSiteName + ","
	sqlStr = sqlStr + Article_ThumbnailsData + ","
	sqlStr = sqlStr + Article_PubStatusCode + ","
	sqlStr = sqlStr + Article_MediaData + ","
	sqlStr = sqlStr + Article_VideoSrc
	sqlStr = sqlStr + ") SELECT ?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,? "
	sqlStr = sqlStr + " WHERE NOT EXISTS(  "
	sqlStr = sqlStr + " SELECT " + Article_DBID + " FROM " + Article_Tab + " WHERE " + Article_SourceWebUrl + "=?"
	sqlStr = sqlStr + ")"

	res, err := objdb.objDB.Exec(sqlStr, article.SourceId, article.Title, article.UserObj.SourceId, article.UserObj.Nickname, article.ThumbnailsUrl, article.SourceHtml, article.ContentHtml, article.Des, article.SourceWebUrl, article.SourcePubtimestamp, article.SourcePubtimestr, article.SourceAuthor, article.SourceSiteTypeCode, article.SourceSiteName, article.GetThumbnailsData(), article.PubStatusCode, article.GetMediaData(), article.VideoSrc, article.SourceWebUrl)
	if err != nil {
		return -1, false, errors.New(err.Error() + sqlStr)
	}
	dbid, err := res.LastInsertId()
	if err != nil {
		return -1, false, errors.New(err.Error() + sqlStr)
	} else if dbid > 0 {
		return dbid, false, nil
	}
	num, err := res.RowsAffected()
	if err != nil {
		return -1, false, errors.New(err.Error() + sqlStr)
	}
	if num == 0 { //已存在此文
		return -1, true, nil
	}
	return -1, false, nil
}

//添加文章,如果文章不存在就添加,根据文章的SourceWebUrl判断文章是否相同
//dbid,false=不存在
func (objdb *ArticleObjDB) InsertArticleIfNotExistBySourceId(article *obj.ArticleObj) (int64, bool, error) {

	sqlStr := "INSERT INTO " + Article_Tab + "("
	sqlStr = sqlStr + Article_SourceID + ","
	sqlStr = sqlStr + Article_Title + ","
	sqlStr = sqlStr + User_SourceId + ","
	sqlStr = sqlStr + User_Nickname + ","
	sqlStr = sqlStr + Article_ThumbnailsUrl + ","
	sqlStr = sqlStr + Article_SourceHtml + ","
	sqlStr = sqlStr + Article_ContentHtml + ","
	sqlStr = sqlStr + Article_Des + ","
	sqlStr = sqlStr + Article_SourceWebUrl + ","
	sqlStr = sqlStr + Article_SourcePubtimestamp + ","
	sqlStr = sqlStr + Article_SourcePubtimestr + ","
	sqlStr = sqlStr + Article_SourceAuthor + ","
	sqlStr = sqlStr + Article_SourceSiteTypeCode + ","
	sqlStr = sqlStr + Article_SourceSiteName + ","
	sqlStr = sqlStr + Article_ThumbnailsData + ","
	sqlStr = sqlStr + Article_PubStatusCode + ","
	sqlStr = sqlStr + Article_MediaData + ","
	sqlStr = sqlStr + Article_VideoSrc
	sqlStr = sqlStr + ") SELECT ?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,? "
	sqlStr = sqlStr + " WHERE NOT EXISTS(  "
	sqlStr = sqlStr + " SELECT " + Article_DBID + " FROM " + Article_Tab + " WHERE " + Article_SourceID + "=?"
	sqlStr = sqlStr + ")"

	res, err := objdb.objDB.Exec(sqlStr, article.SourceId, article.Title, article.UserObj.SourceId, article.UserObj.Nickname, article.ThumbnailsUrl, article.SourceHtml, article.ContentHtml, article.Des, article.SourceWebUrl, article.SourcePubtimestamp, article.SourcePubtimestr, article.SourceAuthor, article.SourceSiteTypeCode, article.SourceSiteName, article.GetThumbnailsData(), article.PubStatusCode, article.GetMediaData(), article.VideoSrc, article.SourceId)
	if err != nil {
		return -1, false, errors.New(err.Error() + sqlStr)
	}
	dbid, err := res.LastInsertId()
	if err != nil {
		return -1, false, errors.New(err.Error() + sqlStr)
	} else if dbid > 0 {
		return dbid, false, nil
	}
	num, err := res.RowsAffected()
	if err != nil {
		return -1, false, errors.New(err.Error() + sqlStr)
	}
	if num == 0 { //已存在此文
		return -1, true, nil
	}
	return -1, false, nil
}

//查询文章是否存在，dbid<0表示不存在,根据文章的url地址
func (objdb *ArticleObjDB) QueryExistedArticle(article *obj.ArticleObj) (dbid int64, err error) {

	sqlstr := fmt.Sprint("SELECT "+Article_DBID+","+Article_ContentHtml+","+Article_SourceHtml, " FROM "+Article_Tab, " WHERE "+Article_SourceWebUrl+"=?;")
	res := objdb.objDB.QueryRow(sqlstr, article.SourceWebUrl)
	var dbId int64
	var contentHtml string
	var sourceHtml string
	err = res.Scan(&dbId, &contentHtml, &sourceHtml)

	if err != nil {
		dbId = -1
		if err == sql.ErrNoRows {
			return dbId, nil
		}
		return dbId, errors.New("查询文章是否存在失败," + sqlstr + article.SourceWebUrl + "_" + article.Title + err.Error())
	}
	article.DBId = dbId
	article.ContentHtml = contentHtml
	article.SourceHtml = sourceHtml
	return dbId, nil
}

//查询文章是否存在，根据SourceId
//dbid<0表示不存在
func (objdb *ArticleObjDB) QueryArticleBySourceId(ctx context.Context, sourceId string, article *obj.ArticleObj) (dbid int64, err error) {
	sqlstr := fmt.Sprint("SELECT "+Article_DBID+","+Article_ContentHtml+","+Article_SourceHtml, " FROM "+Article_Tab, " WHERE "+Article_SourceID+"=?;")
	res := objdb.objDB.QueryRow(sqlstr, sourceId)
	var dbId int64
	var contentHtml string
	var sourceHtml string
	err = res.Scan(&dbId, &contentHtml, &sourceHtml)

	if err != nil {
		dbId = -1
		if err == sql.ErrNoRows {
			return dbId, nil
		}
		return dbId, errors.New("通过sourceID查询文章失败," + sqlstr + article.SourceWebUrl + "_" + article.Title + err.Error())
	}
	//article.DBId = dbId
	article.ContentHtml = contentHtml
	article.SourceHtml = sourceHtml
	return dbId, nil
}

func (objdb *ArticleObjDB) InsertArticlce(article *obj.ArticleObj) (int64, error) {
	sqlStr := "INSERT INTO " + Article_Tab + "("
	sqlStr = sqlStr + Article_SourceID + ","
	sqlStr = sqlStr + Article_Title + ","
	sqlStr = sqlStr + User_SourceId + ","
	sqlStr = sqlStr + User_Nickname + ","
	sqlStr = sqlStr + Article_ThumbnailsUrl + ","
	sqlStr = sqlStr + Article_SourceHtml + ","
	sqlStr = sqlStr + Article_ContentHtml + ","
	sqlStr = sqlStr + Article_Des + ","
	sqlStr = sqlStr + Article_SourceWebUrl + ","
	sqlStr = sqlStr + Article_SourcePubtimestamp + ","
	sqlStr = sqlStr + Article_SourcePubtimestr + ","
	sqlStr = sqlStr + Article_SourceAuthor + ","
	sqlStr = sqlStr + Article_SourceSiteTypeCode + ","
	sqlStr = sqlStr + Article_SourceSiteName + ","
	sqlStr = sqlStr + Article_ThumbnailsData + ","
	sqlStr = sqlStr + Article_PubStatusCode + ","
	sqlStr = sqlStr + Article_MediaData + ","
	sqlStr = sqlStr + Article_VideoSrc
	sqlStr = sqlStr + ") VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)"

	res, err := objdb.objDB.Exec(sqlStr, article.SourceId, article.Title, article.UserObj.SourceId, article.UserObj.Nickname, article.ThumbnailsUrl, article.SourceHtml, article.ContentHtml, article.Des, article.SourceWebUrl, article.SourcePubtimestamp, article.SourcePubtimestr, article.SourceAuthor, article.SourceSiteTypeCode, article.SourceSiteName, article.GetThumbnailsData(), article.PubStatusCode, article.GetMediaData(), article.VideoSrc)
	if err != nil {
		return 0, errors.New(err.Error() + sqlStr)
	}
	return res.LastInsertId()
}

//创建 发布记录数据表
func (objdb *ArticleObjDB) CreatePublishTab(tabPre string) error {
	sqlBUf := bytes.NewBufferString("CREATE TABLE IF NOT EXISTS " + tabPre + Publish_Tab_Suffix + "(")
	sqlBUf.WriteString(Publish_DBId + " INTEGER PRIMARY KEY AUTOINCREMENT,")
	sqlBUf.WriteString(Publish_ArticleDBId + " INT,")
	sqlBUf.WriteString(Publish_ArticleTitle + " NCHAR(128),")
	sqlBUf.WriteString(Publish_ArticleSourceId + " NCHAR(128),")
	sqlBUf.WriteString(Publish_ArticleSourceWebUrl + " NVARCHAR(1024),")
	sqlBUf.WriteString(Publish_Status + " INT,")
	sqlBUf.WriteString(Publish_CreateTime + " DATETIME DEFAULT (datetime('now','localtime')),")
	sqlBUf.WriteString(Publish_LastUpdateTime + " DATETIME")
	sqlBUf.WriteString(");")

	_, err := objdb.objDB.Exec(sqlBUf.String())
	if err != nil {
		return rockgo.NewError("创建Publish数据表失败", err.Error(), objdb.dataname, sqlBUf.String())
	}
	return nil
}

//查询文章发布状态,根据Article.Title
func (objdb *ArticleObjDB) QueryArticlePublishStatus(tabPre string, article *obj.ArticleObj) (status int, err error) {
	if article.Title == "" {
		return 0, rockgo.NewError("文章title为空", article.SourceWebUrl, article.DBId)
	}
	sqlStr := "SELECT " + Publish_DBId + "," + Publish_Status
	sqlStr = sqlStr + " FROM " + tabPre + Publish_Tab_Suffix
	sqlStr = sqlStr + " WHERE " + Publish_ArticleTitle + "=?;"

	res := objdb.objDB.QueryRow(sqlStr, article.Title)
	var pubId int64
	var statusCode int

	err = res.Scan(&pubId, &statusCode)
	article.PubDBId = pubId
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, errors.New("查询文章发布状态错误_" + err.Error() + "_" + sqlStr)
	}
	//log.Println("文章状态查询", sqlStr)
	return statusCode, nil
}

//查询文章发布状态,根据title
func (objdb *ArticleObjDB) QueryArticlePublishStatusByTitle(tabPre string, title string) (pubID int64, status int, err error) {

	sqlStr := "SELECT " + Publish_DBId + "," + Publish_Status
	sqlStr = sqlStr + " FROM " + tabPre + Publish_Tab_Suffix
	sqlStr = sqlStr + " WHERE " + Publish_ArticleTitle + "=?;"

	res := objdb.objDB.QueryRow(sqlStr, title)
	var pubId int64
	var statusCode int

	err = res.Scan(&pubId, &statusCode)

	if err != nil {
		if err == sql.ErrNoRows {
			return pubId, 0, nil
		}
		return pubId, 0, errors.New("查询文章发布状态错误_" + err.Error() + "_" + sqlStr)
	}
	//log.Println("文章状态查询", sqlStr)
	return pubId, statusCode, nil
}

//添加文章发布状态
func (objdb *ArticleObjDB) InsertPublishArticle(tabPre string, article *obj.ArticleObj) (int64, error) {
	sqlStr := "INSERT INTO " + tabPre + Publish_Tab_Suffix + "("
	sqlStr = sqlStr + Publish_ArticleDBId + ","
	sqlStr = sqlStr + Publish_ArticleTitle + ","
	sqlStr = sqlStr + Publish_ArticleSourceId + ","
	sqlStr = sqlStr + Publish_ArticleSourceWebUrl + ","
	sqlStr = sqlStr + Publish_Status + ")"
	sqlStr = sqlStr + " VALUES(?,?,?,?,?);"

	res, err := objdb.objDB.Exec(sqlStr, article.DBId, article.Title, article.SourceId, article.SourceWebUrl, article.PubStatusCode)
	if err != nil {
		return 0, errors.New("新增文章发布数据失败_" + err.Error() + "_" + sqlStr + "_" + article.Title)
	}

	article.PubDBId, err = res.LastInsertId()
	if err != nil {
		return 0, errors.New("新增文章发布数据失败_" + err.Error() + "_" + sqlStr + "_" + article.Title)
	}

	return article.PubDBId, nil
}

//更新文章发布状态码,根据article.pubdbid
func (objdb *ArticleObjDB) UpdateArticlePublishStatus(tabPre string, articleObj *obj.ArticleObj, status int) error {

	if articleObj.PubDBId <= 0 {
		return errors.New("文章PubDBID<=0," + articleObj.Title)
	}

	sqlStr := "UPDATE " + tabPre + Publish_Tab_Suffix + " SET "
	sqlStr = sqlStr + " " + Publish_Status + "=?"
	sqlStr = sqlStr + " WHERE " + Publish_DBId + "=?"

	_, err := objdb.objDB.Exec(sqlStr, status, articleObj.PubDBId)
	if err != nil {
		return errors.New(err.Error() + "_" + articleObj.Title)
	}
	return nil
}

//查询作者文章，根据用户sourceId，或者accountnum
func (objdb *ArticleObjDB) QueryArticlesByUser(ctx context.Context, userSourceId string, accountNum string) ([]*proto.ArticleObj, *proto.UserObj, error) {

	sqlstr := fmt.Sprintf("SELECT %s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s ", Article_DBID, Article_SourceID, Article_Title, Article_ContentHtml, Article_ThumbnailsUrl, Article_SourcePubtimestamp, Article_SourcePubtimestr, User_SourceId, User_Nickname, Article_SourceSiteName, Article_SourceWebUrl)
	sqlstr = sqlstr + " FROM " + Article_Tab
	sqlstr = sqlstr + " WHERE " + fmt.Sprintf("%s=? OR %s=?  order by %s desc LIMIT 100;", User_SourceId, User_SourceId, Article_DBID)

	rows, err := objdb.objDB.QueryContext(ctx, sqlstr, userSourceId, accountNum)
	if err != nil {
		if rows != nil {
			rows.Close()
		}
		return nil, nil, errors.New("查询用户的最近100篇文章失败_" + sqlstr + "_" + err.Error() + "——" + objdb.dataname)
	}
	defer rows.Close()

	user := proto.UserObj{}
	var thumbnailsUrl string
	var siteName string

	resList := make([]*proto.ArticleObj, 0, 101)

	for rows.Next() {
		article := proto.ArticleObj{}

		err := rows.Scan(&article.XsId, &article.SourceId, &article.Title, &article.ContentHtml, &thumbnailsUrl, &article.SourcePublishTimeUTCSec, &article.SourcePublishTimeStr, &user.SourceId, &user.Nickname, &siteName, &article.SourceWebUrl)
		if err != nil {
			log.Println("读取作者的100文章出错", err, sqlstr, userSourceId, )
			continue
		}
		article.ThumbnailsUrl = []string{thumbnailsUrl}
		article.SourceSiteName = siteName
		user.SourceSitename = siteName
		resList = append(resList, &article)
	}
	return resList, &user, nil
}
func (articleDB *ArticleObjDB) Close() {
	if articleDB.objDB != nil {
		err := articleDB.objDB.Close()
		if err != nil {
			log.Println("关闭数据库失败", articleDB.dataname)
		}
	}
}

//查询账号的发布历史记录
func (articleDB *ArticleObjDB) LoadAccountPublishArticles(accountName string) ([]*proto.ArticleObj, error) {

	sqlStr := fmt.Sprintf("SELECT %s,%s,%s,%s From %s ORDER BY %s DESC LIMIT 60", Publish_DBId, Publish_ArticleTitle, Publish_ArticleSourceWebUrl, Publish_CreateTime, accountName+Publish_Tab_Suffix, Publish_DBId)

	resRows, err := articleDB.objDB.Query(sqlStr)

	if err != nil {
		resRows.Close()
		return nil, err
	}
	defer resRows.Close()

	resArticles := make([]*proto.ArticleObj, 0, 60)

	for resRows.Next() {
		article := &proto.ArticleObj{}
		var dateTime time.Time
		err := resRows.Scan(&article.XsId, &article.Title, &article.SourceWebUrl, &dateTime)
		if err != nil {
			log.Println("读取文章发布数据错误", err, sqlStr)
			continue
		}
		article.SourcePublishTimeUTCSec = dateTime.UTC().Unix()
		resArticles = append(resArticles, article)
	}
	return resArticles, nil
}

///删除已发布文章记录
func (articleDB *ArticleObjDB) DeletePublishedArticle(accountname string, articleId int64) error {

	sqlStr := fmt.Sprintf("Delete From %s Where %s=?;", accountname+Publish_Tab_Suffix, Publish_DBId)
	_, err := articleDB.objDB.Exec(sqlStr, articleId)
	return err
}
