package util

import (
	"net/http"
	"time"

	backoff "github.com/cenkalti/backoff/v4"
)

type resourceResponse[T any] struct {
	Resource T
	Response *http.Response
}

type operationWithResourceResponseData[T any] func() (T, *http.Response, error)

type resourceResponseOperation[T any] struct {
	operation operationWithResourceResponseData[T]
}

func (rro *resourceResponseOperation[T]) Execute() (resourceResponse[T], error) {
	resource, response, error := rro.operation()
	return resourceResponse[T]{resource, response}, error
}

func getExponentialBackOff() *backoff.ExponentialBackOff {
	b := &backoff.ExponentialBackOff{
		InitialInterval:     backoff.DefaultInitialInterval,
		RandomizationFactor: backoff.DefaultRandomizationFactor,
		Multiplier:          2,
		MaxInterval:         backoff.DefaultMaxInterval,
		MaxElapsedTime:      5 * time.Minute,
		Stop:                backoff.Stop,
		Clock:               backoff.SystemClock,
	}
	b.Reset()
	return b
}

// For requests that return resource info, a response and a error
// This function combines that into one struct so the retry library can be used and then flattens before returning
func RetryResourceResponse[T any](operationWithResourceResponse operationWithResourceResponseData[T]) (T, *http.Response, error) {
	rro := resourceResponseOperation[T]{operationWithResourceResponse}
	resourceResponse, err := backoff.RetryWithData[resourceResponse[T]](rro.Execute, getExponentialBackOff())
	return resourceResponse.Resource, resourceResponse.Response, err
}

// For requests that just return a response and an error
func RetryResponse(operationWithData backoff.OperationWithData[*http.Response]) (*http.Response, error) {
	return backoff.RetryWithData[*http.Response](operationWithData, getExponentialBackOff())
}
