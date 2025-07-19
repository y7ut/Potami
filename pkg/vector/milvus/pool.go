package milvus

import (
	"container/list"
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/milvus-io/milvus/client/v2/milvusclient"
	"github.com/sirupsen/logrus"
)

const (
	// BatchSize 批量策略执行的并发数量
	BatchSize = 30
)

// MilvusConnectionPool 连接池
type MilvusConnectionPool struct {
	connections map[string]*list.Element            // 连接缓存, 用来快速查找
	connHashMap map[*list.Element]*PooledConnection // 连接缓存
	lru         *list.List                          // LRU链表 List[*PooledConnection] 用于实现LRU

	maxSize         int           // 最大连接数
	idleTimeout     time.Duration // 空闲超时时间
	cleanUpInterval time.Duration // 清理间隔

	mu     sync.RWMutex  // 读写锁
	stopCh chan struct{} // 停止信号

	config *milvusclient.ClientConfig
}

// NewMilvusConnectionPool 创建连接池
func NewMilvusConnectionPool(maxSize int, idleTimeout time.Duration, cleanUpInterval time.Duration, config *milvusclient.ClientConfig) *MilvusConnectionPool {

	return &MilvusConnectionPool{
		connections:     make(map[string]*list.Element, maxSize),
		connHashMap:     make(map[*list.Element]*PooledConnection, maxSize),
		lru:             list.New(),
		maxSize:         maxSize,
		idleTimeout:     idleTimeout,
		cleanUpInterval: cleanUpInterval,
		stopCh:          make(chan struct{}),

		config: config,
	}
}

// Start 启动连接池
func (p *MilvusConnectionPool) Start() {
	go func() {
		cleanUpticker := time.NewTicker(p.cleanUpInterval)
		defer cleanUpticker.Stop()

		logrus.WithFields(logrus.Fields{
			"max_size":         p.maxSize,
			"idle_timeout":     p.idleTimeout,
			"cleanup_interval": p.cleanUpInterval,
		}).Info("Milvus connection pool started")
		var scratch []*list.Element
		for {
			select {
			case <-cleanUpticker.C:
				// 清理LRU和健康检查
				logrus.Debug("Starting connection pool cleanup and health check")
				p.healthCheck(context.Background(), &scratch)
				p.cleanLRU(context.Background(), &scratch)

			case <-p.stopCh:
				logrus.Info("Milvus connection pool stopping")
				p.safeClose(context.Background(), &scratch)
				return
			}
		}
	}()
}

// removeConnCache 从缓存中移除连接
func (p *MilvusConnectionPool) removeConnCache(elements ...*list.Element) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, e := range elements {
		p.lru.Remove(e)
		delete(p.connections, e.Value.(*PooledConnection).params.Hash())
		delete(p.connHashMap, e)
	}
}

// removeConns 批量移除一些的连接
func (p *MilvusConnectionPool) removeConns(ctx context.Context, scratch *[]*list.Element, causeBy string) {
	// 移除不经常使用的连接
	p.removeConnCache((*scratch)...)
	// 分组关闭链接，每次操作5个，避免一次性删除太多
	for i := 0; i < len(*scratch); i += BatchSize {
		var wg sync.WaitGroup
		endOffset := i + BatchSize
		if endOffset > len(*scratch) {
			endOffset = len(*scratch)
		}
		for _, e := range (*scratch)[i:endOffset] {
			wg.Add(1)
			currentElementConn := e.Value.(*PooledConnection)
			logrus.WithFields(logrus.Fields{
				"connection_id": currentElementConn.params.Hash(),
				"collection":    currentElementConn.params.Collection,
				"partition":     currentElementConn.params.Partition,
			}).Infof("connections removed, cause by: %s", causeBy)

			go func(conn *PooledConnection) {
				// 关闭连接,
				err := conn.Close(ctx)
				if err != nil {
					logrus.WithFields(logrus.Fields{
						"connection_id": conn.params.Hash(),
						"collection":    conn.params.Collection,
						"partition":     conn.params.Partition,
					}).WithError(err).Error("Failed to close connection")
				}
				wg.Done()
			}(currentElementConn)
		}
		wg.Wait()

		logrus.WithFields(logrus.Fields{
			"batch_connections_removed": len((*scratch)[i:endOffset]),
		}).Debug("connections removed successfully")
	}
}

// safeClose 退出连接池
func (p *MilvusConnectionPool) safeClose(ctx context.Context, scratch *[]*list.Element) {
	p.mu.RLock()
	// 获取所有连接
	for _, e := range p.connections {
		*scratch = append((*scratch), e)
	}
	p.mu.RUnlock()

	if len(*scratch) == 0 {
		return
	}

	// 清理所有连接
	p.removeConns(ctx, scratch, "pool exit")
	// 清空切片
	*scratch = (*scratch)[:0]
}

// cleanLRU 清理LRU
func (p *MilvusConnectionPool) cleanLRU(ctx context.Context, scratch *[]*list.Element) {
	now := time.Now()

	p.mu.RLock()
	// 遍历所有连接，找到需要清理的连接
	for e, size := p.lru.Front(), p.lru.Len(); e != nil; e, size = e.Next(), size-1 {
		currentElementConn := e.Value.(*PooledConnection)
		if currentElementConn.lastUsed.IsZero() {
			continue
		}
		if now.Sub(currentElementConn.lastUsed) > p.idleTimeout || size > p.maxSize {
			if atomic.LoadInt32(&currentElementConn.refCount) > 0 {
				// 如果当前元素的引用计数大于0，那么就不清理了
				continue
			}
			*scratch = append((*scratch), e)
		} else {
			// 如果最后一个元素没有超时，那么就不需要再遍历了
			break
		}
	}
	p.mu.RUnlock()
	if len(*scratch) == 0 {
		// 如果没有需要清理的连接，那么就直接返回
		return
	}
	// 清理 lru 闲置连接
	p.removeConns(ctx, scratch, "lru idle timeout")
	// 清空切片
	*scratch = (*scratch)[:0]
}

// healthCheck 健康检查所有连接
func (p *MilvusConnectionPool) healthCheck(ctx context.Context, scratch *[]*list.Element) {

	p.mu.RLock()
	for _, e := range p.connections {
		*scratch = append((*scratch), e)
	}
	p.mu.RUnlock()

	// 遍历所有连接进行健康检查, 这里时间可能比较耗时, 所以提前获取全部链接，避免锁占用
	var t int
	for _, e := range *scratch {
		currentElementConn := e.Value.(*PooledConnection)
		// 只检查空闲连接（引用计数为0）
		if atomic.LoadInt32(&currentElementConn.refCount) == 0 {
			// 设定一下超时时间
			ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
			err := currentElementConn.isHealthy(ctx)
			cancel()
			if err != nil {
				warningMsg := "Connection health check failed, marked for removal"
				if errors.Is(err, context.DeadlineExceeded) {
					warningMsg = "Connection health check timed out, marked for removal"
				}
				logrus.WithFields(logrus.Fields{
					"connection_id": currentElementConn.params.Hash(),
					"collection":    currentElementConn.params.Collection,
					"partition":     currentElementConn.params.Partition,
				}).Warn(warningMsg)
				(*scratch)[t] = e
				t++
				continue
			}
		}
	}
	// 检查失败的连接会保留在带删除的切片中
	(*scratch) = (*scratch)[:t]

	// 移除不健康的连接
	p.removeConns(ctx, scratch, "unhealthy")

	// 清空切片
	*scratch = (*scratch)[:0]
}

// Stop 停止连接池
func (p *MilvusConnectionPool) Stop() {
	close(p.stopCh)
	time.Sleep(5000 * time.Millisecond)
}

// cleanOneIdleConnection 清理一个空闲连接，返回是否成功清理
func (p *MilvusConnectionPool) cleanOneIdleConnection(ctx context.Context) bool {
	// 从前往后遍历，找到第一个空闲连接并清理
	for e := p.lru.Front(); e != nil; e = e.Next() {
		currentElementConn := e.Value.(*PooledConnection)
		if atomic.LoadInt32(&currentElementConn.refCount) == 0 {
			logrus.WithFields(logrus.Fields{
				"connection_id": currentElementConn.params.Hash(),
				"collection":    currentElementConn.params.Collection,
				"partition":     currentElementConn.params.Partition,
			}).Warn("Cleaning once idle connection due to pool full")

			p.lru.Remove(e)
			delete(p.connections, currentElementConn.params.Hash())
			delete(p.connHashMap, e)

			err := currentElementConn.Close(ctx)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"connection_id": currentElementConn.params.Hash(),
					"collection":    currentElementConn.params.Collection,
					"partition":     currentElementConn.params.Partition,
				}).WithError(err).Error("Failed to close one idle connection")
				continue
			}

			return true
		}
	}
	return false
}

// createNewConnection 创建一个新的连接
func (p *MilvusConnectionPool) createNewConnection(ctx context.Context, params *MilvusParams) (*PooledConnection, error) {

	client, err := milvusclient.New(ctx, p.config)
	if err != nil {
		return nil, err
	}

	exists, err := client.HasCollection(ctx, milvusclient.NewDescribeCollectionOption(params.Collection))
	if err != nil {
		return nil, fmt.Errorf("failed to describe collection: %v", err)
	}
	if !exists {
		err = createCollection(ctx, client, params.Collection, params.useBM25, params.useContextEmbed, params.Dimensions)
		if err != nil {
			return nil, fmt.Errorf("failed to create collection: %v", err)
		}
	}
	if params.Partition != "" {
		existsPartition, err := client.HasPartition(ctx, milvusclient.NewHasPartitionOption(params.Collection, params.Partition))
		if err != nil {
			return nil, fmt.Errorf("failed to describe partition: %v", err)
		}
		if !existsPartition {
			err = client.CreatePartition(ctx, milvusclient.NewCreatePartitionOption(params.Collection, params.Partition))
			if err != nil {
				return nil, fmt.Errorf("failed to create partition: %v", err)
			}
		}
	}

	loadTask, err := client.LoadCollection(ctx, milvusclient.NewLoadCollectionOption(params.Collection))
	if err != nil {
		return nil, fmt.Errorf("failed to load collection: %v", err)
	}
	// sync wait collection to be loaded
	err = loadTask.Await(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load collection: %v", err)
	}

	if params.Partition != "" {
		loadTask, err = client.LoadPartitions(ctx, milvusclient.NewLoadPartitionsOption(params.Collection, params.Partition))
		if err != nil {
			return nil, fmt.Errorf("failed to load partition: %v", err)
		}
		// sync wait collection to be loaded
		err = loadTask.Await(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to load partition: %v", err)
		}
	}

	return &PooledConnection{
		client:   client,
		params:   params,
		refCount: 0,
	}, nil
}

// GetConnection 从连接缓存中获取连接
func (p *MilvusConnectionPool) GetConnection(ctx context.Context, options ...MilvusParamsOption) (*PooledConnection, error) {

	params := NewMilvusParams()
	for _, option := range options {
		option(params)
	}

	currentConnection, err := p.getFromCache(ctx, params)
	if err != nil {
		return nil, err
	}

	return currentConnection, nil
}

// addToCache 加入连接缓存
func (p *MilvusConnectionPool) addToCache(hash string, connection *PooledConnection) {
	// 加入连接缓存
	p.connections[hash] = p.lru.PushBack(connection)
	// 注册连接缓存
	p.connHashMap[p.lru.Back()] = connection
	logrus.WithField("connection_id", hash).Debug("Connection added to cache")
}

// getFromCache 从连接缓存中获取连接
func (p *MilvusConnectionPool) getFromCache(ctx context.Context, params *MilvusParams) (*PooledConnection, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	var currentConnection *PooledConnection
	hash := params.Hash()
	logrus.WithField("connection_id", hash).Debug("Looking for connection")

	// 如果连接缓存为空
	if p.lru.Len() == 0 {
		var err error
		logrus.Debug("Connection pool is empty, creating new connection")
		currentConnection, err = p.createNewConnection(ctx, params)
		if err != nil {
			return nil, err
		}
		// 加入连接缓存
		p.addToCache(params.Hash(), currentConnection)
		return currentConnection, nil
	}

	// 缓存不为空，先查找是否已存在
	if c, ok := p.connections[hash]; ok {
		logrus.WithField("connection_id", hash).Debug("Connection pool cache hit")
		// 查找到了
		p.lru.MoveToBack(c)
		currentConnection = p.connHashMap[c]
		return currentConnection, nil
	}

	// 没有命中缓存，需要创建新连接
	logrus.WithField("connection_id", hash).Debug("Connection pool cache miss")

	// 检查是否达到最大连接数
	if p.lru.Len() >= p.maxSize {
		// 尝试清理一个空闲连接
		if !p.cleanOneIdleConnection(ctx) {
			return nil, fmt.Errorf("connection pool is full with no idle connections available, cannot create connection [%s]", hash)
		}
	}

	// 创建新连接
	var err error
	currentConnection, err = p.createNewConnection(ctx, params)
	if err != nil {
		return nil, err
	}
	// 加入连接缓存
	p.addToCache(hash, currentConnection)

	return currentConnection, nil
}
