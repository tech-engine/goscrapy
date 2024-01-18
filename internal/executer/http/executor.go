package http

import (
	"context"
)

func NewExecuter(client Client) *Executer {
	return &Executer{
		client: client,
	}
}

func (e *Executer) Execute(ctx context.Context, req RequestReader, res ResponseSetter) error {
	switch req.Method() {
	case "GET":
		return e.Get(ctx, req, res)
	case "POST":
		return e.Post(ctx, req, res)
	case "DELETE":
		return e.Delete(ctx, req, res)
	case "PATCH":
		return e.Patch(ctx, req, res)
	case "PUT":
		return e.Put(ctx, req, res)
	default:
		return e.Get(ctx, req, res)
	}
}

func (e *Executer) Get(ctx context.Context, req RequestReader, res ResponseSetter) (err error) {

	err = e.client.Request().
		SetContext(ctx).
		SetHeaders(req.Headers()).
		Get(res, req.Url())

	return
}

func (e *Executer) Post(ctx context.Context, req RequestReader, res ResponseSetter) (err error) {

	err = e.client.Request().
		SetContext(ctx).
		SetHeaders(req.Headers()).
		SetBody(req.Body()).
		Post(res, req.Url())

	return
}

func (e *Executer) Delete(ctx context.Context, req RequestReader, res ResponseSetter) (err error) {

	err = e.client.Request().
		SetContext(ctx).
		SetHeaders(req.Headers()).
		Delete(res, req.Url())

	return
}

func (e *Executer) Put(ctx context.Context, req RequestReader, res ResponseSetter) (err error) {

	err = e.client.Request().
		SetContext(ctx).
		SetHeaders(req.Headers()).
		SetBody(req.Body()).
		Put(res, req.Url())

	return
}

func (e *Executer) Patch(ctx context.Context, req RequestReader, res ResponseSetter) (err error) {
	err = e.client.Request().
		SetContext(ctx).
		SetHeaders(req.Headers()).
		SetBody(req.Body()).
		Patch(res, req.Url())

	return
}
