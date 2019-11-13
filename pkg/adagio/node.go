package adagio

import "strings"

// RetryCondition is a key used in the node spec retry map
type RetryCondition string

const (
	// OnFail is the retry condition where a node results in a
	// failure
	OnFail RetryCondition = "fail"
	// OnError is the retry condition where a node results in a
	// system related error
	OnError RetryCondition = "error"
)

// CanRetry returns true if the node can be retried
func CanRetry(node *Node) (canRetry bool) {
	VisitLatestAttempt(node, func(result *Node_Result) {
		// check for retries
		retryKey := strings.ToLower(result.Conclusion.String())
		if retry, ok := node.Spec.Retry[retryKey]; ok {
			var retryCount int32

			// count number of existing attempts
			for _, r := range node.Attempts {
				if r.Conclusion == result.Conclusion {
					retryCount++
				}
			}

			canRetry = retryCount < retry.MaxAttempts
		}
	})

	return
}

// VisitLatestAttempt calls the supplied function with the latest attempt result
// if any attempts have been made
func VisitLatestAttempt(node *Node, fn func(*Node_Result)) {
	if len(node.Attempts) < 1 {
		return
	}

	fn(node.Attempts[len(node.Attempts)-1])
}
