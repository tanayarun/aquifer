package aquifer

import "time"

func (p *Pool[C]) reaper() {
	ticker := time.NewTicker(p.cfg.idleTimeout / 2)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.evictExpired()
		case <-p.done:
			return
		}
	}
}

func (p *Pool[C]) evictExpired() {
	n := int64(len(p.idle))

	for range n {
		select {
		case conn := <-p.idle:
			if conn.isExpired(p.cfg.idleTimeout) && p.open.Load() > int64(p.cfg.minConns) {
				conn.conn.Close()
				p.open.Add(-1)
			} else {
				p.idle <- conn
			}
		default:
			return
		}
	}
}
