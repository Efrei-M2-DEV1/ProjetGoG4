package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	cmd2 "github.com/axellelanca/urlshortener/cmd"
	"github.com/axellelanca/urlshortener/internal/api"
	"github.com/axellelanca/urlshortener/internal/monitor"
	"github.com/axellelanca/urlshortener/internal/repository"
	"github.com/axellelanca/urlshortener/internal/services"
	"github.com/axellelanca/urlshortener/internal/workers"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"gorm.io/driver/sqlite" // Driver SQLite pour GORM
	"gorm.io/gorm"
)

// RunServerCmd représente la commande 'run-server' de Cobra.
// C'est le point d'entrée pour lancer le serveur de l'application.
var RunServerCmd = &cobra.Command{
	Use:   "run-server",
	Short: "Lance le serveur API de raccourcissement d'URLs et les processus de fond.",
	Long: `Cette commande initialise la base de données, configure les APIs,
démarre les workers asynchrones pour les clics et le moniteur d'URLs,
puis lance le serveur HTTP.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Créer une variable qui stock la configuration chargée globalement via cmd.Cfg
		cfg := cmd2.Cfg
		if cfg == nil {
			log.Fatalf("FATAL: Configuration non chargée")
		}

		// Initialiser la connexion à la BDD
		db, err := gorm.Open(sqlite.Open(cfg.Database.Name), &gorm.Config{})
		if err != nil {
			log.Fatalf("FATAL: Échec de la connexion à la base de données: %v", err)
		}

		// Initialiser les repositories.
		linkRepo := repository.NewLinkRepository(db)
		clickRepo := repository.NewClickRepository(db)

		// Laissez le log
		log.Println("Repositories initialisés.")

		// Initialiser les services métiers.
		linkService := services.NewLinkService(linkRepo)
		_ = services.NewClickService(clickRepo) // Service initialisé mais non utilisé directement ici

		// Laissez le log
		log.Println("Services métiers initialisés.")

		// Initialiser le channel ClickEventsChannel (api/handlers) des événements de clic et lancer les workers (StartClickWorkers).
		// Le channel est bufferisé avec la taille configurée.
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Initialiser le channel dans SetupRoutes, mais on doit le créer avant
		api.ClickEventsChannel = make(chan api.ClickEvent, cfg.Analytics.BufferSize)
		workers.StartClickWorkers(ctx, cfg.Analytics.WorkerCount, api.ClickEventsChannel, clickRepo)

		log.Printf("Channel d'événements de clic initialisé avec un buffer de %d. %d worker(s) de clics démarré(s).",
			cfg.Analytics.BufferSize, cfg.Analytics.WorkerCount)

		// Initialiser et lancer le moniteur d'URLs.
		// Utilisez l'intervalle configuré
		monitorInterval := time.Duration(cfg.Monitor.IntervalMinutes) * time.Minute
		urlMonitor := monitor.NewUrlMonitor(linkRepo, monitorInterval)

		// Lancez le moniteur dans sa propre goroutine.
		go urlMonitor.Start()

		log.Printf("Moniteur d'URLs démarré avec un intervalle de %v.", monitorInterval)

		// Configurer le routeur Gin et les handlers API.
		router := gin.Default()
		api.SetupRoutes(router, linkService, cfg.Analytics.BufferSize)

		// Pas toucher au log
		log.Println("Routes API configurées.")

		// Créer le serveur HTTP Gin
		serverAddr := fmt.Sprintf(":%d", cfg.Server.Port)
		srv := &http.Server{
			Addr:    serverAddr,
			Handler: router,
		}

		// Démarrer le serveur Gin dans une goroutine anonyme pour ne pas bloquer.
		go func() {
			log.Printf("Serveur HTTP démarré sur %s", serverAddr)
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("FATAL: Erreur lors du démarrage du serveur: %v", err)
			}
		}()

		// Gérer l'arrêt propre du serveur (graceful shutdown).
		// Créez un channel pour les signaux OS (SIGINT, SIGTERM), bufferisé à 1.
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM) // Attendre Ctrl+C ou signal d'arrêt

		// Bloquer jusqu'à ce qu'un signal d'arrêt soit reçu.
		<-quit
		log.Println("Signal d'arrêt reçu. Arrêt du serveur...")

		// Arrêt propre du serveur HTTP avec un timeout.
		log.Println("Arrêt en cours... Donnez un peu de temps aux workers pour finir.")
		cancel() // Annuler le contexte pour arrêter les workers
		time.Sleep(5 * time.Second)

		// Fermer le serveur HTTP
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("Erreur lors de l'arrêt du serveur: %v", err)
		}

		log.Println("Serveur arrêté proprement.")
	},
}

func init() {
	// Ajouter la commande
	cmd2.RootCmd.AddCommand(RunServerCmd)
}
