package main

import (
	"sync"
)

type Clients struct {
	mtx sync.Mutex
	m   map[*Client]bool   //общий список клиентов
	w   map[string]*Client //ждущие клиенты
}

func (c *Clients) getClient(id string) *Client {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	if client, ok := c.w[id]; ok {
		return client
	}
	return nil
}

func (c *Clients) wait(client *Client, id string) {
	c.mtx.Lock()
	c.w[id] = client
	c.mtx.Unlock()
}

func (c *Clients) register(client *Client) {
	c.mtx.Lock()
	c.m[client] = true
	c.mtx.Unlock()
}

func (c *Clients) unregister(client *Client) {
	c.mtx.Lock()
	if _, ok := c.m[client]; ok {
		delete(c.m, client)
	}
	if _, ok := c.w[client.id]; ok {
		delete(c.w, client.id)
	}
	c.mtx.Unlock()
}

func (c *Clients) size() int {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	return len(c.m)
}
