package db

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
	"time"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
)

func InitDB(ctx context.Context) (*sqlx.DB, *pgx.Conn, *mongo.Database) {
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
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		dbUser, dbPassword, dbHost, dbPort, dbName, dbSslMode)

	// Подключаемся через sqlx
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatalf("Не удалось подключиться к БД через sqlx: %v", err)
	}

	// Выполняем миграции, если нужно
	runMigrations(dsn)

	// Подключаемся через pgx
	pgxConn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		log.Fatalf("Не удалось подключиться к БД через pgx: %v", err)
	}

	mongoUser := os.Getenv("MONGO_USER")
	mongoPass := os.Getenv("MONGO_PASS")
	mongoHost := os.Getenv("MONGO_HOST")
	mongoPort := os.Getenv("MONGO_PORT")
	dbName = os.Getenv("MONGO_DB")

	uri := fmt.Sprintf(
		"mongodb://%s:%s@%s:%s/%s?authSource=%s&replicaSet=rs0",
		mongoUser, mongoPass, mongoHost, mongoPort, dbName, dbName,
	)

	clientOpts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		log.Fatalf("Ошибка подключения к MongoDB: %v", err)
	}

	// Проверка соединения
	ctxPing, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := client.Ping(ctxPing, nil); err != nil {
		log.Fatalf("MongoDB не отвечает: %v", err)
	}

	runMongoMigrations(ctx, client.Database(dbName))

	return db, pgxConn, client.Database(dbName)
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
	defer m.Close()

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Ошибка применения миграций: %v", err)
	}
	fmt.Println("Миграции успешно применены.")
}
func runMongoMigrations(ctx context.Context, db *mongo.Database) {
	// Пример: создаём коллекцию "messages" с индексом по полю "ID" (уникальный)
	collectionNames, err := db.ListCollectionNames(ctx, bson.D{{Key: "name", Value: "messages"}})
	if err != nil {
		log.Fatalf("Ошибка получения списка коллекций: %v", err)
	}

	if len(collectionNames) == 0 {
		// Создаём коллекцию явно, если нужно (необязательно)
		err := db.CreateCollection(ctx, "messages")
		if err != nil {
			log.Fatalf("Ошибка создания коллекции messages: %v", err)
		}
		fmt.Println("Создана коллекция messages")
	}
}
