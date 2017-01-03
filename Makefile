run: server worker

server: 
	docker run -p 3001:3001 --name scari-server -d scari-server

worker:
	docker run -d --name scari-worker --link=scari-server:scari-server -v /tmp/out:/out -e SCARI_SERVER="http://scari-server:3001/" scari-worker

build: build-worker build-server

build-worker:
	docker build cmd/scari-worker -t=scari-worker

build-server:
	docker build cmd/scari-server -t=scari-server

test:
	go test ./...

install:
	go install ./...

fetch:
	go get -u ./...
