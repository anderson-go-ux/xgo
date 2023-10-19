package xgo

import (
	"fmt"
	"sync"
	"time"
)

var idworker IdWorker

const (
	snow_nodeBits  uint8 = 10
	snow_stepBits  uint8 = 12
	snow_nodeMax   int64 = -1 ^ (-1 << snow_nodeBits)
	snow_stepMax   int64 = -1 ^ (-1 << snow_stepBits)
	snow_timeShift uint8 = snow_nodeBits + snow_stepBits
	snow_nodeShift uint8 = snow_stepBits
)

var snow_epoch int64 = 1514764800000

type snowflake struct {
	mu        sync.Mutex
	timestamp int64
	node      int64
	step      int64
}

type IdWorker interface {
	GetId() int64
}

func (n *snowflake) GetId() int64 {
	n.mu.Lock()
	defer n.mu.Unlock()
	now := time.Now().UnixNano() / 1e6
	if n.timestamp == now {
		n.step++
		if n.step > snow_stepMax {
			for now <= n.timestamp {
				now = time.Now().UnixNano() / 1e6
			}
		}
	} else {
		n.step = 0
	}
	n.timestamp = now
	result := (now-snow_epoch)<<snow_timeShift | (n.node << snow_nodeShift) | (n.step)
	return result
}

func NewIdWorker(node int64) {
	if node < 0 || node > snow_nodeMax {
		panic(fmt.Sprintf("snowflake节点必须在0-%d之间", node))
	}
	snowflakeIns := &snowflake{
		timestamp: 0,
		node:      node,
		step:      0,
	}
	idworker = snowflakeIns
}
