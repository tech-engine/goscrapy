package stats

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockStatsCollector struct {
	mock.Mock
}

func (m *MockStatsCollector) Name() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockStatsCollector) Snapshot() ComponentSnapshot {
	args := m.Called()
	return args.Get(0).(ComponentSnapshot)
}

type MockStatsObserver struct {
	mock.Mock
}

func (m *MockStatsObserver) OnSnapshot(s GlobalSnapshot) {
	m.Called(s)
}

func TestHub_Broadcast(t *testing.T) {
	interval := 50 * time.Millisecond
	h := NewTelemetryHub(WithInterval(interval))

	mockCollector := new(MockStatsCollector)
	mockCollector.On("Name").Return("mock")
	mockCollector.On("Snapshot").Return("dummy payload")

	h.AddCollector(mockCollector)

	mockObserver := new(MockStatsObserver)
	mockObserver.On("OnSnapshot", mock.AnythingOfType("stats.GlobalSnapshot")).Return().Run(func(args mock.Arguments) {
		snap := args.Get(0).(GlobalSnapshot)
		assert.Equal(t, interval, snap.Interval)
		assert.Equal(t, "dummy payload", snap.Components["mock"])
	})

	h.AddObserver(mockObserver)

	ctx, cancel := context.WithCancel(context.Background())
	
	go func() {
		err := h.Start(ctx)
		assert.NoError(t, err)
	}()

	time.Sleep(interval * 2)
	cancel()

	mockCollector.AssertExpectations(t)
	mockObserver.AssertExpectations(t)
}
