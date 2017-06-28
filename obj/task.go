package obj

type TaskObj struct {
	TaskUrl     string
	Name        string
	CollectCode int //采集标识，0=不采集，1=采集
	Publish     int //发布标识，0=不发布，1=发布
}
