package crawler

import "github.com/xfort/RockCrawler/protoobj"

/**
采集处理抽象,负责从任务中的起始页开始 采集文章，自动检查去重
 */
type CrawlerIn interface {
	LoadArticleList(taskobj *protoobj.CrawlerTaskObj) ([]*protoobj.ArticleObj, error)

	LoadArticlesDetail(articleList []*protoobj.ArticleObj) ([]*protoobj.ArticleObj, error)

	LoadArticleDetail(article *protoobj.ArticleObj) (*protoobj.ArticleObj, error)

	Cancel()
}
