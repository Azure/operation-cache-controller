package reconciler

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOperations(t *testing.T) {
	t.Run("ContinueOperationResult", func(t *testing.T) {
		result := ContinueOperationResult()
		assert.Equal(t, OperationResult{
			RequeueDelay:   0,
			RequeueRequest: false,
			CancelRequest:  false,
		}, result)
	})
	t.Run("StopOperationResult", func(t *testing.T) {
		result := StopOperationResult()
		assert.Equal(t, OperationResult{
			RequeueDelay:   0,
			RequeueRequest: false,
			CancelRequest:  true,
		}, result)
	})

	t.Run("StopProcessing", func(t *testing.T) {
		result, err := StopProcessing()
		assert.Nil(t, err)
		assert.Equal(t, OperationResult{
			RequeueDelay:   0,
			RequeueRequest: false,
			CancelRequest:  true,
		}, result)
	})

	t.Run("Requeue", func(t *testing.T) {
		result, err := Requeue()
		assert.Nil(t, err)
		assert.Equal(t, OperationResult{
			RequeueDelay:   DefaultRequeueDelay,
			RequeueRequest: true,
			CancelRequest:  false,
		}, result)
	})

	t.Run("RequeueWithError", func(t *testing.T) {
		result, err := RequeueWithError(errors.New("test"))
		assert.NotNil(t, err)
		assert.Equal(t, OperationResult{
			RequeueDelay:   DefaultRequeueDelay,
			RequeueRequest: true,
			CancelRequest:  false,
		}, result)
	})

	t.Run("RequeueOnErrorOrStop", func(t *testing.T) {
		result, err := RequeueOnErrorOrStop(errors.New("test"))
		assert.NotNil(t, err)
		assert.Equal(t, OperationResult{
			RequeueDelay:   DefaultRequeueDelay,
			RequeueRequest: false,
			CancelRequest:  true,
		}, result)
	})

	t.Run("RequeueOnErrorOrContinue", func(t *testing.T) {
		result, err := RequeueOnErrorOrContinue(errors.New("test"))
		assert.NotNil(t, err)
		assert.Equal(t, OperationResult{
			RequeueDelay:   DefaultRequeueDelay,
			RequeueRequest: false,
			CancelRequest:  false,
		}, result)
	})

	t.Run("RequeueAfter", func(t *testing.T) {
		result, err := RequeueAfter(1, errors.New("test"))
		assert.NotNil(t, err)
		assert.Equal(t, OperationResult{
			RequeueDelay:   1,
			RequeueRequest: true,
			CancelRequest:  false,
		}, result)
	})

	t.Run("ContinueProcessing", func(t *testing.T) {
		result, err := ContinueProcessing()
		assert.Nil(t, err)
		assert.Equal(t, result, OperationResult{
			RequeueDelay:   0,
			RequeueRequest: false,
			CancelRequest:  false,
		})
	})

	t.Run("RequeueOrCancel", func(t *testing.T) {
		result1 := OperationResult{
			RequeueDelay:   0,
			RequeueRequest: false,
			CancelRequest:  false,
		}
		assert.Equal(t, result1.RequeueOrCancel(), false)

		result2 := OperationResult{
			RequeueDelay:   0,
			RequeueRequest: true,
			CancelRequest:  false,
		}
		assert.Equal(t, result2.RequeueOrCancel(), true)

		result3 := OperationResult{
			RequeueDelay:   0,
			RequeueRequest: false,
			CancelRequest:  true,
		}
		assert.Equal(t, result3.RequeueOrCancel(), true)
	})
}
