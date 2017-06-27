package crawler

import (
	"testing"
	"github.com/xfort/rockgo"
	"log"
	"encoding/json"
)

var mpCrawler *MiaopaiCrawler = &MiaopaiCrawler{}

func TestMiaopaiCrawler_LoadHomeData(t *testing.T) {

	mpCrawler.mphttp = rockgo.NewRockHttp()

	articleArray, _, err := mpCrawler.LoadHomeArticles("http://www.miaopai.com/v2_index/u/paike_zkbv7sra9b")

	if err != nil {
		log.Fatalln(err)
	}
	jsonByte, err := json.Marshal(articleArray)

	log.Println(err, string(jsonByte))
}
