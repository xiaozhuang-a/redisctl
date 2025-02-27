package utils

import (
	"strings"
	"sync"
)

func UnionStrSlice(slice1, slice2 []string) []interface{} {
	unionMap := make(map[interface{}]struct{})

	for _, item := range slice1 {
		unionMap[item] = struct{}{}
	}

	for _, item := range slice2 {
		unionMap[item] = struct{}{}
	}

	var union []interface{}
	for item := range unionMap {
		union = append(union, item)
	}

	return union
}

// Difference slice1 - slice2
func DiffABStrSlice(slice1, slice2 []string) []string {
	differenceMap := make(map[string]struct{})

	for _, item := range slice1 {
		differenceMap[item] = struct{}{}
	}

	for _, item := range slice2 {
		delete(differenceMap, item)
	}

	var difference []string
	for item := range differenceMap {
		difference = append(difference, item)
	}

	return difference
}

func InterStrSlice(slice1, slice2 []string) []string {
	intersectionMap := make(map[string]struct{})
	for _, item := range slice1 {
		intersectionMap[item] = struct{}{}
	}
	var intersection []string
	for _, item := range slice2 {
		if _, exists := intersectionMap[item]; exists {
			intersection = append(intersection, item)
		}
	}

	return intersection
}

func Slice2Map(s []string) map[string]struct{} {
	ma := make(map[string]struct{})
	for _, item := range s {
		ma[item] = struct{}{}
	}

	return ma
}

func SplitString(s, sep string) []string {
	if s == "" {
		return []string{}
	}
	return strings.Split(s, sep)
}

// SafeSlice 是一个并发安全的泛型切片类型，T尽量为值类型，如果是指针类型，需要自己实现深拷贝
type SafeSlice[T any] struct {
	mu    sync.RWMutex
	slice []T
}

// NewSafeSlice 创建一个新的 SafeSlice 实例
func NewSafeSlice[T any]() *SafeSlice[T] {
	return &SafeSlice[T]{
		slice: make([]T, 0),
	}
}

// Append 向切片中添加元素，支持并发写操作
func (s *SafeSlice[T]) Append(item ...T) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.slice = append(s.slice, item...)
}

// Get 根据索引获取元素，支持并发读操作
func (s *SafeSlice[T]) Get(index int) (T, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if index < 0 || index >= len(s.slice) {
		var zeroValue T
		return zeroValue, false
	}

	return s.slice[index], true
}

// Len 返回切片的长度，支持并发读操作
func (s *SafeSlice[T]) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.slice)
}

// Remove 根据索引删除元素，支持并发写操作
func (s *SafeSlice[T]) Remove(index int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if index < 0 || index >= len(s.slice) {
		return false
	}

	s.slice = append(s.slice[:index], s.slice[index+1:]...)
	return true
}

// All 返回切片的所有元素，支持并发读操作,值copy,如果T是指针类型，需要自己实现深拷贝
func (s *SafeSlice[T]) All() []T {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return append([]T(nil), s.slice...)
}
