deploy-server:
	git push heroku master

scari-worker:
	go install ./cmd/scari-worker && GOOGLE_APPLICATION_CREDENTIALS=./cmd/scari-worker/scari-8a1786479a6f.json SCARI_SERVER="http://localhost:3001/" SCARI_OUTDIR="/tmp/out/" scari-worker

scari-server:
	go install ./cmd/scari-server && HTTP_LOG=truescari-server

docker-run: scari-server docker-worker

docker-worker:
	docker run -d -v /tmp/out:/out -e SCARI_SERVER="http://localhost:3001/" -e SCARI_OUTDIR="/out/" gdlwcz/scari-worker:latest

docker-push-worker:	docker-build-worker
	docker push gdlwcz/scari-worker:latest

docker-build-worker:
	docker build --no-cache cmd/scari-worker -t=gdlwcz/scari-worker:latest

docker-build-server:
	(cd cmd/scari-server && GOOS=linux go build -v && docker build .)

test:
	go test $(go list ./... | grep -v vendor)

acceptance-test:
	(go install ./cmd/scari-server && scari-server &) && (cd acceptance_tests && mix test) && killall scari-server


ci: test acceptance-test