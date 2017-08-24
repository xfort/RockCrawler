package main

import (
	"github.com/xfort/RockCrawler/crawler"
	"github.com/xfort/rockgo"
	"log"
	"encoding/json"
	"github.com/xfort/RockCrawler/publish"
	"github.com/xfort/RockCrawler/server"
	"time"
	"path/filepath"
	"os"
)

var currentDir string

func main() {
	defer func() {
		if err := recover(); err != nil {
			log.Println("程序出现严重错误，终止运行")
			time.Sleep(1 * time.Hour)
		}
	}()
	var err error
	currentDir, err = filepath.Abs(filepath.Base(os.Args[0]))
	if err != nil {
		log.Fatalln("读取当前文件路径失败，终止运行")
	}

	//go startMiaoPai()
	//go startDuoWan()
	go startUC()
	go startJinriTouTiao()
	go startQQKuaibao()

	err = server.StartHttpServer("127.0.0.1:10000")
	if err != nil {
		panic("启动grpc失败_" + err.Error())
	}
	log.Println("程序异常，终止运行")
}

func startMiaoPai() {
	mpCrawler := &crawler.MiaopaiCrawler{}
	err := mpCrawler.Init(rockgo.NewRockHttp(), nil)
	if err != nil {
		log.Fatalln(err)
	}
	mpCrawler.Start()

	for {
		resArticle, ok := mpCrawler.GetResArticle()
		if !ok {
			break
		}
		resBytes, err := json.Marshal(resArticle)
		log.Println(err, string(resBytes))
	}
}

func startDuoWan() {
	dwCrawler := &crawler.DuowanCrawler{}
	dwCrawler.TypeName = "duowan"

	rockhttp := rockgo.NewRockHttp()
	xsPublisher := &publish.XSPublish{}
	xsPublisher.HttpObj = rockhttp

	dwCrawler.XSPublish = xsPublisher

	err := dwCrawler.Init(rockhttp, nil)

	if err != nil {
		log.Fatalln(err)
	}
	err = xsPublisher.Init(rockhttp, dwCrawler.CoDB)
	if err != nil {
		log.Fatalln(err)
	}

	xsPublisher.Start()

	dwCrawler.Start()
	for {
		item, ok := dwCrawler.GetOutArticle()
		if !ok {
			break
		}
		log.Println(item.Title)
	}
}

func startUC() {
	defer func() {
		if err := recover(); err != nil {
			log.Println("uc采集器异常", err)
		}
	}()
	ucCrawler := crawler.UCCrawler{}
	ucCrawler.ConfigDirPath = currentDir
	ucCrawler.DBDirPath = filepath.Join(currentDir, "data")
	rockhttp := rockgo.NewRockHttp()
	err := ucCrawler.InitUC(rockhttp, nil)
	if err != nil {
		panic("初始化UC采集器失败_" + err.Error())
	}
	ucCrawler.Start()
}

func startJinriTouTiao() {
	defer func() {
		if err := recover(); err != nil {
			log.Println("今日头条采集器异常", err)
		}
	}()
	jrTT := crawler.JinRiTouTiaoCrawler{}
	jrTT.ConfigDirPath = currentDir
	jrTT.DBDirPath = filepath.Join(currentDir, "data")

	rockhttp := rockgo.NewRockHttp()
	err := jrTT.InitJR(rockhttp, nil)
	if err != nil {
		panic("今日头条采集器初始化失败_" + err.Error())
	}
	jrTT.Start()
}

func startQQKuaibao() {

	defer func() {
		if err := recover(); err != nil {
			log.Println("uc采集器异常", err)
		}
	}()

	qqkb := &crawler.QQKuaibaoCrawler{}
	qqkb.ConfigDirPath = currentDir
	qqkb.DBDirPath = filepath.Join(currentDir, "data")
	rockhttp := rockgo.NewRockHttp()
	err := qqkb.InitQQKB(rockhttp, nil)
	if err != nil {
		panic("天天快报采集器启动失败_" + err.Error())
	}
	qqkb.Start()

}
