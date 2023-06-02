package database

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/url"
	"time"

	_ "github.com/ClickHouse/clickhouse-go"
	"github.com/elastic/go-elasticsearch/v8"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gocql/gocql"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PostgresDB struct {
	Host     string
	Port     string
	Username string
	Password string
	DBName   string
	SSLMode  string
}

type MongoDB struct {
	Host     string
	Port     string
	Username string
	Password string
}

type ElasticSearchDB struct {
	Host     string
	Username string
	Password string
	Index    string
}

type CassandraDB struct {
	Host     string
	Keyspace string
	Username string
	Password string
}

type MariaDB struct {
	Host     string
	Port     string
	Username string
	Password string
	DBName   string
}

type ClickHouseDB struct {
	Host     string
	Port     string
	Username string
	Password string
	DBName   string
}

type CockroachDB struct {
	Host     string
	Port     string
	Username string
	Password string
	DBName   string
}

func NewPostgresDB(database PostgresDB) (*sql.DB, error) {
	db, err := sql.Open("postgres", fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		database.Username, database.Password, database.Host, database.Port, database.DBName, database.SSLMode))
	if err != nil {

		return nil, fmt.Errorf("error connecting to database:%s", err)
	}
	err = db.Ping()
	if err != nil {
		logrus.Errorf("DB ping error:%s", err)
		return nil, err
	}
	_, err = db.Exec(TODO_SCHEMA)
	if err != nil {
		logrus.Errorf("Error executing initial migration:%s", err)
		return nil, fmt.Errorf("error executing initial migration:%s", err)
	}
	_, err = db.Exec(USER_SCHEMA)
	if err != nil {
		logrus.Errorf("Error executing initial migration:%s", err)
		return nil, fmt.Errorf("error executing initial migration:%s", err)
	}
	return db, nil
}

const TODO_SCHEMA = `
	CREATE TABLE IF NOT EXISTS todos (
		id serial not null primary key,
		title varchar(225) NOT NULL,
		done bool DEFAULT FALSE
	);
`
const USER_SCHEMA = `
CREATE TABLE IF NOT EXISTS users
(
    id serial not null primary key,
    name varchar(225) not null,
    email varchar(225) not null UNIQUE,
    phone varchar(225) not null UNIQUE,
    password varchar(225) not null,
    role varchar(225) not null default 'USER',
    CONSTRAINT proper_email CHECK (email ~* '^[A-Za-z0-9._+%-]+@[A-Za-z0-9.-]+[.][A-Za-z]+$')
);
`

func ConnectToMongo(database MongoDB) (*mongo.Client, error) {
	mongoURI := fmt.Sprintf("mongodb://%s:%s@%s:%s",
		database.Username,
		url.QueryEscape(database.Password),
		database.Host,
		database.Port,
	)

	client, err := mongo.NewClient(options.Client().ApplyURI(mongoURI))
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func NewElasticSearchDB(database ElasticSearchDB) (*elasticsearch.Client, error) {

	cfg := elasticsearch.Config{
		Addresses: []string{database.Host},
		Username:  database.Username,
		Password:  database.Password,
	}
	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	_, err = client.Ping(client.Ping.WithContext(context.Background()))
	if err != nil {
		return nil, err
	}

	res, err := client.Indices.Exists([]string{database.Index})
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		createRes, err := client.Indices.Create(database.Index)
		if err != nil {
			return nil, err
		}
		defer createRes.Body.Close()

		if createRes.IsError() {
			return nil, fmt.Errorf("failed to create index: %s", createRes.String())
		}
	}

	return client, nil
}

func ConnectToCassandra(database CassandraDB) (*gocql.Session, error) {
	cluster := gocql.NewCluster(database.Host)
	cluster.Consistency = gocql.Quorum
	cluster.Keyspace = database.Keyspace
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: database.Username,
		Password: database.Password,
	}
	session, err := cluster.CreateSession()
	if err != nil {
		return nil, err
	}

	return session, nil
}

func NewMariaDB(database MariaDB) (*sql.DB, error) {
	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", database.Username, database.Password, database.Host, database.Port, database.DBName)
	db, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %s", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("error pinging database: %s", err)
	}

	_, err = db.Exec(TODO_SCHEMA_MariaDB)
	if err != nil {
		return nil, fmt.Errorf("error executing initial migration: %s", err)
	}

	return db, nil
}

const TODO_SCHEMA_MariaDB = `
	CREATE TABLE IF NOT EXISTS todos (
		id INT AUTO_INCREMENT PRIMARY KEY,
		title VARCHAR(225) NOT NULL,
		completed BOOLEAN DEFAULT FALSE
	);
`

func NewClickHouseDB(database ClickHouseDB) (*sql.DB, error) {
	connect, err := sql.Open("clickhouse", fmt.Sprintf("tcp://%s:%s?username=%s&password=%s&database=%s", database.Host, database.Port, database.Username, database.Password, database.DBName))
	if err != nil {
		return nil, fmt.Errorf("error connecting to ClickHouse: %s", err)
	}
	err = connect.Ping()
	if err != nil {
		return nil, fmt.Errorf("ClickHouse ping error: %s", err)
	}
	return connect, nil
}

func NewCockroachDB(database CockroachDB) (*sql.DB, error) {
	connString := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=require",
		database.Username, database.Password, database.Host, database.Port, database.DBName)

	db, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %s", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("error pinging database: %s", err)
	}

	_, err = db.Exec(TODO_SCHEMA_CockroachDB)
	if err != nil {
		return nil, fmt.Errorf("error executing initial migration: %s", err)
	}

	return db, nil
}

const TODO_SCHEMA_CockroachDB = `
	CREATE TABLE IF NOT EXISTS todos (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		title VARCHAR(225) NOT NULL,
		completed BOOL DEFAULT FALSE
	);
`
