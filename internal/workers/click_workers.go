package workers

import (
	"context"
	"log"
	"time"

	"github.com/axellelanca/urlshortener/internal/api"
	"github.com/axellelanca/urlshortener/internal/models"
	"github.com/axellelanca/urlshortener/internal/repository"
)

// clickWorker consomme des api.ClickEvent depuis le channel et les persiste en base via clickRepo.
// Il écoute le contexte pour un arrêt propre.
func clickWorker(ctx context.Context, id int, in <-chan api.ClickEvent, clickRepo repository.ClickRepository) {
	log.Printf("clickWorker %d: started", id)
	defer log.Printf("clickWorker %d: stopped", id)

	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-in:
			if !ok {
				// channel fermé
				return
			}

			// Convertir l'événement en modèle GORM Click
			click := &models.Click{
				LinkID:    ev.LinkID,
				Timestamp: ev.Timestamp,
				UserAgent: ev.UserAgent,
				IPAddress: ev.IP,
			}

			// Tenter de persister le clic
			if err := clickRepo.CreateClick(click); err != nil {
				// Log et continue (on ne veut pas bloquer le worker sur une erreur)
				log.Printf("clickWorker %d: failed to persist click for link %d: %v", id, ev.LinkID, err)
			} else {
				log.Printf("clickWorker %d: persisted click for link %d", id, ev.LinkID)
			}

			// Petite pause pour éviter hot-loop si nécessaire (configurable si besoin)
			time.Sleep(5 * time.Millisecond)
		}
	}
}

// StartClickWorkers démarre n workers et retourne immédiatement.
// Le caller doit fournir un contexte annulable pour gérer l'arrêt propre.
func StartClickWorkers(ctx context.Context, n int, in <-chan api.ClickEvent, clickRepo repository.ClickRepository) {
	log.Printf("Starting %d click worker(s)...", n)
	for i := 0; i < n; i++ {
		go clickWorker(ctx, i, in, clickRepo)
	}
}
