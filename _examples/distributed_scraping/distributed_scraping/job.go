package distributed_scraping

type Job struct {
	id string
}

func NewJob(id string) *Job {
	return &Job{
		id: id,
	}
}

func (j *Job) Id() string {
	return j.id
}
