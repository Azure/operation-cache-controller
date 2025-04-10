package reconciler

import (
	"context"
	"time"
)

var DefaultRequeueDelay = 10 * time.Second

type ReconcileOperation func(ctx context.Context) (OperationResult, error)

type OperationResult struct {
	RequeueDelay   time.Duration
	RequeueRequest bool
	CancelRequest  bool
}

func (r OperationResult) RequeueOrCancel() bool {
	return r.RequeueRequest || r.CancelRequest
}

func ContinueOperationResult() OperationResult {
	return OperationResult{
		RequeueDelay:   0,
		RequeueRequest: false,
		CancelRequest:  false,
	}
}
func StopOperationResult() OperationResult {
	return OperationResult{
		RequeueDelay:   0,
		RequeueRequest: false,
		CancelRequest:  true,
	}
}

func StopProcessing() (OperationResult, error) {
	result := StopOperationResult()
	return result, nil
}

func Requeue() (result OperationResult, err error) {
	result = OperationResult{
		RequeueDelay:   DefaultRequeueDelay,
		RequeueRequest: true,
		CancelRequest:  false,
	}
	return result, nil
}

func RequeueWithError(err error) (OperationResult, error) {
	result := OperationResult{
		RequeueDelay:   DefaultRequeueDelay,
		RequeueRequest: true,
		CancelRequest:  false,
	}
	return result, err
}

func RequeueOnErrorOrStop(err error) (OperationResult, error) {
	return OperationResult{
		RequeueDelay:   DefaultRequeueDelay,
		RequeueRequest: false,
		CancelRequest:  true,
	}, err
}

func RequeueOnErrorOrContinue(err error) (OperationResult, error) {
	return OperationResult{
		RequeueDelay:   DefaultRequeueDelay,
		RequeueRequest: false,
		CancelRequest:  false,
	}, err
}

func RequeueAfter(delay time.Duration, err error) (OperationResult, error) {
	return OperationResult{
		RequeueDelay:   delay,
		RequeueRequest: true,
		CancelRequest:  false,
	}, err
}
func ContinueProcessing() (OperationResult, error) {
	return ContinueOperationResult(), nil
}
