package adagio

// VisitLatestAttempt calls the supplied function with the latest attempt result
// if any attempts have been made
func VisitLatestAttempt(node *Node, fn func(*Node_Result)) {
	if len(node.Attempts) < 1 {
		return
	}

	fn(node.Attempts[len(node.Attempts)-1])
}
