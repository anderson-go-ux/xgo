package xgo

import (
	"fmt"
	"sync"
	"time"
)

var snow_worker snowworker

type snowflake struct {
	mu        sync.Mutex
	timestamp int64
	node      int64
	step      int64
}

type snowworker interface {
	GetId() int64
}

func (n *snowflake) GetId() int64 {
	n.mu.Lock()
	defer n.mu.Unlock()
	now := time.Now().UnixNano() / 1e6
	if n.timestamp == now {
		n.step++
		if n.step > -1 ^ (-1 << 12) {
			for now <= n.timestamp {
				now = time.Now().UnixNano() / 1e6
			}
		}
	} else {
		n.step = 0
	}
	n.timestamp = now
	result := (now-1514764800000)<<10 + 12 | (n.node << 12) | (n.step)
	return result
}

func newIdWorker(node int) {
	if node < 0 || node > -1 ^ (-1 << 10) {
		panic(fmt.Sprintf("snowflake节点必须在0-%d之间", node))
	}
	snowflakeIns := &snowflake{
		timestamp: 0,
		node:      int64(node),
		step:      0,
	}
	snow_worker = snowflakeIns
}
