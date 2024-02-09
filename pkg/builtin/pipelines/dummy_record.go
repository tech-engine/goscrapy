package pipelines

type dummyRecord struct {
	Id   string `json:"id" csv:"id"`
	Name string `json:"name" csv:"name"`
}
