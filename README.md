# HTTP-балансировщика нагрузки

Простой балансировщик нагрузки, который принимает входящие HTTP-запросы и распределяет их по пулу бэкенд-серверов + Rate limiter.

## Сборка и заупск

Все компоненты системы запускаются и собираются в docker контейнера (см. docker-compose.yaml и Dockerfile). Для удобства написан Makefile скрипт для быстрого запуска:

```bash
# Сборка и запуск балансировщика нагрузки + 3 тестовых бекендов
make start

# Остановка системы + удаление контейнеров, без удаления можно отправить сигнал через Ctrl+C или другим способом
make stop

# Запуск тестов на Golang юнит тесты и интеграционные (см. tests)
make tests

# Посмотреть логи контейнеров
make logs

# Нагрузочные тесты Apache Bench (В схеме с API-TOKEN это запросы от одного клиента)
make loadtest_ab

# Нагрузочные тесты Vegeta (10 разных клиентов: 80% 200 OK и 20 % 400, но это можно настроить в vegeta/targets.txt)
make loadtest_vegeta

# Локальная сборка бинаря
make build
```

Также можно запустить тесты локально:

```bash
# Go-тесты
go test -race ./...

# Нагрузочные тесты Apache Bench (см. доку AB)
ab -H "X-API-Key: client1" -n 5000 -c 1000 "http://localhost:8081/"

# Нагрузочные тесты Vegeta (см. доку https://github.com/tsenart/vegeta).
vegeta attack -targets="./vegeta/local.txt" -rate=1000 -duration=10s \
  | tee ./vegeta/reports/results.bin \
  | vegeta report --type=text
vegeta plot vegeta/reports/results.bin > vegeta/reports/plot.html
```

## Архитектруа решения

```bash
.
├── Dockerfile
├── Makefile
├── README.md
├── TASK.md
├── cmd
│   └── server
│       └── main.go
├── configs
│   └── config.json
├── docker-compose.yml
├── go.mod
├── go.sum
├── internal
│   ├── backend
│   │   └── backend.go
│   ├── balancer
│   │   ├── algorithms
│   │   │   ├── roundrobin.go
│   │   │   └── roundrobin_test.go
│   │   ├── balancer.go
│   │   ├── balancer_factory.go
│   │   └── manager.go
│   ├── client
│   │   └── client.go
│   ├── config
│   │   └── config.go
│   ├── gateways
│   │   └── http
│   │       ├── middleware.go
│   │       ├── proxy.go
│   │       └── server.go
│   └── ratelimit
│       ├── algorithms
│       │   ├── tokenbucket.go
│       │   └── tokenbucket_test.go
│       ├── limiter.go
│       ├── limiter_factory.go
│       └── manager.go
├── nginx
│   └── echo.conf
├── tests
│   ├── balancer_test.go
│   ├── limiter_test.go
│   └── utils
│       └── switch_server.go
└── vegeta
    ├── reports
    └── targets.txt
```

