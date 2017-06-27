package crawler

import (
	"github.com/xfort/RockCrawler/obj"
	"github.com/xfort/rockgo"
)

type DuowanCrawler struct {
	CrawlerObj
}

func (dw *DuowanCrawler) LoadHomeArtiles(task *obj.TaskObj) ([]*obj.ArticleObj, error) {

	return nil, rockgo.NewError("hello")
}
