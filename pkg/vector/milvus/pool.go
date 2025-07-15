package milvus

import (
	"container/list"
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/milvus-io/milvus/client/v2/milvusclient"
)

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

func (p *MilvusConnectionPool) Start() {
	cleanUpticker := time.NewTicker(p.cleanUpInterval)
	defer cleanUpticker.Stop()
	go func() {
		for {
			select {
			case <-cleanUpticker.C:
				// 清理LRU和健康检查
				log.Println("milvus connection pool start clean up and health check")
				p.cleanLRU()
				p.healthCheck()
			case <-p.stopCh:
				log.Println("milvus connection pool stop")
				p.safeExit()
				return
			}
		}
	}()

}

func (p *MilvusConnectionPool) safeExit() {
	p.mu.Lock()
	defer p.mu.Unlock()
	// 断开所有连接
	for p.lru.Len() > 0 {
		e := p.lru.Front()
		currentElementConn := e.Value.(*PooledConnection)
		err := currentElementConn.Close(context.Background())
		p.lru.Remove(e)
		delete(p.connections, currentElementConn.params.Hash())
		delete(p.connHashMap, e)

		if err != nil {
			log.Printf("milvus connection pool exit failed[%s], params: %s: %s \n", err, currentElementConn.params.Collection, currentElementConn.params.Partition)
		}
	}
}

func (p *MilvusConnectionPool) cleanLRU() {
	now := time.Now()
	p.mu.Lock()
	defer p.mu.Unlock()

	for e, size := p.lru.Front(), p.lru.Len(); e != nil; e, size = e.Next(), size-1 {
		currentElementConn := e.Value.(*PooledConnection)
		if now.Sub(currentElementConn.lastUsed) > p.idleTimeout || size > p.maxSize {
			if atomic.LoadInt32(&currentElementConn.refCount) > 0 {
				// 如果当前元素的引用计数大于0，那么就不清理了
				continue
			}
			log.Printf("milvus connection pool clean up, params: %s: %s \n", currentElementConn.params.Collection, currentElementConn.params.Partition)

			err := currentElementConn.Close(context.Background())
			if err != nil {
				log.Printf("milvus connection pool clean up failed[%s], params: %s: %s \n", err, currentElementConn.params.Collection, currentElementConn.params.Partition)
				continue
			}
			p.lru.Remove(e)
			delete(p.connections, currentElementConn.params.Hash())
			delete(p.connHashMap, e)
		} else {
			// 如果最后一个元素没有超时，那么就不需要再遍历了
			break
		}
	}
}

// cleanOneIdleConnection 清理一个空闲连接，返回是否成功清理
func (p *MilvusConnectionPool) cleanOneIdleConnection() bool {
	// 从前往后遍历，找到第一个空闲连接并清理
	for e := p.lru.Front(); e != nil; e = e.Next() {
		currentElementConn := e.Value.(*PooledConnection)
		if atomic.LoadInt32(&currentElementConn.refCount) == 0 {
			log.Printf("清理空闲连接: %s: %s \n", currentElementConn.params.Collection, currentElementConn.params.Partition)
			
			err := currentElementConn.Close(context.Background())
			if err != nil {
				log.Printf("清理空闲连接失败[%s], params: %s: %s \n", err, currentElementConn.params.Collection, currentElementConn.params.Partition)
				continue
			}
			
			p.lru.Remove(e)
			delete(p.connections, currentElementConn.params.Hash())
			delete(p.connHashMap, e)
			return true
		}
	}
	return false
}

// healthCheck 健康检查所有连接
func (p *MilvusConnectionPool) healthCheck() {
	p.mu.Lock()
	defer p.mu.Unlock()

	var toRemove []*list.Element
	
	// 遍历所有连接进行健康检查
	for e := p.lru.Front(); e != nil; e = e.Next() {
		currentElementConn := e.Value.(*PooledConnection)
		
		// 只检查空闲连接（引用计数为0）
		if atomic.LoadInt32(&currentElementConn.refCount) == 0 {
			if !currentElementConn.isHealthy() {
				log.Printf("连接健康检查失败，标记为移除: %s: %s", currentElementConn.params.Collection, currentElementConn.params.Partition)
				toRemove = append(toRemove, e)
			}
		}
	}
	
	// 移除不健康的连接
	for _, e := range toRemove {
		currentElementConn := e.Value.(*PooledConnection)
		log.Printf("移除不健康的连接: %s: %s", currentElementConn.params.Collection, currentElementConn.params.Partition)
		
		err := currentElementConn.Close(context.Background())
		if err != nil {
			log.Printf("关闭不健康连接失败[%s], params: %s: %s", err, currentElementConn.params.Collection, currentElementConn.params.Partition)
		}
		
		p.lru.Remove(e)
		delete(p.connections, currentElementConn.params.Hash())
		delete(p.connHashMap, e)
	}
	
	if len(toRemove) > 0 {
		log.Printf("健康检查完成，移除了 %d 个不健康的连接", len(toRemove))
	}
}

func (p *MilvusConnectionPool) Stop() {
	close(p.stopCh)
}

func NewMilvusConnectionPool(maxSize int, idleTimeout time.Duration, config *milvusclient.ClientConfig) *MilvusConnectionPool {

	return &MilvusConnectionPool{
		connections:     make(map[string]*list.Element, maxSize),
		connHashMap:     make(map[*list.Element]*PooledConnection, maxSize),
		lru:             list.New(),
		maxSize:         maxSize,
		idleTimeout:     idleTimeout,
		cleanUpInterval: time.Minute,
		stopCh:          make(chan struct{}),

		config: config,
	}
}

func (p *MilvusConnectionPool) createNewConnection(params *MilvusParams) (*PooledConnection, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	client, err := milvusclient.New(ctx, p.config)
	if err != nil {
		return nil, err
	}

	exists, err := client.HasCollection(ctx, milvusclient.NewDescribeCollectionOption(params.Collection))
	if err != nil {
		return nil, fmt.Errorf("failed to describe collection: %v", err)
	}
	if !exists {
		err = createCollection(ctx, client, params.Collection, params.useBM25, params.useContextEmbed)
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
func (p *MilvusConnectionPool) GetConnection(options ...MilvusParamsOption) (*PooledConnection, error) {

	params := NewMilvusParams()
	for _, option := range options {
		option(params)
	}

	currentConnection, err := p.getFromCache(params)
	if err != nil {
		return nil, err
	}

	atomic.AddInt32(&currentConnection.refCount, 1)
	currentConnection.lastUsed = time.Now()

	return currentConnection, nil
}

func (p *MilvusConnectionPool) addToCache(hash string, connection *PooledConnection) {
	// 加入连接缓存
	p.connections[hash] = p.lru.PushBack(connection)
	// 注册连接缓存
	p.connHashMap[p.lru.Back()] = connection
	log.Printf("add connection to cache success, params: %s \n", hash)
}

func (p *MilvusConnectionPool) getFromCache(params *MilvusParams) (*PooledConnection, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	var currentConnection *PooledConnection
	hash := params.Hash()
	log.Println("looking for connection", hash)
	
	// 如果连接缓存为空
	if p.lru.Len() == 0 {
		var err error
		log.Printf("milvus connection pool is empty, create new connection \n")
		currentConnection, err = p.createNewConnection(params)
		if err != nil {
			return nil, err
		}
		// 加入连接缓存
		p.addToCache(params.Hash(), currentConnection)
		return currentConnection, nil
	}

	// 缓存不为空，先查找是否已存在
	if c, ok := p.connections[hash]; ok {
		log.Printf("milvus connection pool hit, params: %s \n", hash)
		// 查找到了
		p.lru.MoveToBack(c)
		currentConnection = p.connHashMap[c]
		return currentConnection, nil
	}

	// 没有命中缓存，需要创建新连接
	log.Printf("milvus connection pool miss, params: %s \n", hash)
	
	// 检查是否达到最大连接数
	if p.lru.Len() >= p.maxSize {
		// 尝试清理一个空闲连接
		if !p.cleanOneIdleConnection() {
			return nil, errors.New("连接池已满且无可清理的空闲连接")
		}
	}

	// 创建新连接
	var err error
	currentConnection, err = p.createNewConnection(params)
	if err != nil {
		return nil, err
	}
	// 加入连接缓存
	p.addToCache(hash, currentConnection)

	return currentConnection, nil
}

type PooledConnection struct {
	client   *milvusclient.Client
	params   *MilvusParams
	lastUsed time.Time

	refCount int32 // 引用计数
}
