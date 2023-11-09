package main

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"time"
)

func NewConnectorPostgres(jsonConfig map[string]string) (Connector, error) {
	host, _ := jsonConfig["host"]
	port, _ := jsonConfig["port"]
	databaseName, _ := jsonConfig["databaseName"]
	user, _ := jsonConfig["user"]
	password, _ := jsonConfig["password"]
	sslMode, _ := jsonConfig["sslMode"]

	if host == "" || port == "" || databaseName == "" || user == "" || password == "" || sslMode == "" {
		return nil, errors.New("host, port, databaseName, user, password, and sslMode cannot be empty")
	}

	connectionString := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s", host, port, databaseName, user, password, sslMode)

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}

	return ConnectorPostgres{
		Host:             host,
		Port:             port,
		DatabaseName:     databaseName,
		User:             user,
		Password:         password,
		SslMode:          sslMode,
		ConnectionString: connectionString,
		DB:               db,
	}, nil
}

type ConnectorPostgres struct {
	Host             string
	Port             string
	DatabaseName     string
	User             string
	Password         string
	SslMode          string
	ConnectionString string
	DB               *sql.DB
}

func (c ConnectorPostgres) Name() string {
	return "postgres"
}

func (c ConnectorPostgres) Query(input string) (*QueryResults, error) {
	startTime := time.Now().UnixMilli()
	dbRows, err := c.DB.Query(input)
	endTime := time.Now().UnixMilli()

	if err != nil {
		log.Println(fmt.Sprintf(`error for "%v": %v`, input, err))
		return nil, err
	}

	duration := endTime - startTime

	columnTypes, err := dbRows.ColumnTypes()
	if err != nil {
		log.Println(fmt.Sprintf(`error for "%v": %v`, input, err))
		return nil, err
	}

	headers := []Header{}
	for _, columnType := range columnTypes {
		header := Header{
			Name: columnType.Name(),
			Type: columnType.DatabaseTypeName(),
		}

		headers = append(headers, header)
	}

	queryResultRows := [][]string{}
	for dbRows.Next() {
		values := make([]interface{}, len(headers))
		valuePointers := make([]interface{}, len(headers))

		for i := range headers {
			valuePointers[i] = &values[i]
		}

		err = dbRows.Scan(valuePointers...)
		if err != nil {
			return nil, err
		}

		stringValues := []string{}
		for _, v := range values {
			stringValues = append(stringValues, fmt.Sprintf("%v", v))
		}

		queryResultRows = append(queryResultRows, stringValues)
	}

	queryResult := QueryResults{
		Headers:    headers,
		Rows:       queryResultRows,
		DurationMs: duration,
	}

	log.Println(fmt.Sprintf(`"%v" took %v milliseconds`, input, duration))

	return &queryResult, nil
}

func (c ConnectorPostgres) Execute(input string) (*ExecuteResult, error) {
	//TODO implement me
	panic("implement me")
}
