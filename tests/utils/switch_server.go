package utils

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
)

type SwitchServer struct {
	Server *httptest.Server
	down   int32 // atomic: 0 — work, 1 — fail
}

func NewSwitchServer() *SwitchServer {
	ss := &SwitchServer{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.LoadInt32(&ss.down) == 1 {
			http.Error(w, "fail", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	ss.Server = httptest.NewServer(handler)
	return ss
}

func (ss *SwitchServer) SetDown(down bool) {
	if down {
		atomic.StoreInt32(&ss.down, 1)
	} else {
		atomic.StoreInt32(&ss.down, 0)
	}
}

func (ss *SwitchServer) URL() string {
	return ss.Server.URL
}

func (ss *SwitchServer) Close() {
	ss.Server.Close()
}
