package scrapeThisSite

import (
	"errors"
)

// you can define your errors here

var ERR_EXTRACTING_META = errors.New("meta could not be parsed from context")
var ERR_EXTRACTING_JOB = errors.New("job could not be parsed from context.meta")
