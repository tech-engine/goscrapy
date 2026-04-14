package google_maps_scraper

import "errors"

// you can define your errors here
var ERR_GEO_CODING = errors.New("geocoding failed")
var ERR_INVALID_CURSOR = errors.New("invalid cursor")
var ERR_HTTP_CLIENT_NOT_SET = errors.New("http client not set")

var ERR_EXTRACTING_META = errors.New("meta could not be parsed from context")
var ERR_EXTRACTING_JOB = errors.New("job could not be parsed from context.meta")
