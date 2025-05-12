package integration_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/zahartd/load_balancer/internal/balancer"
	"github.com/zahartd/load_balancer/tests/utils"
)

func TestBalancer_FailsOverWhenBackendDown(t *testing.T) {
	healthy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("healthy"))
	}))
	defer healthy.Close()

	ss := utils.NewSwitchServer()
	defer fs.Close()

	backends := []string{healthy.URL, ss.URL()}
	lb := balancer.New(backends, "round_robin")

	// 3. Запускаем health-check с интервалом 1 сек
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go health.Watch(lb, time.Second)

	// 4. HTTP-сервер балансировщика на случайном порту
	proxy := server.NewProxy(lb)
	srv := httptest.NewServer(proxy)
	defer srv.Close()

	// 5. Проверяем, что оба бэка дают 200
	resp1, err := http.Get(srv.URL + "/")
	require.NoError(t, err)
	b1, _ := io.ReadAll(resp1.Body)
	resp1.Body.Close()
	require.Equal(t, "healthy", string(b1))

	resp2, err := http.Get(srv.URL + "/")
	require.NoError(t, err)
	b2, _ := io.ReadAll(resp2.Body)
	resp2.Body.Close()
	require.Equal(t, "ok", string(b2))

	// 6. «Отключаем» второй сервер на 10 секунд
	ss.SetDown(true)
	time.Sleep(100 * time.Millisecond) // дождёмся health-check

	// 7. Все запросы теперь должны идти на «healthy»
	for i := 0; i < 5; i++ {
		resp, err := http.Get(srv.URL + "/")
		require.NoError(t, err)
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		require.Equal(t, "healthy", string(body))
	}

	// 8. Включаем второй назад
	ss.SetDown(false)
	time.Sleep(100 * time.Millisecond) // ждать health-check

	// 9. Теперь round-robin снова чередует
	got := map[string]bool{}
	for i := 0; i < 4; i++ {
		resp, err := http.Get(srv.URL + "/")
		require.NoError(t, err)
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		got[string(body)] = true
	}
	require.Contains(t, got, "healthy")
	require.Contains(t, got, "ok")
}
