package stats

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockSnapshotter is a mock type for Snapshotter
type MockSnapshotter struct {
	mock.Mock
}

func (m *MockSnapshotter) Snapshot() Snapshot {
	args := m.Called()
	return args.Get(0).(Snapshot)
}

// MockStatsObserver is a mock type for StatsObserver
type MockStatsObserver struct {
	mock.Mock
}

func (m *MockStatsObserver) OnSnapshot(s Snapshot) {
	m.Called(s)
}

func TestBroadcaster_Broadcast(t *testing.T) {
	mockSnapshotter := new(MockSnapshotter)
	interval := 50 * time.Millisecond
	b := NewBroadcaster(mockSnapshotter, WithInterval(interval))

	expectedSnapshot := Snapshot{
		TotalRequests: 10,
		TotalBytes:    1024,
	}

	mockSnapshotter.On("Snapshot").Return(expectedSnapshot)

	mockObserver := new(MockStatsObserver)
	mockObserver.On("OnSnapshot", expectedSnapshot).Return().Run(func(args mock.Arguments) {
		// Signal completion
	})

	b.Subscribe(mockObserver)

	ctx, cancel := context.WithCancel(context.Background())
	
	go func() {
		err := b.Start(ctx)
		assert.NoError(t, err)
	}()

	// Wait for at least one broadcast
	time.Sleep(interval * 2)
	cancel()

	mockSnapshotter.AssertExpectations(t)
	mockObserver.AssertExpectations(t)
}
