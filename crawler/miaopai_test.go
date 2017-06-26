package crawler

import (
	"testing"
	"github.com/xfort/rockgo"
	"github.com/xfort/RockCrawler/obj"
	"log"
)

var mpCrawler *MiaopaiCrawler = &MiaopaiCrawler{}

func TestMiaopaiCrawler_LoadHomeData(t *testing.T) {

	mpCrawler.mphttp = rockgo.NewRockHttp()
	userObj := &obj.UserObj{}
	suid, userObj, err := mpCrawler.LoadHomeData("http://www.miaopai.com/v2_index/u/paike_zkbv7sra9b", userObj)

	log.Println(err, suid, userObj.Nickname)
}
