package main

import (
	"database/sql"
	"dbms/internal/controller"
	"dbms/internal/repository"
	"dbms/internal/router"
	"dbms/internal/service"
	"dbms/pkg"
	"log"

	_ "github.com/microsoft/go-mssqldb"
)

func main() {

	config := pkg.LoadConfig()
	db, err := sql.Open("sqlserver", config.DBURL)
	if err != nil {
		log.Fatal("Error opening database: ", err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatal("Không thể login vào SQL Server:", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(0)

	repo := &repository.Repository{DB: db}
	svc := &service.Service{Repo: repo, Ssm: service.NewSessionManager()}
	ctrl := &controller.Controller{Svc: svc, Config: config}

	r := router.SetupRouter(ctrl)
	r.Run(":8080")
}
