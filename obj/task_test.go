package obj

import (
	"testing"
	"log"
	"encoding/json"
)

func TestParseConfigFile(t *testing.T) {

	taskArray, err := ParseConfigFile("/Users/xs/work/go/code/work/src/github.com/xfort/RockCrawler/config_duowan.json", "/Users/xs/work/go/code/work/src/github.com/xfort/RockCrawler/publisher_config.json")

	log.Println(err)
	resBytes, err := json.Marshal(taskArray)
	log.Println(string(resBytes), err)
}
