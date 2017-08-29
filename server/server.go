package server

import (
	"net"
	"google.golang.org/grpc"
	"github.com/xfort/RockCrawler/proto"

	"golang.org/x/net/context"
	"log"
	"github.com/xfort/RockCrawler/db"
	"path/filepath"
	"os"
)

//启动本地http服务
func StartHttpServer(addr string) error {
	log.Println("准备在", addr, "启动gRPC服务")
	netlis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	grpcServer := grpc.NewServer()
	proto.RegisterRockCrawlerServer(grpcServer, &CrawlerServer{})

	return grpcServer.Serve(netlis)
}

type CrawlerServer struct {
	articleDBMap map[string]*db.ArticleObjDB
}

//读取作者的所有文章
func (cs *CrawlerServer) LoadArticlesByUser(ctx context.Context, useract *proto.UserAct) (*proto.ResultArticlesList, error) {
	resData := &proto.ResultArticlesList{Code: 1}
	if useract.User == nil || useract.User.SourceSitename == "" || useract.User.SourceId == "" {
		resData.Errmsg = "SourceSitename，SourceId不能为空"
		return resData, nil
	}
	if cs.articleDBMap == nil {
		cs.articleDBMap = make(map[string]*db.ArticleObjDB)
	}
	articleDB := cs.articleDBMap[useract.User.SourceSitename]
	if articleDB == nil {
		currentDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			resData.Code = -50
			resData.Errmsg = "无法读取数据库文件_" + err.Error()
			return resData, nil
		}
		articleDB = &db.ArticleObjDB{}
		dbpath := filepath.Join(currentDir, "data", useract.User.SourceSitename+".db")

		err = articleDB.OpenDB("sqlite3", dbpath)
		if err != nil {
			resData.Code = -50
			resData.Errmsg = "无法打开数据库_" + err.Error()
			return resData, nil
		}
		cs.articleDBMap[useract.User.SourceSitename] = articleDB
	}

	articleList, user, err := articleDB.QueryArticlesByUser(ctx, useract.User.SourceId, useract.User.AccountNum)

	if err != nil {
		resData.Code = -50
		resData.Errmsg = "数据库查询错误_" + err.Error()
	} else {
		resData.Articles = articleList
		resData.Author = user
	}
	return resData, nil
}
