package nuts_test

import (
	"github.com/RealFax/RedQueen/internal/rqd/store/nuts"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	watcher = &nuts.Watcher{}
)

func TestWatcher_UseTarget(t *testing.T) {
	child := watcher.UseTarget("rqd")
	assert.NotNil(t, child)
	assert.Equal(t, "rqd", child.Namespace)

	child2 := watcher.UseTarget("rqd2")
	assert.NotNil(t, child2)
	assert.Equal(t, "rqd2", child2.Namespace)
}

func TestWatcherChild_Watch(t *testing.T) {
	child := watcher.UseTarget("rqd")
	notifier := child.Watch(pair1.Key)
	assert.NotNil(t, notifier)
	defer func() {
		assert.NoError(t, notifier.Close())
	}()

	for i := 0; i < 10; i++ {
		child.Update(pair1.Key, pair1.Value, 0)
		value := <-notifier.Notify()
		assert.NotNil(t, value.Value)
		assert.Equal(t, pair1.Value, *value.Value)
	}
}

func TestWatcherChild_WatchPrefix(t *testing.T) {
	child := watcher.UseTarget("rqd")
	notifier := child.WatchPrefix([]byte("K"))
	assert.NotNil(t, notifier)
	defer func() {
		assert.NoError(t, notifier.Close())
	}()

	for _, node := range pairMatrix {
		child.Update(node.Key, node.Value, 0)
		value := <-notifier.Notify()
		assert.NotNil(t, value.Value)
		assert.Equal(t, node.Value, *value.Value)
	}
}
