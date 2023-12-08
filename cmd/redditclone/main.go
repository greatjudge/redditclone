package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"text/template"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"go.uber.org/zap"

	_ "github.com/go-sql-driver/mysql"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/greatjudge/redditclone/pkg/handlers"
	"github.com/greatjudge/redditclone/pkg/middleware"
	"github.com/greatjudge/redditclone/pkg/post"
	"github.com/greatjudge/redditclone/pkg/session"
	"github.com/greatjudge/redditclone/pkg/user"
)

func init() {
	// loads values from .env into the system
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

func initSQLDB() (*sql.DB, error) {
	dsn := os.Getenv("MYSQL_DSN")
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("fail sql.Open: %w", err)
	}
	db.SetMaxOpenConns(10)
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}

func initMongoDB() (*mongo.Collection, error) {
	ctx := context.Background()
	url := os.Getenv("MONGO_URL")
	sess, err := mongo.Connect(ctx, options.Client().ApplyURI(url))
	if err != nil {
		return nil, fmt.Errorf("fail connect mongo: %w", err)
	}
	db := os.Getenv("MONGO_DB")
	collection := os.Getenv("MONGO_COLLECTION")
	return sess.Database(db).Collection(collection), nil
}

func main() {
	// entries, err := os.ReadDir("./06_databases/99_hw/redditclone")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// dirs := make([]string, 0)
	// for _, e := range entries {
	// 	dirs = append(dirs, e.Name())
	// }
	// panic(strings.Join(dirs, " "))

	collection, err := initMongoDB()
	if err != nil {
		panic(err)
	}
	db, err := initSQLDB()
	if err != nil {
		panic(err)
	}

	tokenSecret := []byte(os.Getenv("TOKEN_SECRET"))
	templatePath := os.Getenv("TEMPLATE_DIR")
	tmpl := template.Must(template.ParseFiles(templatePath))

	zapLogger, err := zap.NewProduction()
	if err != nil {
		panic("log init error")
	}
	defer func() {
		err = zapLogger.Sync()
		if err != nil {
			fmt.Println(err)
		}
	}()
	logger := zapLogger.Sugar()

	userRepo := user.NewMysqlRepo(db, logger)

	sm := session.NewSessionsManagerMySQL(
		db,
		userRepo,
		tokenSecret,
		logger,
	)

	userHandler := &handlers.UserHandler{
		UserRepo: userRepo,
		Logger:   logger,
		Sessions: sm,
	}

	postRepo := post.NewMongoDBRepo(collection)
	postHandler := &handlers.PostHandler{
		PostRepo: postRepo,
		Logger:   logger,
	}

	router := mux.NewRouter()

	staticDir := os.Getenv("STATIC_DIR")
	routerStatic := router.PathPrefix("/static/").Subrouter()
	staticHandler := http.StripPrefix(
		"/static/",
		http.FileServer(http.Dir(staticDir)),
	)
	routerStatic.PathPrefix("/").Handler(staticHandler).Methods("GET")

	router.HandleFunc("/api/register", userHandler.Register).Methods("POST")
	router.HandleFunc("/api/login", userHandler.Login).Methods("POST")
	router.HandleFunc("/api/posts/", postHandler.List).Methods("GET")
	router.HandleFunc("/api/posts/{CATEGORY_NAME}", postHandler.ListByCategory).Methods("GET")
	router.HandleFunc("/api/post/{POST_ID}", postHandler.GetByID).Methods("GET")
	router.HandleFunc("/api/user/{USER_LOGIN}", postHandler.GetUserPosts).Methods("GET")

	router.Handle("/api/posts", middleware.Auth(sm, http.HandlerFunc(postHandler.Add))).Methods("POST")
	router.Handle("/api/post/{POST_ID}", middleware.Auth(sm, http.HandlerFunc(postHandler.AddComment))).Methods("POST")
	router.Handle("/api/post/{POST_ID}/{COMMENT_ID}", middleware.Auth(sm, http.HandlerFunc(postHandler.DeleteComment))).Methods("DELETE")
	router.Handle("/api/post/{POST_ID}/upvote", middleware.Auth(sm, http.HandlerFunc(postHandler.Upvote))).Methods("GET")
	router.Handle("/api/post/{POST_ID}/downvote", middleware.Auth(sm, http.HandlerFunc(postHandler.Downvote))).Methods("GET")
	router.Handle("/api/post/{POST_ID}/unvote", middleware.Auth(sm, http.HandlerFunc(postHandler.Unvote))).Methods("GET")
	router.Handle("/api/post/{POST_ID}", middleware.Auth(sm, http.HandlerFunc(postHandler.Delete))).Methods("DELETE")

	router.PathPrefix("/").HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			err = tmpl.Execute(w, struct{}{})
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		},
	)

	handler := middleware.AccessLog(logger, router)
	handler = middleware.Panic(logger, handler)

	fmt.Println("listening...")
	err = http.ListenAndServe(":8080", handler)
	if err != nil {
		panic(err)
	}
}
