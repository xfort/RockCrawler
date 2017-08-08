package obj

import (
	"github.com/bitly/go-simplejson"
	"errors"
	"io/ioutil"
)

type TaskObj struct {
	TaskUrl     string
	Name        string
	CollectCode int //采集标识，0=不采集，1=采集
	PublishCode int //发布标识，0=不发布，1=发布

	Publisers []*PublisherObj
}

//存储发布配置
type PublisherObj struct {
	Id        string
	Name      string
	Url       string
	HeaderObj map[string]string
	BodyObj   map[string]string
}

func ParseConfigFile(taskFile string, publisherFile string) ([]*TaskObj, error) {
	contentByte, err := ioutil.ReadFile(taskFile)
	if err != nil {
		return nil, err
	}
	taskArray, err := ParseTaskConfig(contentByte)
	if err != nil {
		return taskArray, err
	}
	contentByte, err = ioutil.ReadFile(publisherFile)
	if err != nil {
		return nil, err
	}
	publisherArray, err := ParsePublisherConfig(contentByte)
	if err != nil {
		return nil, err
	}
	publisherDefaultMap := make(map[string]*PublisherObj, len(publisherArray))

	for _, item := range publisherArray {
		publisherDefaultMap[item.Id] = item
	}

	for _, item := range taskArray {

		for _, itemPublisher := range item.Publisers {
			defaultPublisher := publisherDefaultMap[itemPublisher.Id]
			itemPublisher.Name = defaultPublisher.Name
			itemPublisher.Url = defaultPublisher.Url

			if defaultPublisher.HeaderObj != nil && len(defaultPublisher.HeaderObj) > 0 {

				if itemPublisher.HeaderObj == nil {
					itemPublisher.HeaderObj = make(map[string]string, len(defaultPublisher.HeaderObj))
				}
				for key, value := range defaultPublisher.HeaderObj {
					if itemPublisher.HeaderObj[key] == "" {
						itemPublisher.HeaderObj[key] = value
					}
				}
			}

			if defaultPublisher.BodyObj != nil && len(defaultPublisher.BodyObj) > 0 {

				if itemPublisher.BodyObj == nil {
					itemPublisher.BodyObj = make(map[string]string, len(defaultPublisher.BodyObj))
				}
				for key, value := range defaultPublisher.BodyObj {
					if itemPublisher.BodyObj[key] == "" {
						itemPublisher.BodyObj[key] = value
					}
				}
			}
		}
	}

	//if resByte, err := json.Marshal(publisherArray); err != nil || err == nil {
	//	log.Println("publisher_array")
	//	log.Println(string(resByte), err)
	//}
	return taskArray, nil
}

func ParseTaskConfig(bytes []byte) ([]*TaskObj, error) {
	rootJson, err := simplejson.NewJson(bytes)
	if err != nil {
		return nil, err
	}
	rootJson = rootJson.Get("tasks")
	arrayLen := len(rootJson.MustArray())

	if arrayLen <= 0 {
		return nil, errors.New("任务配置文件内，任务数为空")
	}

	taskArray := make([]*TaskObj, 0, arrayLen+1)
	for index := 0; index < arrayLen; index++ {
		itemJson := rootJson.GetIndex(index)
		taskUrl := itemJson.Get("url").MustString()

		if taskUrl == "" {
			return nil, errors.New("解析任务文件错误,任务数据异常,url字段错误——" + itemJson.Get("name").MustString())
		}
		taskObj := &TaskObj{}
		taskObj.TaskUrl = taskUrl
		taskObj.Name = itemJson.Get("name").MustString()
		taskObj.CollectCode = itemJson.Get("collect").MustInt(1)

		taskObj.PublishCode = itemJson.Get("publish").MustInt(0)
		if taskObj.PublishCode == 1 {
			publisherJson := itemJson.Get("publisher")
			publiserLen := len(publisherJson.MustArray())
			if publiserLen <= 0 {
				return nil, errors.New("解析任务文件错误,任务数据异常,publisher字段长度为空")
			}

			publisherArray, err := ParsePublishersJsonArray(publisherJson)
			if err != nil {
				return nil, err
			}
			taskObj.Publisers = publisherArray
		}

		taskArray = append(taskArray, taskObj)
	}
	return taskArray, nil
}

func ParsePublishersJsonArray(arrayJson *simplejson.Json) ([]*PublisherObj, error) {

	arrayLen := len(arrayJson.MustArray())

	if arrayLen <= 0 {
		return nil, errors.New("发布配置文件内，publisher为空")
	}
	publisherArray := make([]*PublisherObj, 0, arrayLen)

	//var err error
	for index := 0; index < arrayLen; index++ {
		item := arrayJson.GetIndex(index)

		publisher := &PublisherObj{}
		publisher.Id = item.Get("id").MustString()
		publisher.Name = item.Get("name").MustString()
		publisher.Url = item.Get("url").MustString()
		headerMap, err := item.Get("header").Map()
		if err == nil {
			publisher.HeaderObj = make(map[string]string, len(headerMap))
			for key, value := range headerMap {
				publisher.HeaderObj[key] = value.(string)
			}
			err = nil
		}

		bodyMap, err := item.Get("body").Map()
		if err == nil {
			publisher.BodyObj = make(map[string]string, len(bodyMap))
			for key, value := range bodyMap {
				publisher.BodyObj[key] = value.(string)
			}
		}

		publisherArray = append(publisherArray, publisher)
	}

	return publisherArray, nil
}
func ParsePublisherConfig(bytes []byte) ([]*PublisherObj, error) {
	rootJson, err := simplejson.NewJson(bytes)
	if err != nil {
		return nil, err
	}
	return ParsePublishersJsonArray(rootJson)
}
