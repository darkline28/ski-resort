-include .env
# Переменные которые будут использоваться в наших командах
DB_DSN := "postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable"
MIGRATE := migrate -path ./backend/migrations -database $(DB_DSN)

# Таргет для создания новой миграции
migrate-new:
	migrate create -ext sql -dir ./backend/migrations -seq ${NAME}

# Применение миграций
migrate:
	$(MIGRATE) up

# Откат миграций
migrate-down:
	$(MIGRATE) down
	
# для удобства добавим команду run, которая будет запускать наше приложение
run:
	cd backend && go run ./cmd/app

gen:
	cd backend && oapi-codegen \
		-config api/oapi-codegen.yaml \
		-include-tags auth \
		api/openapi.yaml > internal/auth/api/api.gen.go
	
gen-category:
	cd backend && oapi-codegen \
		-config api/oapi-codegen.yaml \
		-include-tags Equipment\
		api/openapi.yaml > internal/catalog/api/api.gen.go


gen-instructors:
	cd backend && oapi-codegen \
		-config api/oapi-codegen.yaml \
		-include-tags Instructor \
		api/openapi.yaml > internal/instructors/api/api.gen.go

gen-services:
	cd backend && oapi-codegen \
		-config api/oapi-codegen.yaml \
		-include-tags services \
		api/openapi.yaml > internal/services/api/api.gen.go

gen-bookings:
	cd backend && oapi-codegen \
		-config api/oapi-codegen.yaml \
		-include-tags bookings \
		api/openapi.yaml > internal/bookings/api/api.gen.go

gen-rental:
	cd backend && oapi-codegen \
		-config api/oapi-codegen.yaml \
		-include-tags rental \
		api/openapi.yaml > internal/rental/api/api.gen.go

gen-admin:
	cd backend && oapi-codegen \
		-config api/oapi-codegen.yaml \
		-include-tags Admin \
		api/openapi.yaml > internal/admin/api/api.gen.go