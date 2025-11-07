package repository

import (
	"fmt"

	"github.com/axellelanca/urlshortener/internal/models"
	"gorm.io/gorm"
)

// ClickRepository est une interface qui définit les méthodes d'accès aux données
// pour les opérations sur les clics. Cette abstraction permet à la couche service
// de rester indépendante de l'implémentation spécifique de la base de données.
//
// Pourquoi cette interface ?
// - Les workers vont l'utiliser pour enregistrer les clics
// - Le LinkService l'utilise pour compter les clics (statistiques)
type ClickRepository interface {
	// CreateClick insère un nouvel événement de clic dans la base de données
	CreateClick(click *models.Click) error
	
	// CountClicksByLinkID compte le nombre de clics pour un lien spécifique
	// Utilisé par LinkService pour les stats
	CountClicksByLinkID(linkID uint) (int, error)
}

// GormClickRepository est l'implémentation de l'interface ClickRepository utilisant GORM.
type GormClickRepository struct {
	db *gorm.DB // Référence à l'instance de la base de données GORM
}

// NewClickRepository crée et retourne une nouvelle instance de GormClickRepository.
// C'est la méthode recommandée pour obtenir un dépôt, garantissant que la connexion à la base de données est injectée.
func NewClickRepository(db *gorm.DB) *GormClickRepository {
	return &GormClickRepository{db: db}
}

// CreateClick insère un nouvel enregistrement de clic dans la base de données.
// Elle reçoit un pointeur vers une structure models.Click et la persiste en utilisant GORM.
//
// Cette méthode est appelée par les workers de clics de manière asynchrone.
func (r *GormClickRepository) CreateClick(click *models.Click) error {
	// db.Create() génère : INSERT INTO clicks (link_id, timestamp, user_agent, ip_address) VALUES (?, ?, ?, ?)
	// GORM va automatiquement remplir click.ID avec l'ID auto-incrémenté
	result := r.db.Create(click)
	if result.Error != nil {
		return fmt.Errorf("erreur lors de la création du clic : %w", result.Error)
	}
	return nil
}

// CountClicksByLinkID compte le nombre total de clics pour un ID de lien donné.
// Cette méthode est utilisée pour fournir des statistiques pour une URL courte.
func (r *GormClickRepository) CountClicksByLinkID(linkID uint) (int, error) {
	var count int64 // GORM retourne un int64 pour les décomptes
	// db.Model(&models.Click{}) spécifie la table 'clicks'
	// .Where("link_id = ?", linkID) filtre pour ce lien spécifique
	// .Count(&count) génère : SELECT COUNT(*) FROM clicks WHERE link_id = ?
	result := r.db.Model(&models.Click{}).Where("link_id = ?", linkID).Count(&count)
	if result.Error != nil {
		return 0, fmt.Errorf("erreur lors du comptage des clics pour LinkID %d : %w", linkID, result.Error)
	}
	return int(count), nil // Convert the int64 count to an int
}
