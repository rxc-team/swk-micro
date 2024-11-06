package jobx

// ConcurrencyQueue 并发队列
type ConcurrencyQueue struct {
	queue chan struct{}
}

// Add 添加任务到队列中
func (cq *ConcurrencyQueue) Add() {
	cq.queue <- struct{}{}
}

// Done 出对列
func (cq *ConcurrencyQueue) Done() {
	<-cq.queue
}
