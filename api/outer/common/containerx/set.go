package containerx

import (
	"sync"
)

// Set 不重复集合
type Set struct {
	m map[string]bool
	sync.RWMutex
}

//New 初始化
func New() *Set {
	return &Set{
		m: map[string]bool{},
	}
}

// Add 添加
func (s *Set) Add(item string) {
	s.Lock()
	defer s.Unlock()
	s.m[item] = true
}

// AddAll 添加多个元素
func (s *Set) AddAll(items ...string) {
	s.Lock()
	defer s.Unlock()
	for _, item := range items {
		s.m[item] = true
	}
}

// Remove 删除
func (s *Set) Remove(item string) {
	s.Lock()
	defer s.Unlock()
	delete(s.m, item)
}

// Clean 清空
func (s *Set) Clean() {
	s.m = map[string]bool{}
}

// Len 长度
func (s *Set) Len() int {
	return len(s.m)
}

// Contains 包含
func (s *Set) Contains(item string) bool {
	_, ok := s.m[item]
	return ok
}

// IsEmpty 是否为空
func (s *Set) IsEmpty() bool {
	return len(s.m) > 0
}

// ToList 变成数组
func (s *Set) ToList() (l []string) {
	var list []string
	for item := range s.m {
		list = append(list, item)
	}

	return list
}
