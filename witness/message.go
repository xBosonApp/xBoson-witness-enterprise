package witness

import (
	"sync"
)

const (
	LOG_STACK_LEN = 99
)

//
// 消息缓冲区队列, 当队列被填满, 旧的消息将被覆盖(删除)
//
type Message struct {
	log_stack [LOG_STACK_LEN]*StackItem
	log_safe sync.Mutex
	log_point int
	log_write_count int
	log_read_count int
}

type StackItem struct {
	id  int
	log []interface{}
}


//
// 发送消息
//
func (m *Message) Send(a ...interface{}) {
	m.log_safe.Lock()
	defer m.log_safe.Unlock()

	m.log_stack[m.log_point] = &StackItem{ m.log_write_count, a }
	m.log_write_count++
	m.log_point++
	if m.log_point >= LOG_STACK_LEN {
		m.log_point = 0;
	}
}


//
// 读取最新的消息, 没有新消息返回 nil
// 多个节点请求消息将产生竞争, 调用者自己记录读取计数器
// 当 log_read_count <= 0 将使用默认计数器作为读取计数器当前值.
// 返回未读取的所有消息, 并返回一个读取计数器值用于下一次读取.
//
func (m *Message) Read(log_read_count int) ([][]interface{}, int) {
	m.log_safe.Lock()
	defer m.log_safe.Unlock()

	if (log_read_count <= 0) {
		log_read_count = m.log_read_count
	}

	if log_read_count < m.log_write_count {
		var ret = make([][]interface{}, 0)

		for i:=1; i<LOG_STACK_LEN; i++ {
			p := m.log_point - i
			if p < 0 {
				p = LOG_STACK_LEN + p
			}
			if m.log_stack[p] == nil {
				break;
			}
			if log_read_count <= m.log_stack[p].id {
				ret = append(ret, m.log_stack[p].log)
				// m.log_stack[p] = nil
			}
		}
		m.log_read_count = m.log_write_count
		return ret, m.log_write_count
	}
	return nil, 0
}