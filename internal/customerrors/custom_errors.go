package customerrors

import "fmt"

// ErrCodeCollision est retournée lorsqu'un code court existe déjà dans la base de données.
// Cette erreur personnalisée permet de distinguer une collision de code d'autres erreurs de base de données.
type ErrCodeCollision struct {
	Code     string // Le code court qui a généré la collision
	Attempts int    // Nombre de tentatives effectuées
}

// Error implémente l'interface error pour ErrCodeCollision
func (e *ErrCodeCollision) Error() string {
	return fmt.Sprintf("collision détectée pour le code court '%s' après %d tentative(s)", e.Code, e.Attempts)
}

// ErrLinkNotFound est retournée lorsqu'un lien n'est pas trouvé dans la base de données.
// Cette erreur personnalisée facilite la gestion des cas "404 Not Found" dans les handlers HTTP.
type ErrLinkNotFound struct {
	ShortCode string // Le code court recherché qui n'a pas été trouvé
}

// Error implémente l'interface error pour ErrLinkNotFound
func (e *ErrLinkNotFound) Error() string {
	return fmt.Sprintf("le lien avec le code court '%s' est introuvable", e.ShortCode)
}

// ErrInvalidURL est retournée lorsqu'une URL fournie est invalide ou mal formée.
type ErrInvalidURL struct {
	URL    string // L'URL invalide
	Reason string // La raison pour laquelle l'URL est invalide
}

// Error implémente l'interface error pour ErrInvalidURL
func (e *ErrInvalidURL) Error() string {
	return fmt.Sprintf("URL invalide '%s': %s", e.URL, e.Reason)
}

// ErrMaxRetriesExceeded est retournée lorsque le nombre maximum de tentatives est atteint.
// Utilisée principalement lors de la génération de codes courts avec gestion des collisions.
type ErrMaxRetriesExceeded struct {
	MaxRetries int    // Nombre maximum de tentatives autorisées
	Operation  string // Description de l'opération qui a échoué
}

// Error implémente l'interface error pour ErrMaxRetriesExceeded
func (e *ErrMaxRetriesExceeded) Error() string {
	return fmt.Sprintf("nombre maximum de tentatives (%d) dépassé pour l'opération: %s", e.MaxRetries, e.Operation)
}

// ErrDatabaseOperation est retournée pour les erreurs génériques de base de données.
type ErrDatabaseOperation struct {
	Operation string // Type d'opération (CREATE, READ, UPDATE, DELETE)
	Entity    string // Entité concernée (Link, Click)
	Err       error  // Erreur originale wrappée
}

// Error implémente l'interface error pour ErrDatabaseOperation
func (e *ErrDatabaseOperation) Error() string {
	return fmt.Sprintf("erreur lors de l'opération %s sur %s: %v", e.Operation, e.Entity, e.Err)
}

// Unwrap permet d'accéder à l'erreur originale wrappée
func (e *ErrDatabaseOperation) Unwrap() error {
	return e.Err
}
