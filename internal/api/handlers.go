package api

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/axellelanca/urlshortener/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm" // Pour gérer gorm.ErrRecordNotFound
)

// ClickEvent définit la structure minimale d'un événement de clic transmis via le channel.
// On le définit ici localement pour découpler la dépendance en attendant que la
// Personne 1 ait complété models.ClickEvent si nécessaire.
type ClickEvent struct {
	LinkID    uint
	ShortCode string
	Timestamp time.Time
	UserAgent string
	IP        string
}

// ClickEventsChannel est le channel bufferisé global utilisé pour envoyer les événements
// de clic aux workers asynchrones. Il est initialisé dans SetupRoutes si nil.
var ClickEventsChannel chan ClickEvent

// LinkServiceInterface définit le contrat minimal attendu par les handlers.
// Nous déclarons une interface locale pour rester découplés de l'implémentation
// concrète fournie par la Personne 2 (services.LinkService). Si ce dernier
// implémente ces méthodes, il satisfera automatiquement cette interface.
type LinkServiceInterface interface {
	CreateLink(longURL string) (*models.Link, error)
	GetLinkByShortCode(shortCode string) (*models.Link, error)
	GetLinkStats(shortCode string) (*models.Link, int, error)
}

// SetupRoutes configure toutes les routes de l'API Gin et injecte les dépendances nécessaires.
// bufferSize permet de configurer la taille du channel pour les événements de clic.
// Si bufferSize <= 0, on utilise une valeur par défaut raisonnable.
func SetupRoutes(router *gin.Engine, linkService LinkServiceInterface, bufferSize int) {
	// Défaut si non fourni
	if bufferSize <= 0 {
		bufferSize = 100
	}

	// Initialisation du channel bufferisé si nécessaire
	if ClickEventsChannel == nil {
		ClickEventsChannel = make(chan ClickEvent, bufferSize)
	}

	// Route de Health Check
	router.GET("/health", HealthCheckHandler)

	// Routes API
	api := router.Group("/api/v1")
	{
		api.POST("/links", CreateShortLinkHandler(linkService))
		api.GET("/links/:shortCode/stats", GetLinkStatsHandler(linkService))
	}

	// Route de Redirection (au niveau racine pour les short codes)
	router.GET("/:shortCode", RedirectHandler(linkService))
}

// HealthCheckHandler gère la route /health pour vérifier l'état du service.
func HealthCheckHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// CreateLinkRequest représente le corps de la requête JSON pour la création d'un lien.
type CreateLinkRequest struct {
	LongURL string `json:"long_url" binding:"required,url"`
}

// CreateShortLinkHandler gère la création d'une URL courte.
func CreateShortLinkHandler(linkService LinkServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateLinkRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		link, err := linkService.CreateLink(req.LongURL)
		if err != nil {
			log.Printf("CreateLink error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create short link"})
			return
		}

		// Note: on utilise localhost:8080 ici comme fallback; l'appelant (cmd/server)
		// peut facilement construire l'URL complète en utilisant la config si disponible.
		c.JSON(http.StatusCreated, gin.H{
			"short_code":     link.ShortCode,
			"long_url":       link.LongURL,
			"full_short_url": "http://localhost:8080/" + link.ShortCode,
		})
	}
}

// RedirectHandler gère la redirection d'une URL courte vers l'URL longue et l'enregistrement asynchrone des clics.
func RedirectHandler(linkService LinkServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		shortCode := c.Param("shortCode")

		link, err := linkService.GetLinkByShortCode(shortCode)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "link not found"})
				return
			}
			log.Printf("Error retrieving link for %s: %v", shortCode, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		// Construire l'événement de clic à envoyer au worker.
		clickEvent := ClickEvent{
			LinkID:    link.ID,
			ShortCode: shortCode,
			Timestamp: time.Now().UTC(),
			UserAgent: c.GetHeader("User-Agent"),
			IP:        c.ClientIP(),
		}

		// Envoi non-bloquant dans le channel pour ne jamais ralentir la redirection.
		select {
		case ClickEventsChannel <- clickEvent:
			// envoyé avec succès
		default:
			log.Printf("Warning: ClickEventsChannel is full, dropping click event for %s.", shortCode)
		}

		// Redirection instantanée vers l'URL longue
		c.Redirect(http.StatusFound, link.LongURL)
	}
}

// GetLinkStatsHandler gère la récupération des statistiques pour un lien spécifique.
func GetLinkStatsHandler(linkService LinkServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		shortCode := c.Param("shortCode")

		link, totalClicks, err := linkService.GetLinkStats(shortCode)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "link not found"})
				return
			}
			log.Printf("Error getting stats for %s: %v", shortCode, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"short_code":   link.ShortCode,
			"long_url":     link.LongURL,
			"total_clicks": totalClicks,
		})
	}
}
