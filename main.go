package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq" // 改用 PostgreSQL 驱动
)

const (
	dbHost = "lifeplan-instance-1.c1w4eg0mwku5.us-west-2.rds.amazonaws.com"
	dbPort = "5432"
)

func main() {
	// 从环境变量获取数据库凭据
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	// 检查必要的环境变量是否存在
	if dbUser == "" || dbPassword == "" || dbName == "" {
		log.Println("必要的数据库环境变量未设置 (DB_USER, DB_PASSWORD, DB_NAME)")
	}

	// 構建連接字符串
	dbURI := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=require",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	db, err := sql.Open("postgres", dbURI)
	if err != nil {
		log.Println(err)
	}
	defer db.Close()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, World!")
	})

	http.HandleFunc("/dbinfo", func(w http.ResponseWriter, r *http.Request) {
		// 测试数据库连接
		err := db.Ping()
		if err != nil {
			http.Error(w, "数据库连接失败: "+err.Error(), http.StatusInternalServerError)
			//	return
		}

		// 获取数据库统计信息
		stats := db.Stats()

		// 构建响应信息
		info := fmt.Sprintf(`数据库连接信息:
			主机: %s
			端口: %s
			数据库名: %s
			打开的连接数: %d
			使用中的连接数: %d
			空闲连接数: %d`,
			dbHost, dbPort, dbName,
			stats.OpenConnections,
			stats.InUse,
			stats.Idle)

		fmt.Fprintf(w, info)
	})

	log.Println("Server is running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
