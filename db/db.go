package db

import (
	"database/sql"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
)

var db *sqlx.DB

func InitDB() *sqlx.DB {
	// Загружаем .env файл
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Ошибка загрузки .env файла: %v", err)
	}

	// Чтение переменных из .env
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	dbSslMode := os.Getenv("DB_SSLMODE")

	// Формируем строку подключения
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", dbUser, dbPassword, dbHost, dbPort, dbName, dbSslMode)

	// Подключаемся к базе данных
	var errConnecting error
	db, errConnecting = sqlx.Connect("postgres", dsn)
	if errConnecting != nil {
		log.Fatalf("Не удалось подключиться к БД: %v", errConnecting)
	}

	runMigrations(dsn)

	return db
}

func runMigrations(dsn string) {
	// открываем обычный sql.Conn для мигратора
	sqlDB, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Ошибка при открытии соединения для миграций: %v", err)
	}
	defer sqlDB.Close()

	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		log.Fatalf("Ошибка создания драйвера миграции: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://db/sql", // работает при запуске из project-root
		"postgres", driver,
	)
	if err != nil {
		log.Fatalf("Ошибка инициализации миграций: %v", err)
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Ошибка применения миграций: %v", err)
	}
	fmt.Println("Миграции успешно применены.")
}
