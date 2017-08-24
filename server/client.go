package server

import (
	"google.golang.org/grpc"
	"github.com/xfort/RockCrawler/proto"
)

type CrawlerClient struct {
	proto.RockCrawlerClient
}

func NewCrawlerClient(addr string) (*CrawlerClient, error) {

	grpcClient, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	client := proto.NewRockCrawlerClient(grpcClient)
	cc := &CrawlerClient{RockCrawlerClient: client}
	return cc, nil
}

func (cc *CrawlerClient) Close() {
	if cc.RockCrawlerClient != nil {
		cc.Close()
	}
}
