package crawler

import (
	"github.com/xfort/rockgo"
	"github.com/xfort/RockCrawler/db"
	"github.com/xfort/RockCrawler/obj"
	"time"
	"log"
	"fmt"
	"errors"
	"path/filepath"
)

type CrawlerObj struct {
	CoHttp *rockgo.RockHttp
	CoDB   *db.ArticleObjDB

	sourceTaskChan chan *obj.TaskObj
	outResChan     chan *obj.ArticleObj

	TypeName      string
	DBDirPath     string
	ConfigDirPath string

	LoadArticles    LoadArticlesHandler //采集文章的具体实现
	PublishArticles func([]*obj.ArticleObj) error
}

//执行任务的具体实现
type LoadArticlesHandler func(taskObj *obj.TaskObj) ([]*obj.ArticleObj, error)

func (co *CrawlerObj) AddLog(lv int, v ...interface{}) {
	log.Println(lv, v)
}
func (co *CrawlerObj) Init(cohttp *rockgo.RockHttp, codb *db.ArticleObjDB) error {

	co.sourceTaskChan = make(chan *obj.TaskObj, 1024)
	co.outResChan = make(chan *obj.ArticleObj, 20480)
	co.CoHttp = cohttp
	if codb == nil {
		err := co.OpenDB()
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
	err := co.CoDB.OpenDB("sqlite3", filepath.Join(co.DBDirPath, co.TypeName+".db"))
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

	configPath := filepath.Join(co.ConfigDirPath, "config_"+co.TypeName+".json")

	taskArray, err := obj.ParseConfigFile(configPath, filepath.Join(co.ConfigDirPath, "publisher_config.json"))
	if err != nil {
		co.AddLog(6, "读取解析配置文件错误", err.Error(), configPath)
		time.AfterFunc(1*time.Minute, co.readConfig)
		return
	}
	tasksLen := len(taskArray)

	for _, item := range taskArray {
		co.AddTaskObj(item)
	}

	if tasksLen < 10 {
		tasksLen = 10
	} else if tasksLen > 60 {
		tasksLen = 60
	}
	//tasksLen = 1
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
		if item.CollectCode == 0 {
			co.AddLog(4, "忽略此任务，标记为不采集", item.Name, item.TaskUrl)
			continue
		}

		if co.LoadArticles == nil {
			co.AddLog(6, "执行任务错误，负责读取处理文章函数为空", item.Name, item.TaskUrl)
			continue
		}

		articleArray, err := co.LoadArticles(item)
		co.AddLog(4, "采集结束", item.Name, "采集数据", len(articleArray))

		if err != nil {
			co.AddLog(6, "执行任务错误", err, item.Name, item.TaskUrl)
		}

		//if item.PublishCode == 0 {
		//	co.AddLog(4, "不发布此任务文章，标记为不发布", item.Name, item.TaskUrl)
		//	continue
		//}

		if articleArray != nil && len(articleArray) > 0 {
			if co.PublishArticles != nil {

				err = co.PublishArticles(articleArray)
				if err != nil {
					co.AddLog(6, "发布文章出现错误", co.TypeName, err.Error())
				}
			}
			//go co.sendRes(articleArray)
		}
	}
}

func (co *CrawlerObj) StopNow() {
	//TODO
}

func NewError(v ...interface{}) error {
	return errors.New(fmt.Sprint(v))
}
