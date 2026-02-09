test:
	go test ./...

test-coverage:
	go test -coverprofile c.out ./...
	go tool cover -html c.out -o coverage.html
	go tool cover -func c.out

docker-build:
	docker build -f .ci/Dockerfile -t srd:latest .

run:
	air serve

simple-loop:
	http get 127.0.0.1:9200 host:a.loop.test.srd.sh
	http get 127.0.0.1:9200 host:b.loop.test.srd.sh
