package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"novobanco/internal/infrastructure/config"
	"novobanco/internal/infrastructure/database"
	router "novobanco/internal/infrastructure/http"
	"novobanco/internal/infrastructure/http/handler"
	"novobanco/internal/usecase"
)

func main() {
	log.Println("Iniciando microservicio de NovoBanco...")

	// 1. Cargar configuración de variables de entorno
	cfg := config.LoadConfig()

	// 2. Conectar a PostgreSQL
	db, err := database.NewPostgresDB(cfg.DBConn)
	if err != nil {
		log.Fatalf("Error al conectar a PostgreSQL: %v", err)
	}
	defer db.Close()

	log.Println("Conexión a PostgreSQL establecida con éxito.")

	// 3. Inicializar Repositorios (Infraestructura)
	accountRepo := database.NewPostgresAccountRepository(db)
	txRepo := database.NewPostgresTransactionRepository(db)
	txManager := database.NewSQLTxManager(db)

	// 4. Inicializar Casos de Uso (Usecases)
	accountUsecase := usecase.NewAccountUsecase(accountRepo)
	txUsecase := usecase.NewTransactionUsecase(accountRepo, txRepo, txManager)

	// 5. Inicializar Handlers HTTP
	accountHandler := handler.NewAccountHandler(accountUsecase)
	txHandler := handler.NewTransactionHandler(txUsecase)

	// 6. Configurar Rutas
	r := router.NewRouter(accountHandler, txHandler)

	// 7. Arrancar Servidor con Graceful Shutdown
	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		log.Printf("Servidor HTTP corriendo en el puerto %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error al levantar el servidor: %v", err)
		}
	}()

	// Capturar señales del sistema operativo para apagar de forma segura
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Señal de apagado recibida. Deteniendo el servidor de forma ordenada...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Error durante el apagado del servidor: %v", err)
	}

	log.Println("Microservicio detenido exitosamente.")
}
