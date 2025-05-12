.PHONY: start down logs build

start:
	docker compose up --build

stop:
	docker compose down --rmi all --volumes --remove-orphans

logs:
	docker compose logs -f

build:
	CGO_ENABLED=0 GOOS=linux go build -o load_balancer cmd/server/main.go