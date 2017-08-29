package server

import (
	"testing"
	"log"
	"time"
	"golang.org/x/net/context"
	"github.com/xfort/RockCrawler/proto"
	_ "github.com/mattn/go-sqlite3"
)

func TestGRPC(t *testing.T) {

	go testGRPCServer()
	testGRPCClient()
	time.Sleep(10 * time.Second)
}

func testGRPCServer() {
	err := StartHttpServer("127.0.0.1:9900")
	if err != nil {
		log.Fatalln("启动grpc失败", err)
	}
}

func testGRPCClient() {
	client, err := NewCrawlerClient("127.0.0.1:9900")
	if err != nil {
		log.Fatalln(err)
	}
	res, err := client.LoadArticlesByUser(context.TODO(), &proto.UserAct{User: &proto.UserObj{SourceId: "5237828", SourceSitename: "QQKuaiBao"}})
	if err != nil {
		log.Fatalln("LoadArticlesByUser 异常", err)
	}
	log.Println(res.Errmsg)
}
