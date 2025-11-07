package workers

import (
    "context"
    "log"
    "time"

    "github.com/axellelanca/urlshortener/internal/api"
    "github.com/axellelanca/urlshortener/internal/models"
)

// ClickRepository définit le contrat minimal attendu par les workers pour persister un clic.
// On le déclare ici localement pour éviter une dépendance forte sur une interface qui
// pourrait ne pas être encore formalisée dans le package repository.
type ClickRepository interface {
    CreateClick(c *models.Click) error
}

// clickWorker consomme des api.ClickEvent depuis le channel et les persiste en base via clickRepo.
// Il écoute le contexte pour un arrêt propre.
func clickWorker(ctx context.Context, id int, in <-chan api.ClickEvent, clickRepo ClickRepository) {
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
func StartClickWorkers(ctx context.Context, n int, in <-chan api.ClickEvent, clickRepo ClickRepository) {
    for i := 0; i < n; i++ {
        go clickWorker(ctx, i, in, clickRepo)
    }
}
package workers

import (
	"log"

	"github.com/axellelanca/urlshortener/internal/models"
	"github.com/axellelanca/urlshortener/internal/repository" // Nécessaire pour interagir avec le ClickRepository
)

// StartClickWorkers lance un pool de goroutines "workers" pour traiter les événements de clic.
// Chaque worker lira depuis le même 'clickEventsChan' et utilisera le 'clickRepo' pour la persistance.
func StartClickWorkers(workerCount int, clickEventsChan <-chan models.ClickEvent, clickRepo repository.ClickRepository) {
	log.Printf("Starting %d click worker(s)...", workerCount)
	for i := 0; i < workerCount; i++ {
		// Lance chaque worker dans sa propre goroutine.
		// Le channel est passé en lecture seule (<-chan) pour renforcer l'immutabilité du channel à l'intérieur du worker.
		go clickWorker(clickEventsChan, clickRepo)
	}
}

// clickWorker est la fonction exécutée par chaque goroutine worker.
// Elle tourne indéfiniment, lisant les événements de clic dès qu'ils sont disponibles dans le channel.
func clickWorker(clickEventsChan <-chan models.ClickEvent, clickRepo repository.ClickRepository) {
	for event := range clickEventsChan { // Boucle qui lit les événements du channel
		// TODO 1: Convertir le 'ClickEvent' (reçu du channel) en un modèle 'models.Click'.

		// TODO 2: Persister le clic en base de données via le 'clickRepo' (CreateClick).
		// Implémentez ici une gestion d'erreur simple : loggez l'erreur si la persistance échoue.
		// Pour un système en production, une logique de retry

		if err != nil {
			// Si une erreur se produit lors de l'enregistrement, logguez-la.
			// L'événement est "perdu" pour ce TP, mais dans un vrai système,
			// vous pourriez le remettre dans une file de retry ou une file d'erreurs.
			log.Printf("ERROR: Failed to save click for LinkID %d (UserAgent: %s, IP: %s): %v",
				event.LinkID, event.UserAgent, event.IPAddress, err)

		} else {
			// Log optionnel pour confirmer l'enregistrement (utile pour le débogage)
			log.Printf("Click recorded successfully for LinkID %d", event.LinkID)
		}
	}
}
