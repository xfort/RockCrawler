package data

import "github.com/xfort/RockCrawler/protoobj"

/**
数据处理抽象
 */
type DataHandlerIn interface {
	SaveArticle(article *protoobj.ArticleObj) (*protoobj.ArticleObj, error)

	/**检查文章是否已存在，是否需要采集文章详情
	 */
	CheckArticleExisted(article *protoobj.ArticleObj) (*protoobj.ArticleObj, error)

	/**
	查询出文章所有数据
	 */
	QueryArticle(article *protoobj.ArticleObj) (*protoobj.AccountObj, error)

}
