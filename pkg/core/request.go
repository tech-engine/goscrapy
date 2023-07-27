package core

func (r *Request) SetUrl(url string) RequestWriter {
	r.url = url
	return r
}

func (r *Request) SetMethod(method string) RequestWriter {
	r.method = method
	return r
}

func (r *Request) SetBody(body any) RequestWriter {
	r.body = body
	return r
}

func (r *Request) SetHeaders(headers map[string]string) RequestWriter {
	r.headers = headers
	return r
}

func (r *Request) SetMetaData(key string, val any) RequestWriter {
	if r.meta == nil {
		r.meta = make(map[string]any)
	}
	r.meta[key] = val
	return r
}

func (r *Request) MetaData() map[string]any {
	if r.meta == nil {
		return nil
	}

	return r.meta
}

func (r *Request) MetaDataKey(key string) (any, bool) {
	if r.meta == nil {
		return nil, false
	}

	val, ok := r.meta[key]
	return val, ok
}

func (r *Request) Url() string {
	return r.url
}

func (r *Request) Headers() map[string]string {
	return r.headers
}

func (r *Request) Body() any {
	return r.body
}

func (r *Request) Method() string {
	return r.method
}

func (r *Request) Reset() {
	r.method = "GET"
	r.url = ""
	r.headers = nil
	r.body = nil
	r.meta = nil
}
