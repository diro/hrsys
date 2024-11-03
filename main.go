package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	_ "github.com/lib/pq" // 改用 PostgreSQL 驱动
)

const (
	dbHost     = "lifeplan-instance-1.c1w4eg0mwku5.us-west-2.rds.amazonaws.com"
	dbPort     = "5432"
	secretName = "rds!cluster-6de70cb9-03ec-43aa-b93c-06f9ea68965d"
	region     = "us-west-2"
)

func fetchDBCredentials(secretName, region string) (username, password, dbName string, err error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region)},
	)
	if err != nil {
		log.Println("Error creating AWS session:", err)
		return "", "", "", err
	}

	svc := secretsmanager.New(sess)
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	}

	result, err := svc.GetSecretValue(input)
	if err != nil {
		log.Println("Error retrieving secret value:", err)
		return "", "", "", err
	}

	var secretString string
	if result.SecretString != nil {
		secretString = *result.SecretString
	}

	var secretMap map[string]string
	err = json.Unmarshal([]byte(secretString), &secretMap)
	if err != nil {
		log.Println("Error unmarshalling secret string:", err)
		return "", "", "", err
	}

	return secretMap["username"], secretMap["password"], secretMap["dbname"], nil
}

func createTable(db *sql.DB) {
	var err error
	//select db
	_, err = db.Exec("SELECT * FROM lifeplan")
	if err != nil {
		log.Println("Error selecting database:", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS learning (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		year INTEGER NOT NULL,
		bigthing TEXT NOT NULL,
		learned TEXT NOT NULL
	)`)
	if err != nil {
		log.Println("Error creating table:", err)
	}
}

func createDB(db *sql.DB) {
	// 检查数据库是否存在
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = 'lifeplan')").Scan(&exists)
	if err != nil {
		log.Println("Error checking if database exists:", err)
		return
	}

	if !exists {
		_, err := db.Exec("CREATE DATABASE lifeplan")
		if err != nil {
			log.Println("Error creating database:", err)
		}
	}
}

func main() {
	// Fetch database credentials from Secrets Manager
	dbUser, dbPassword, dbName, err := fetchDBCredentials(secretName, region)
	if err != nil {
		log.Fatalf("Unable to fetch database credentials: %v", err)
	}

	// Construct connection string
	dbURI := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=require",
		dbUser, url.QueryEscape(dbPassword), dbHost, dbPort, dbName)

	db, err := sql.Open("postgres", dbURI)
	if err != nil {
		log.Println("Error opening database connection:", err)
	}
	createDB(db)
	createTable(db)
	insertMockData(db)

	defer db.Close()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello, World!")
		fmt.Fprintf(w, "dbUser: %s", dbUser)
	})

	http.HandleFunc("/dbinfo", func(w http.ResponseWriter, r *http.Request) {
		displayDBInfo(db, w, dbName)
		displayTableInfo(db, 1, w)
	})

	log.Println("Server is running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func displayDBInfo(db *sql.DB, w http.ResponseWriter, dbName string) {
	err := db.Ping()
	if err != nil {
		http.Error(w, "Database connection failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	stats := db.Stats()

	info := fmt.Sprintf(`Database connection info:
			Host: %s
			Port: %s
			Database name: %s
			Open connections: %d
			In use connections: %d
			Idle connections: %d`,
		dbHost, dbPort, dbName,
		stats.OpenConnections,
		stats.InUse,
		stats.Idle)

	fmt.Fprintf(w, info)
}

func insertMockData(db *sql.DB) {
	// 定义 randomThingName 函数
	randomThingName := func() string {
		randomThings := []string{"tennis", "get married", "buy a house"}
		return randomThings[rand.Intn(len(randomThings))] // 返回一个随机选择的示例值
	}

	_, err := db.Exec("INSERT INTO learning (name, year, bigthing, learned) VALUES ($1, $2, $3, $4)", "diro", 2022, randomThingName(), "mock")
	if err != nil {
		log.Println("Error inserting mock data:", err)
	}
}

func displayTableInfo(db *sql.DB, id int, w http.ResponseWriter) {

	// 查询表

	rows, err := db.Query("SELECT * FROM lifespan.learning") // WHERE id = $1", id)
	if err != nil {
		http.Error(w, "Error querying database: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close() // 确保在使用后关闭 rows

	// 处理查询结果
	for rows.Next() {
		var name string
		var year int
		var bigthing, learned string
		if err := rows.Scan(&id, &name, &year, &bigthing, &learned); err != nil {
			http.Error(w, "Error scanning row: "+err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "ID: %d, Name: %s, Year: %d, Bigthing: %s, Learned: %s\n", id, name, year, bigthing, learned)
	}
	// ... 处理未处理的错误
}
