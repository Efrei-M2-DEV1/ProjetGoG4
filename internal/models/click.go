package models

import "time"

// Click représente un événement de clic sur un lien raccourci.
// GORM utilisera ces tags pour créer la table 'clicks'.
type Click struct {
	ID        uint      `gorm:"primaryKey"`        // Clé primaire
	LinkID    uint      `gorm:"index"`             // Clé étrangère vers la table 'links', indexée pour des requêtes efficaces
	Link      Link      `gorm:"foreignKey:LinkID"` // Relation GORM: indique que LinkID est une FK vers le champ ID de Link
	Timestamp time.Time // Horodatage précis du clic
	UserAgent string    `gorm:"size:255"` // User-Agent de l'utilisateur qui a cliqué (informations sur le navigateur/OS)
	IPAddress string    `gorm:"size:50"`  // Adresse IP de l'utilisateur
}

// ClickEvent représente un événement de clic brut, destiné à être passé via un channel.
// Ce n'est PAS un modèle GORM direct, mais une structure légère pour la communication
// entre les goroutines (du handler HTTP vers les workers).
// 
// Pourquoi deux structs différentes ?
// - Click : struct GORM complète avec relation, utilisée pour la persistance en BDD
// - ClickEvent : struct simple et légère, utilisée pour passer des données via un channel
//   Elle ne contient que les infos essentielles (pas besoin de la relation Link complète)
type ClickEvent struct {
	LinkID    uint      // ID du lien cliqué
	Timestamp time.Time // Moment du clic
	UserAgent string    // User-Agent du navigateur
	IPAddress string    // Adresse IP du visiteur
}
