package db

import (
	"database/sql"
	"bytes"
	"github.com/xfort/rockgo"
	"fmt"

	"github.com/xfort/RockCrawler/obj"
	"errors"
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
	sqlBUf := bytes.NewBufferString("CREATE TABLE IF NOT EXIST " + User_Tab + "(")
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

//添加文章,如果文章不存在就添加
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

//查询文章是否存在，dbid<0表示不存在
func (objdb *ArticleObjDB) QueryExistedArticle(article *obj.ArticleObj) (dbid int64, err error) {

	sqlstr := fmt.Sprint("SELECT "+Article_DBID, " FROM "+Article_Tab, " WHERE "+Article_SourceWebUrl+"=?;")
	res := objdb.objDB.QueryRow(sqlstr, article.SourceWebUrl)

	err = res.Scan(&dbid)
	if err != nil {
		dbid = -1
		if err == sql.ErrNoRows {
			dbid = -1
			return dbid, nil
		}
		return dbid, errors.New("查询文章是否存在失败," + sqlstr + article.SourceWebUrl + "_" + article.Title + err.Error())
	}
	return dbid, err
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
