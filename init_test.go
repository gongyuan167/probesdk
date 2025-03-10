package probesdk

import (
	"context"
	"testing"
)

func TestMetrics(t *testing.T) {
	RequestCount.Add(context.Background(), 1)
}
