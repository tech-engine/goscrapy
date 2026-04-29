package google_finance

type Job struct {
	id    string
	query string
}

func NewJob(query string) *Job {
	return &Job{
		id:    query,
		query: query,
	}
}

func (j *Job) Id() string {
	return j.id
}

func (j *Job) Query() string {
	return j.query
}
