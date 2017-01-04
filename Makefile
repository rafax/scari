worker:
	go install ./cmd/scari-worker && SCARI_SERVER="http://localhost:3001/" SCARI_OUTDIR="/tmp/out/" scari-worker

server:
	go install ./cmd/scari-server && scari-server

docker-run: docker-server docker-worker

docker-server: 
	docker run -p 3001:3001 --name scari-server -d scari-server

docker-worker:
	docker run -d --name scari-worker --link=scari-server:scari-server -v /tmp/out:/out -e SCARI_SERVER="http://scari-server:3001/" -e SCARI_OUTDIR="/out/" scari-worker

docker-build: docker-build-worker docker-build-server

docker-build-worker:
	docker build cmd/scari-worker -t=scari-worker

docker-build-server:
	docker build cmd/scari-server -t=scari-server

test:
	go test ./...

install:
	go install ./...

fetch:
	go get -u ./...
