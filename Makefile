.PHONY: start stop logs build loadtest

start:
	docker compose up --build

stop:
	docker compose down --rmi all --volumes --remove-orphans

loadtest:
	docker compose --profile loadtest run --rm loadtest

logs:
	docker compose logs -f

build:
	@mkdir -p bin
	go mod tidy
	go build -o load_balancer cmd/server/main.go