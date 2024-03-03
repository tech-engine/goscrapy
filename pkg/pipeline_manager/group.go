package pipelinemanager

import (
	"context"
	"sync"

	"github.com/tech-engine/goscrapy/pkg/core"
	"golang.org/x/sync/errgroup"
)

type Group[OUT any] struct {
	nodes        []IPipeline[OUT]
	ignoreErrors bool
}

// Create a Group which is a collection of pipelines intended to run concurrently
// as opposed to sequentially. Group implements IPipeline interface, meaning it behaves
// like a single pipeline.
//
// Common usecase: Since each pipeline in a Group runs concurrently it is not meant
// for data transformation but only for data export or other similar independent tasks.
// A Group must not modify
func NewGroup[OUT any]() *Group[OUT] {
	return &Group[OUT]{
		nodes: make([]IPipeline[OUT], 0),
	}
}

func (g *Group[OUT]) Open(ctx context.Context) error {
	if g.ignoreErrors {
		var wg sync.WaitGroup
		wg.Add(len(g.nodes))

		for _, p := range g.nodes {
			go func(_p IPipeline[OUT]) {
				defer wg.Done()
				_p.Open(ctx)
			}(p)
		}

		wg.Wait()
		return nil
	}

	group, groupCtx := errgroup.WithContext(ctx)
	for _, p := range g.nodes {
		p := p
		group.Go(func() error {
			return p.Open(groupCtx)
		})
	}
	return group.Wait()
}

func (g *Group[OUT]) Close() {
	var wg sync.WaitGroup
	wg.Add(len(g.nodes))

	for _, p := range g.nodes {
		go func(_p IPipeline[OUT]) {
			defer wg.Done()
			_p.Close()
		}(p)
	}

	wg.Wait()
}

// WithIgnoreError sets ignoreErrors = true
//
// When ignoreErrors = true, Group's ProcessItem & Open function will always return
// a nil error.
//
// When ignoreErrors = false(default), Group's ProcessItem & Open will return the first non-nil
// error. In addition to that context passed to Open function is also cancelled.
func (g *Group[OUT]) WithIgnoreError() {
	g.ignoreErrors = true
}

func (g *Group[OUT]) Add(p ...IPipeline[OUT]) {
	g.nodes = append(g.nodes, p...)
}

func (g *Group[OUT]) ProcessItem(pi IPipelineItem, out core.IOutput[OUT]) error {

	if g.ignoreErrors {
		var wg sync.WaitGroup
		wg.Add(len(g.nodes))

		for _, p := range g.nodes {
			go func(_p IPipeline[OUT]) {
				defer wg.Done()
				_p.ProcessItem(pi, out)
			}(p)
		}

		wg.Wait()
		return nil
	}

	errGroup := errgroup.Group{}
	for _, p := range g.nodes {
		p := p
		errGroup.Go(func() error {
			return p.ProcessItem(pi, out)
		})
	}
	return errGroup.Wait()
}
