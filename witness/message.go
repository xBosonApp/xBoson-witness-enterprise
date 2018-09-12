package witness

import (
	"sync"
)

const (
	LOG_STACK_LEN = 99
)

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


func (m *Message) Read() ([][]interface{}) {
	m.log_safe.Lock()
	defer m.log_safe.Unlock()

	if m.log_read_count < m.log_write_count {
		var ret = make([][]interface{}, 0)

		for i:=1; i<LOG_STACK_LEN; i++ {
			p := m.log_point - i
			if p < 0 {
				p = LOG_STACK_LEN + p
			}
			if m.log_stack[p] == nil {
				break;
			}
			if m.log_read_count <= m.log_stack[p].id {
				ret = append(ret, m.log_stack[p].log)
				m.log_stack[p] = nil
			}
		}
		m.log_read_count = m.log_write_count
		return ret;
	}
	return nil
}