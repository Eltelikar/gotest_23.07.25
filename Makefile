MAINFILE=./main/main.go

all: swagger build
	docker-compose up

build:
	docker-compose build

clean-build: swagger clear-build build
	docker-compose up

clear-build:
	docker-compose down --rmi all --volumes

swagger: clean-swagger
	swag init --generalInfo $(MAINFILE)

clean-swagger:
	rm -rf docs