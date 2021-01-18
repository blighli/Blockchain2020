//@author: hdsfade
//@date: 2021-01-16-22:13
package raft

import "sync"

// 定义persistent state的存储接口
type Storage interface {
	Set(key string, value []byte)
	Get(key string) ([]byte, bool)
	HasData() bool
}

// MapStorage是Storage接口的内存存储实现
type MapStorage struct {
	mu sync.Mutex
	m  map[string][]byte
}

func NewMapStorage() *MapStorage {
	m := make(map[string][]byte)
	return &MapStorage{
		m: m,
	}
}

func (ms *MapStorage) Get(key string) ([]byte, bool) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	v, found := ms.m[key]
	return v, found
}

func (ms *MapStorage) Set(key string, value []byte) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.m[key] = value
}

func (ms *MapStorage) HasData() bool {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	return len(ms.m) > 0
}
