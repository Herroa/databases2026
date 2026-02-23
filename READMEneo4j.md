
QUICK START

cd docker
docker compose up -d neo4j

wait until up 10 seconds and:
```
go run ../cmd/main.go -neo4j-test
```

render in browser:
```
http://localhost:7474/browser/
```