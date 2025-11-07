package models

import "time"

// Link représente un lien raccourci dans la base de données.
// Les tags `gorm:"..."` définissent comment GORM doit mapper cette structure à une table SQL.
// GORM utilisera ces tags pour créer automatiquement la table 'links' avec les bonnes contraintes.
type Link struct {
	// ID est la clé primaire auto-incrémentée par GORM
	ID uint `gorm:"primaryKey"`
	
	// ShortCode est le code court unique (ex: "abc123")
	// - unique : garantit qu'aucun doublon ne peut exister en BDD
	// - index : crée un index pour des recherches rapides par ShortCode
	// - size:10 : limite la longueur du champ VARCHAR en BDD à 10 caractères
	ShortCode string `gorm:"unique;index;size:10"`
	
	// LongURL est l'URL originale complète à laquelle le ShortCode redirige
	// - not null : ce champ est obligatoire, ne peut pas être vide
	LongURL string `gorm:"not null"`
	
	// CreatedAt est l'horodatage de création du lien
	// GORM gère automatiquement ce champ (le remplit à la création)
	CreatedAt time.Time
}
