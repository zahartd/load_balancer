.PHONY: start stop logs build loadtest_ab loadtest_vegeta tests

start:
	docker compose up --build

stop:
	docker compose down --rmi all --volumes --remove-orphans

loadtest_ab:
	docker compose --profile loadtest-ab run --rm loadtest-ab

loadtest_vegeta:
	docker compose --profile loadtest-vegeta run --rm loadtest-vegeta
	@echo "Ready! Report in the Vegeta/Reports folder: see plot.html"

tests:
	go test -race ./...

logs:
	docker compose logs -f

build:
	@mkdir -p bin
	go mod tidy
	go build -o load_balancer cmd/server/main.go