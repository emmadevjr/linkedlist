package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type Node struct {
	Value int
	Next  *Node
}

type LinkedList struct {
	head   *Node
	length uint
	mutex  sync.Mutex
}

func NewLinkedList() *LinkedList {
	return &LinkedList{}
}

func (l *LinkedList) Find(n int) (index uint, found bool) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	current := l.head
	index = 0
	for current != nil {
		if current.Value == n {
			return index, true
		}
		current = current.Next
		index++
	}
	return 0, false
}

func (l *LinkedList) Get(index uint) (int, bool) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	current := l.head
	for i := uint(0); i < index; i++ {
		if current == nil {
			return 0, false
		}
		current = current.Next
	}

	if current == nil {
		return 0, false
	}

	return current.Value, true
}

func (l *LinkedList) Insert(index uint, val int) bool {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if index > l.length {
		return false
	}

	newNode := &Node{Value: val}

	if index == 0 {
		newNode.Next = l.head
		l.head = newNode
		l.length++
		return true
	}

	current := l.head
	for i := uint(0); i < index-1; i++ {
		if current == nil {
			return false
		}
		current = current.Next
	}

	if current.Next != nil {
		newNode.Next = current.Next.Next
	} else {
		newNode.Next = nil
	}
	current.Next = newNode
	l.length++

	return true
}

func (l *LinkedList) Remove(index uint) bool {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if index >= l.length {
		return false
	}

	if index == 0 {
		l.head = l.head.Next
		l.length--
		return true
	}

	current := l.head
	for i := uint(0); i < index-1; i++ {
		if current.Next == nil {
			return false
		}
		current = current.Next
	}

	if current.Next == nil {
		return false
	}

	current.Next = current.Next.Next
	l.length--
	return true
}

func handleInsert(w http.ResponseWriter, r *http.Request, list *LinkedList) {
	var req struct {
		Index uint `json:"index"`
		Value int  `json:"value"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	success := list.Insert(req.Index, req.Value)
	if !success {
		http.Error(w, "Insert failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Insert successful"})
}

func handleGet(w http.ResponseWriter, r *http.Request, list *LinkedList) {
	indexStr := strings.TrimPrefix(r.URL.Path, "/get/")
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		http.Error(w, "Invalid index", http.StatusBadRequest)
		return
	}

	value, found := list.Get(uint(index))
	if !found {
		http.Error(w, "Index out of range", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"value": value})
}

func handleRemove(w http.ResponseWriter, r *http.Request, list *LinkedList) {
	indexStr := strings.TrimPrefix(r.URL.Path, "/remove/")
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		http.Error(w, "Invalid index", http.StatusBadRequest)
		return
	}

	success := list.Remove(uint(index))
	if !success {
		http.Error(w, "Remove failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Remove successful"})
}

func handleFind(w http.ResponseWriter, r *http.Request, list *LinkedList) {
	valueStr := strings.TrimPrefix(r.URL.Path, "/find/")
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		http.Error(w, "Invalid value", http.StatusBadRequest)
		return
	}

	index, found := list.Find(value)
	if !found {
		http.Error(w, "Value not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]uint{"index": index})
}

func handleList(w http.ResponseWriter, _ *http.Request, list *LinkedList) {
	current := list.head
	var values []int
	for current != nil {
		values = append(values, current.Value)
		current = current.Next
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(values)
}

func main() {
	list := NewLinkedList()
	http.HandleFunc("/insert", func(w http.ResponseWriter, r *http.Request) {
		handleInsert(w, r, list)
	})
	http.HandleFunc("/get/", func(w http.ResponseWriter, r *http.Request) {
		handleGet(w, r, list)
	})
	http.HandleFunc("/remove/", func(w http.ResponseWriter, r *http.Request) {
		handleRemove(w, r, list)
	})
	http.HandleFunc("/find/", func(w http.ResponseWriter, r *http.Request) {
		handleFind(w, r, list)
	})
	http.HandleFunc("/list", func(w http.ResponseWriter, r *http.Request) {
		handleList(w, r, list)
	})

	fmt.Println("Server started at :8080")
	http.ListenAndServe(":8080", nil)
}
