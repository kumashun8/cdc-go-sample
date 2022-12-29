up:
	docker-compose up -d

down:
	docker-compose down

exec:
	docker-compose exec db bash

setup:
	go run ./dbsetup

pg.login:
	psql -h localhost -p 5432 -U postgres -d postgres
