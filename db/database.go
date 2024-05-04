package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // Driver de conexión con Postgres
)

type DbData struct {
	Host     string
	Port     string
	DbName   string
	Username string
	Password string
}

func LoadEnv() (DbData, error) {
	err := godotenv.Load(".env")
	if err != nil && !os.IsNotExist(err) {
		return DbData{}, fmt.Errorf("error loading .env file: %w", err)
	}
	host, exists := os.LookupEnv("DB_HOST")
	if !exists {
		return DbData{}, errors.New("DB_HOST not found in environment variables")
	}

	port, exists := os.LookupEnv("DB_PORT")
	if !exists {
		return DbData{}, errors.New("DB_PORT not found in environment variables")
	}

	dbName, exists := os.LookupEnv("DB_NAME")
	if !exists {
		return DbData{}, errors.New("DB_NAME not found in environment variables")
	}

	username, exists := os.LookupEnv("DB_USER")
	if !exists {
		return DbData{}, errors.New("DB_USER not found in environment variables")
	}

	password, exists := os.LookupEnv("DB_PASSWORD")
	if !exists {
		return DbData{}, errors.New("DB_PASSWORD not found in environment variables")
	}

	return DbData{
		Host:     host,
		Port:     port,
		DbName:   dbName,
		Username: username,
		Password: password,
	}, nil
}

var (
	once sync.Once
	db   *sql.DB
)

func EstablishDbConnection() (*sql.DB, error) {
	var err error

	dbData, err := LoadEnv()
	if err != nil {
		return nil, err
	}

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbData.Host, dbData.Port, dbData.Username, dbData.Password, dbData.DbName)

	//Singleton, establish pool db connection once
	once.Do(func() {
		db, err = sql.Open("postgres", psqlInfo)
		if err != nil {
			log.Fatal("error opening database connection: %w", err)
		}

	})

	if err = db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("error ping database: %w", err)
	}

	fmt.Println("Conexión exitosa a la base de datos")

	return db, nil
}
