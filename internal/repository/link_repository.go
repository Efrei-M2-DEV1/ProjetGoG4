package repository

import (
	"fmt"

	"github.com/axellelanca/urlshortener/internal/models"
	"gorm.io/gorm"
)

// LinkRepository est une interface qui définit les méthodes d'accès aux données
// pour les opérations CRUD sur les liens.
// 
// Pourquoi une interface ?
// - Permet de découpler la logique métier (services) de l'implémentation (GORM)
// - Facilite les tests unitaires (on peut créer un mock/fake repository)
// - Respecte le principe SOLID "Dependency Inversion Principle"
type LinkRepository interface {
	// CreateLink insère un nouveau lien dans la base de données
	CreateLink(link *models.Link) error
	
	// GetLinkByShortCode récupère un lien par son code court unique
	// Retourne gorm.ErrRecordNotFound si non trouvé
	GetLinkByShortCode(shortCode string) (*models.Link, error)
	
	// GetAllLinks récupère tous les liens de la base de données
	// Utilisé par le moniteur pour vérifier toutes les URLs
	GetAllLinks() ([]models.Link, error)
	
	// CountClicksByLinkID compte le nombre total de clics pour un lien donné
	CountClicksByLinkID(linkID uint) (int, error)
}

// GormLinkRepository est l'implémentation de LinkRepository utilisant GORM.
// Elle contient une référence à la connexion GORM pour effectuer les requêtes SQL.
type GormLinkRepository struct {
	db *gorm.DB // Connexion à la base de données GORM
}

// NewLinkRepository crée et retourne une nouvelle instance de GormLinkRepository.
// Cette fonction est un "constructeur" en Go (Go n'a pas de constructeurs natifs).
// Elle retourne *GormLinkRepository, qui implémente l'interface LinkRepository.
func NewLinkRepository(db *gorm.DB) *GormLinkRepository {
	return &GormLinkRepository{db: db}
}

// CreateLink insère un nouveau lien dans la base de données.
// GORM va automatiquement générer le SQL INSERT et remplir l'ID du lien.
func (r *GormLinkRepository) CreateLink(link *models.Link) error {
	// db.Create() insère un nouvel enregistrement dans la table 'links'
	// GORM va :
	// 1. Générer : INSERT INTO links (short_code, long_url, created_at) VALUES (?, ?, ?)
	// 2. Remplir automatiquement link.ID avec l'ID auto-incrémenté
	// 3. Remplir link.CreatedAt si c'est un champ time.Time
	result := r.db.Create(link)
	if result.Error != nil {
		return fmt.Errorf("erreur lors de la création du lien : %w", result.Error)
	}
	return nil
}

// GetLinkByShortCode récupère un lien de la base de données en utilisant son shortCode.
// Il renvoie gorm.ErrRecordNotFound si aucun lien n'est trouvé avec ce shortCode.
func (r *GormLinkRepository) GetLinkByShortCode(shortCode string) (*models.Link, error) {
	var link models.Link
	// db.Where().First() génère : SELECT * FROM links WHERE short_code = ? LIMIT 1
	// First() renvoie le premier résultat trouvé
	// Si aucun résultat : retourne gorm.ErrRecordNotFound
	result := r.db.Where("short_code = ?", shortCode).First(&link)
	if result.Error != nil {
		// On wrappe l'erreur pour ajouter du contexte
		return nil, fmt.Errorf("erreur lors de la récupération du lien par shortCode '%s' : %w", shortCode, result.Error)
	}
	return &link, nil
}

// GetAllLinks récupère tous les liens de la base de données.
// Cette méthode est utilisée par le moniteur d'URLs pour vérifier l'état de toutes les URLs.
func (r *GormLinkRepository) GetAllLinks() ([]models.Link, error) {
	var links []models.Link
	// db.Find() génère : SELECT * FROM links
	// Find() récupère tous les enregistrements et les mappe dans le slice
	result := r.db.Find(&links)
	if result.Error != nil {
		return nil, fmt.Errorf("erreur lors de la récupération de tous les liens : %w", result.Error)
	}
	return links, nil
}

// CountClicksByLinkID compte le nombre total de clics pour un ID de lien donné.
// Cette méthode compte les enregistrements dans la table 'clicks' où link_id = linkID.
func (r *GormLinkRepository) CountClicksByLinkID(linkID uint) (int, error) {
	var count int64 // GORM retourne un int64 pour les comptes
	// db.Model(&models.Click{}) spécifie quelle table utiliser ('clicks')
	// .Where("link_id = ?", linkID) filtre les clics pour ce lien spécifique
	// .Count(&count) génère : SELECT COUNT(*) FROM clicks WHERE link_id = ?
	result := r.db.Model(&models.Click{}).Where("link_id = ?", linkID).Count(&count)
	if result.Error != nil {
		return 0, fmt.Errorf("erreur lors du comptage des clics pour LinkID %d : %w", linkID, result.Error)
	}
	return int(count), nil // Convertit int64 en int
}
