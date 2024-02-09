package pipelines

type dummyJob struct {
	id string
}

func (j *dummyJob) Id() string {
	return "dummyJob"
}
