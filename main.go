package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	_ "github.com/lib/pq" // 改用 PostgreSQL 驱动
)

const (
	dbHost = "lifeplan-instance-1.c1w4eg0mwku5.us-west-2.rds.amazonaws.com"
	dbPort = "5432"
)

func getDBCredentials(secretName string, region string) (string, string, string, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region)},
	)
	if err != nil {
		log.Println(err)
		return "", "", "", err
	}

	svc := secretsmanager.New(sess)
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	}

	result, err := svc.GetSecretValue(input)
	if err != nil {
		log.Println(err)
		return "", "", "", err
	}

	var secretString string
	if result.SecretString != nil {
		secretString = *result.SecretString
	}

	var secretMap map[string]string
	err = json.Unmarshal([]byte(secretString), &secretMap)
	if err != nil {
		log.Println(err)
		return "", "", "", err
	}

	return secretMap["username"], secretMap["password"], secretMap["dbname"], nil
}

func main() {
	// 从 Secrets Manager 获取数据库凭据
	dbUser, dbPassword, dbName, err := getDBCredentials("rds!cluster-6de70cb9-03ec-43aa-b93c-06f9ea68965d", "us-west-2")
	if err != nil {
		log.Fatalf("无法获取数据库凭据: %v", err)
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
		fmt.Fprint(w, "Hello, World!")
		fmt.Fprintf(w, "dbUser: %s", dbUser)
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
