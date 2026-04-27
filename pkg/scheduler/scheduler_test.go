package scheduler

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tech-engine/goscrapy/pkg/core"
)

func TestScheduler_New(t *testing.T) {
	s, err := New(nil)
	require.NoError(t, err)
	assert.NotNil(t, s)
}

func TestScheduler_ScheduleAndNextRequest(t *testing.T) {
	s, err := New(&Config{
		WorkQueueSize: 10,
	})
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go s.Start(ctx)

	req := &core.Request{URL: nil}
	cbName := "test_cb"

	err = s.Schedule(req, cbName)
	assert.NoError(t, err)

	pickedReq, pickedCb, handle, err := s.NextRequest(ctx)
	assert.NoError(t, err)
	assert.Equal(t, req, pickedReq)
	assert.Equal(t, cbName, pickedCb)
	assert.NotNil(t, handle)
}

func TestScheduler_AckNack(t *testing.T) {
	s, err := New(&Config{
		WorkQueueSize: 10,
	})
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go s.Start(ctx)

	req := &core.Request{}
	_ = s.Schedule(req, "cb")

	_, _, handle, _ := s.NextRequest(ctx)

	// test Ack
	err = s.Ack(handle)
	assert.NoError(t, err)

	// test Nack
	err = s.Nack(handle)
	assert.NoError(t, err)
}
