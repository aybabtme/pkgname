package main

type leakingQueue struct {
	max int
	vec []string
}

func newQueue(size int) *leakingQueue {
	return &leakingQueue{
		max: size,
		vec: make([]string, 0, size),
	}
}

func (l *leakingQueue) Enqueue(s string) {
	if len(l.vec) >= l.max-1 {
		l.vec = l.vec[1:]
	}
	l.vec = append(l.vec, s)
}

func (l *leakingQueue) Last(size int) []string {
	return l.vec[min(len(l.vec)-size, 0):]
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
