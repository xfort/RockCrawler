package crawler

import (
	"github.com/xfort/rockgo"
	"github.com/xfort/RockCrawler/db"
	"github.com/xfort/RockCrawler/obj"
	"os"
	"path"
	"io/ioutil"
	"time"
	"github.com/bitly/go-simplejson"
	"log"
)

type CrawlerObj struct {
	CoHttp *rockgo.RockHttp
	CoDB   *db.ArticleObjDB

	sourceTaskChan chan *obj.TaskObj
	outResChan     chan *obj.ArticleObj

	TypeName string
	execpath string

	LoadArticles LoadArticlesHandler
}

//执行任务的具体实现
type LoadArticlesHandler func(taskObj *obj.TaskObj) ([]*obj.ArticleObj, error)

func (co *CrawlerObj) AddLog(lv int, v ...interface{}) {
	log.Println(lv, v)
}
func (co *CrawlerObj) Init(cohttp *rockgo.RockHttp, codb *db.ArticleObjDB) error {

	var err error
	co.execpath, err = os.Getwd()
	if err != nil {
		return err
	}

	co.sourceTaskChan = make(chan *obj.TaskObj, 1024)
	co.outResChan = make(chan *obj.ArticleObj, 20480)
	co.CoHttp = cohttp
	if codb == nil {
		err = co.OpenDB()
		if err != nil {
			return err
		}
	} else {
		co.CoDB = codb
	}
	return nil
}

func (co *CrawlerObj) OpenDB() error {
	co.CoDB = &db.ArticleObjDB{}
	err := co.CoDB.OpenDB("sqlite3", path.Join(co.execpath, "data", co.TypeName+".db"))
	if err != nil {
		return err
	}
	err = co.CoDB.CreateTables()
	if err != nil {
		return err
	}
	return nil
}

//启动任务，自动读取配置文件
func (co *CrawlerObj) Start() {
	go co.startHandlerTask()
	co.readConfig()

}

func (co *CrawlerObj) readConfig() {
	configPath := path.Join(co.execpath, "config_"+co.TypeName+".json")
	configByte, err := ioutil.ReadFile(configPath)
	if err != nil {
		co.AddLog(rockgo.Log_Error, "读取配置文件错误", err.Error(), configPath)
		time.AfterFunc(1*time.Minute, co.readConfig)
		return
	}
	configjson, err := simplejson.NewJson(configByte)
	if err != nil {
		co.AddLog(rockgo.Log_Error, "解析配置文件为json错误", err.Error(), configPath)
		time.AfterFunc(1*time.Minute, co.readConfig)

		return
	}
	tasksJson := configjson.Get("tasks")
	tasksLen := len(tasksJson.MustArray())
	if tasksLen <= 0 {
		co.AddLog(rockgo.Log_Error, "解析配置文件json的tasks错误，长度<=0", configPath)
		time.AfterFunc(1*time.Minute, co.readConfig)

		return
	}

	for index := 0; index < tasksLen; index++ {
		itemJson := tasksJson.GetIndex(index)
		taskUrl := itemJson.Get("url").MustString()
		if taskUrl == "" {
			co.AddLog(rockgo.Log_Error, "解析配置文件task的url为空", configPath)
			continue
		}
		taskObj := &obj.TaskObj{TaskUrl: taskUrl}
		taskObj.Name = itemJson.Get("name").MustString()

		co.AddTaskObj(taskObj)
	}

	if tasksLen < 10 {
		tasksLen = 10
	} else if tasksLen > 60 {
		tasksLen = 60
	}

	time.AfterFunc(time.Duration(tasksLen)*time.Minute, co.readConfig)

}

func (co *CrawlerObj) AddTaskObj(task *obj.TaskObj) {
	co.sourceTaskChan <- task
}
func (co *CrawlerObj) GetOutArticle() (*obj.ArticleObj, bool) {
	res, ok := <-co.outResChan
	return res, ok
}

func (co *CrawlerObj) sendRes(articleArray []*obj.ArticleObj) {
	if articleArray == nil || len(articleArray) <= 0 {
		return
	}
	for _, item := range articleArray {
		co.outResChan <- item
	}
}

//开始处理任务，阻塞死循环
func (co *CrawlerObj) startHandlerTask() {
	for {
		item, ok := <-co.sourceTaskChan
		if !ok {
			break
		}
		articleArray, err := co.LoadArticles(item)
		if err != nil {
			co.AddLog(rockgo.Log_Error, "执行任务错误", err, item.Name, item.TaskUrl)
		}
		if articleArray != nil && len(articleArray) > 0 {
			go co.sendRes(articleArray)
		}
	}
}
