build:
	go build -o ./.bin/todo ./cmd/main.go

runs: build
	./.bin/todo

redis:
     docker run -d --name redis-container -p 6379:6379 -v /database/red:/data -e REDIS_PASSWORD=mypassword redis redis-server --requirepass mypassword

mongo:
    docker run -d --name mongo-container -p 27019:27017 -v ./mongo_init.js:/docker-entrypoint-initdb.d/mongo_init.js -v /database/dbdata:/data/db -e MONGO_INITDB_ROOT_USERNAME=root -e MONGO_INITDB_ROOT_PASSWORD=qwerty mongo

cassandra:
    docker run -d --name cassandra-container -p 9042:9042 -e CASSANDRA_CLUSTER_NAME=MyCluster -e CASSANDRA_LISTEN_ADDRESS=auto or (127.0.0.1) -e CASSANDRA_KEYSPACE=todo -e CASSANDRA_USERNAME=cass -e CASSANDRA_PASSWORD=testpassword -v /database/cass:/var/lib/cassandra cassandra:4.0.9
maria:
    docker run -d --name mariadb-container -p 3306:3306 -e MYSQL_ROOT_PASSWORD=password123 -e MYSQL_USER=mari -e MYSQL_PASSWORD=password123 -e MYSQL_DATABASE=mydatabase -v database/maria:/var/lib/mysql mariadb:latest
clickhouse:
    docker run -d --name clickhouse-container -p 9000:9000 -p 8123:8123 -e CLICKHOUSE_USER=click -e CLICKHOUSE_PASSWORD=password345 -v /database/click:/var/lib/clickhouse yandex/clickhouse-server

lint:
	golangci-lint run

prometeus:
    docker run -d --name prometheus-container -p 9090:9090 -v data/prometheus.yml:/etc/prometheus/prometheus.yml -v prometheus-data:/prometheus prom/prometheus:latest --config.file=/etc/prometheus/prometheus.yml --storage.tsdb.path=/prometheus

grafana:
    docker run -d -p 3000:3000 --name grafana-container  grafana/grafana:latest

rabbitmq:
    docker run -d -p 5672:5672 -p 15672:15672 --name rabbitmq-server -e RABBITMQ_DEFAULT_USER=username -e RABBITMQ_DEFAULT_PASS=password -v /database/rabbit:/var/lib/rabbitmq rabbitmq:3-management

build-image:
	docker build -t service_todo:v1 .

start-container:
	docker run --name service-todo-api -p 8080:8080 --env-file .env todo:v1

run:
	go run cmd/main.go


