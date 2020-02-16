package main

import (
	"log"
	"strings"
)

type List struct {
	data map[string]struct{}
}

func NewList() *List {
	return &List{
		data: make(map[string]struct{}),
	}
}

func (b *List) Add(server string) bool {
	server = strings.Trim(server, " ")
	if len(server) == 0 {
		return false
	}

	if !strings.HasSuffix(server, ".") {
		server += "."
	}
	b.data[server] = struct{}{}

	return true
}

func (b *List) AddList(servers []string) (count int) {
	for _, server := range servers {
		if b.Add(server) {
			count++
		}
	}

	return
}

func (b *List) Contains(server string) bool {
	_, ok := b.data[server]
	return ok
}

func UpdateList(configList []string) *List {
	list := NewList()
	cnt := list.AddList(configList)
	log.Println("[list] Loaded", cnt, "servers:", configList)
	return list
}
