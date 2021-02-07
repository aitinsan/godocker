package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/golangcollege/sessions"
	"github.com/jackc/pgx/v4/pgxpool"
	"html/template"
	"log"
	"net/http"
	"os"
	"go.com/pkg/models/pg"
	"time"
)

type application struct {
	errorLog      *log.Logger
	infoLog       *log.Logger
	snippets      *pg.SnippetModel
	templateCache map[string]*template.Template
	session       *sessions.Session
}

func main() {
	username := flag.String("username", "postgres", "username")
	password := flag.String("password", "postgres", "password")

	host := flag.String("host", "localhost", "host")
	port := flag.String("port", "5432", "port")
	dbname := flag.String("dbname", "snippetbox", "dbname")
	addr := flag.String("addr", "4000", "HTTP network address")

	secret := flag.String("secret", "s6Ndh+pPbnzHbS*+9Pk8qGWhTzbpa@ge", "Secret key")
	flag.Parse()
	connString := fmt.Sprintf("postgres://%v:%v@%v:%v/%v", *username, *password, *host, *port, *dbname)

	conn, err := pgxpool.Connect(context.Background(), connString)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()
	var greeting string
	err = conn.QueryRow(context.Background(), "select 'Hello, world!'").Scan(&greeting)
	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(greeting)

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	templateCache, err := newTemplateCache("./ui/html/")
	if err != nil {
		errorLog.Fatal(err)
	}

	session := sessions.New([]byte(*secret))
	session.Lifetime = 12 * time.Hour

	app := &application{
		errorLog:      errorLog,
		infoLog:       infoLog,
		snippets:      &pg.SnippetModel{DB: conn},
		templateCache: templateCache,
		session:       session,
	}

	srv := &http.Server{
		Addr:     fmt.Sprintf(":%v", *addr),
		ErrorLog: errorLog,
		Handler:  app.routes(),
	}
	infoLog.Printf("Starting server on %s", *addr)
	infoLog.Println(connString)
	myErr := srv.ListenAndServe()
	errorLog.Fatal(myErr)
}
