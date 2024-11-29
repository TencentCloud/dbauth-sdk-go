package timer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tencentcloud/dbauth-sdk-go/dbauth/internal/constants"
)

func TestSaveTimer_ValidKeyAndDelay(t *testing.T) {
	manager := NewManager()
	taskExecuted := false
	task := func() { taskExecuted = true }

	manager.SaveTimer("validKey", 100, task)
	time.Sleep(150 * time.Millisecond)

	assert.True(t, taskExecuted)
}

func TestSaveTimer_EmptyKey(t *testing.T) {
	manager := NewManager()
	taskExecuted := false
	task := func() { taskExecuted = true }

	manager.SaveTimer("", 100, task)
	time.Sleep(150 * time.Millisecond)

	assert.False(t, taskExecuted)
}

func TestSaveTimer_NegativeDelay(t *testing.T) {
	manager := NewManager()
	taskExecuted := false
	task := func() { taskExecuted = true }

	manager.SaveTimer("validKey", -100, task)
	time.Sleep(150 * time.Millisecond)

	assert.False(t, taskExecuted)
}

func TestSaveTimer_ZeroDelay(t *testing.T) {
	manager := NewManager()
	taskExecuted := false
	task := func() { taskExecuted = true }

	manager.SaveTimer("validKey", 0, task)
	time.Sleep(150 * time.Millisecond)

	assert.False(t, taskExecuted)
}

func TestSaveTimer_DelayExceedsMaxDelay(t *testing.T) {
	manager := NewManager()
	taskExecuted := false
	task := func() { taskExecuted = true }

	manager.SaveTimer("validKey", constants.MaxDelay+1, task)
	time.Sleep(150 * time.Millisecond)

	assert.False(t, taskExecuted)
}

func TestSaveTimer_OverwriteExistingTimer(t *testing.T) {
	manager := NewManager()
	taskExecuted1 := false
	task1 := func() { taskExecuted1 = true }
	taskExecuted2 := false
	task2 := func() { taskExecuted2 = true }

	manager.SaveTimer("validKey", 200, task1)
	manager.SaveTimer("validKey", 100, task2)
	time.Sleep(300 * time.Millisecond)

	assert.False(t, taskExecuted1)
	assert.True(t, taskExecuted2)
}
