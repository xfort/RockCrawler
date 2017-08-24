package crawler

import (
	"testing"
	"github.com/xfort/rockgo"
	"log"
	"github.com/xfort/RockCrawler/obj"
	"time"
)

func TestUCCrawler_InitUC(t *testing.T) {

	ucCrawler := &UCCrawler{}
	ucCrawler.ConfigDirPath = "/Users/xs/work/go/code/work/src/github.com/xfort/RockCrawler"
	ucCrawler.DBDirPath = "/Users/xs/work/go/code/work/src/github.com/xfort/RockCrawler/data"
	err := ucCrawler.InitUC(rockgo.NewRockHttp(), nil)
	if err != nil {
		log.Fatalln(err)
	}
	taskobj := &obj.TaskObj{}

	taskobj.TaskUrl = "http://a.mp.uc.cn/media.html?mid=b0ae46038fe645cdb494b5ec454d494f&uc_biz_str=S:custom%7CC:iflow_ncmt&uc_param_str=frdnsnpfvecpntnwprdssskt&from=media"
	taskobj.CollectCode = 1


	resArray, err := ucCrawler.LoadUCArticles(taskobj)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(len(resArray))
	time.Sleep(1 * time.Minute)
}
