package queue

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/taerc/vpublish/internal/model"
	"github.com/taerc/vpublish/internal/repository"
)

// ErrorQueue 错误记录队列
// 使用内存队列缓冲报错记录，批量写入数据库，减少数据库压力
type ErrorQueue struct {
	queue          chan *model.ErrorRecord
	repo           *repository.ErrorRecordRepository
	batchSize      int
	flushInterval  time.Duration
	mu             sync.Mutex
	buffer         []*model.ErrorRecord
	stats          QueueStats
	statsMu        sync.RWMutex
	ctx            context.Context
	cancel         context.CancelFunc
}

// QueueStats 队列统计信息
type QueueStats struct {
	TotalPushed   int64 // 总推入数量
	TotalFlushed  int64 // 总刷新数量
	TotalDropped  int64 // 总丢弃数量
	TotalFailed   int64 // 总失败数量
	CurrentSize   int   // 当前队列大小
	BufferSize    int   // 当前缓冲区大小
}

// NewErrorQueue 创建错误队列
// repo: 数据库仓库
// queueSize: 队列最大容量
// batchSize: 批量写入大小
// flushInterval: 刷新间隔
func NewErrorQueue(repo *repository.ErrorRecordRepository, queueSize, batchSize int, flushInterval time.Duration) *ErrorQueue {
	ctx, cancel := context.WithCancel(context.Background())

	return &ErrorQueue{
		queue:         make(chan *model.ErrorRecord, queueSize),
		repo:          repo,
		batchSize:     batchSize,
		flushInterval: flushInterval,
		buffer:        make([]*model.ErrorRecord, 0, batchSize),
		ctx:           ctx,
		cancel:        cancel,
	}
}

// Push 推入报错记录到队列
// 返回 true 表示成功推入，false 表示队列已满被丢弃
func (q *ErrorQueue) Push(record *model.ErrorRecord) bool {
	select {
	case q.queue <- record:
		q.statsMu.Lock()
		q.stats.TotalPushed++
		q.stats.CurrentSize = len(q.queue)
		q.statsMu.Unlock()
		return true
	default:
		// 队列满，丢弃
		q.statsMu.Lock()
		q.stats.TotalDropped++
		q.statsMu.Unlock()
		log.Printf("[ErrorQueue] Queue full, dropping record: %s", record.RequestID)
		return false
	}
}

// Start 启动队列处理协程
func (q *ErrorQueue) Start() {
	ticker := time.NewTicker(q.flushInterval)
	defer ticker.Stop()

	log.Printf("[ErrorQueue] Started with batchSize=%d, flushInterval=%v", q.batchSize, q.flushInterval)

	for {
		select {
		case <-q.ctx.Done():
			// 优雅关闭，刷新剩余数据
			log.Printf("[ErrorQueue] Shutting down, flushing remaining records...")
			q.flush()
			return
		case record := <-q.queue:
			q.mu.Lock()
			q.buffer = append(q.buffer, record)
			q.statsMu.Lock()
			q.stats.BufferSize = len(q.buffer)
			q.statsMu.Unlock()

			if len(q.buffer) >= q.batchSize {
				q.flush()
			}
			q.mu.Unlock()
		case <-ticker.C:
			q.mu.Lock()
			q.flush()
			q.mu.Unlock()
		}
	}
}

// Stop 停止队列
func (q *ErrorQueue) Stop() {
	if q.cancel != nil {
		q.cancel()
	}
}

// flush 批量刷新缓冲区到数据库
func (q *ErrorQueue) flush() {
	if len(q.buffer) == 0 {
		return
	}

	batch := make([]*model.ErrorRecord, len(q.buffer))
	copy(batch, q.buffer)
	q.buffer = q.buffer[:0]

	q.statsMu.Lock()
	q.stats.BufferSize = 0
	q.statsMu.Unlock()

	if err := q.repo.CreateBatch(context.Background(), batch); err != nil {
		q.statsMu.Lock()
		q.stats.TotalFailed += int64(len(batch))
		q.statsMu.Unlock()
		log.Printf("[ErrorQueue] Failed to flush error batch (size=%d): %v", len(batch), err)

		// 简化处理：失败的数据不再重试，实际生产环境应该有重试机制
		// 考虑到报错数据的非关键性，这里选择丢弃失败的数据
		return
	}

	q.statsMu.Lock()
	q.stats.TotalFlushed += int64(len(batch))
	q.stats.CurrentSize = len(q.queue)
	q.statsMu.Unlock()

	log.Printf("[ErrorQueue] Successfully flushed %d records", len(batch))
}

// GetStats 获取队列统计信息
func (q *ErrorQueue) GetStats() QueueStats {
	q.statsMu.RLock()
	defer q.statsMu.RUnlock()

	return q.stats
}

// HealthCheck 健康检查
// 返回队列是否健康（队列未溢出）
func (q *ErrorQueue) HealthCheck() bool {
	q.statsMu.RLock()
	defer q.statsMu.RUnlock()

	// 队列使用率超过90%时认为不健康
	capacity := cap(q.queue)
	if capacity == 0 {
		return true
	}

	usage := float64(q.stats.CurrentSize) / float64(capacity)
	return usage < 0.9
}