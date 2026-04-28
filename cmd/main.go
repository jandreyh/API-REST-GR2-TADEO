package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"todo-api/api/auth"
	"todo-api/api/handlers"
	"todo-api/api/middleware"
	"todo-api/api/repository"
)

func main() {
	port := getenv("PORT", "8080")
	dbPath := getenv("DB_PATH", "todo.db")
	jwtSecret := getenv("JWT_SECRET", "mi-secreto-super-seguro-cambiar-en-produccion")

	repo, err := repository.NewSQLiteRepository(dbPath)
	if err != nil {
		log.Fatalf("Error al inicializar la base de datos: %v", err)
	}
	defer repo.Close()

	jwtService := auth.NewJWTService(jwtSecret)

	taskHandler := handlers.NewTaskHandler(repo)
	authHandler := handlers.NewAuthHandler(repo, jwtService)

	mux := http.NewServeMux()

	// Rutas públicas
	mux.HandleFunc("POST /api/auth/register", authHandler.Register)
	mux.HandleFunc("POST /api/auth/login", authHandler.Login)
	mux.HandleFunc("GET /api/health", healthCheck)

	// Rutas protegidas
	protected := middleware.Chain(
		middleware.Logger,
		middleware.Auth(jwtService),
	)

	mux.Handle("GET /api/tasks", protected(http.HandlerFunc(taskHandler.GetAll)))
	mux.Handle("POST /api/tasks", protected(http.HandlerFunc(taskHandler.Create)))
	mux.Handle("GET /api/tasks/{id}", protected(http.HandlerFunc(taskHandler.GetByID)))
	mux.Handle("PUT /api/tasks/{id}", protected(http.HandlerFunc(taskHandler.Update)))
	mux.Handle("DELETE /api/tasks/{id}", protected(http.HandlerFunc(taskHandler.Delete)))
	mux.Handle("PATCH /api/tasks/{id}/complete", protected(http.HandlerFunc(taskHandler.MarkComplete)))

	log.Printf("Servidor corriendo en http://localhost:%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
	}
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"status":"ok","message":"API funcionando correctamente"}`)
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
