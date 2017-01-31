deploy-server:
	git push heroku master

scari-worker:
	go install ./cmd/scari-worker && GOOGLE_APPLICATION_CREDENTIALS=./cmd/scari-worker/scari-8a1786479a6f.json SCARI_SERVER="http://localhost:3001/" SCARI_OUTDIR="/tmp/out/" scari-worker

scari-server:
	go install ./cmd/scari-server && DATABASE_URL='user=scari dbname=scari sslmode=disable' scari-server

docker-run: scari-server docker-worker

docker-worker:
	docker run -d -v /tmp/out:/out -e SCARI_SERVER="http://localhost:3001/" -e SCARI_OUTDIR="/out/" gdlwcz/scari-worker:latest

docker-push-worker:	docker-build-worker
	docker push gdlwcz/scari-worker:latest

docker-build-worker:
	docker build --no-cache cmd/scari-worker -t=gdlwcz/scari-worker:latest

test:
	go test ./...

acceptance-test:
	(go install ./cmd/scari-server && DATABASE_URL='user=scari dbname=scari sslmode=disable' scari-server &) && (cd acceptance_tests && mix test) && killall scari-server


ci: test acceptance-test