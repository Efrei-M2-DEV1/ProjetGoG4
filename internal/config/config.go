package config

import (
	"fmt"
	"log" // Pour logger les informations ou erreurs de chargement de config

	"github.com/spf13/viper" // La bibliothèque pour la gestion de configuration
)

// Config est la structure principale qui mappe l'intégralité de la configuration de l'application.
// Les tags `mapstructure` sont utilisés par Viper pour mapper les clés du fichier de config
// (ou des variables d'environnement) aux champs de la structure Go.
//
// Structure hiérarchique qui reflète le fichier config.yaml :
// config.yaml:
//   server:
//     port: 8080
//   devient en Go:
//   Config.Server.Port = 8080
type Config struct {
	Server    ServerConfig    `mapstructure:"server"`    // Configuration du serveur web
	Database  DatabaseConfig  `mapstructure:"database"`  // Configuration de la base de données
	Analytics AnalyticsConfig `mapstructure:"analytics"` // Configuration des analytics (workers)
	Monitor   MonitorConfig   `mapstructure:"monitor"`   // Configuration du moniteur d'URLs
}

// ServerConfig contient les paramètres du serveur HTTP Gin
type ServerConfig struct {
	Port    int    `mapstructure:"port"`     // Port d'écoute (ex: 8080)
	BaseURL string `mapstructure:"base_url"` // URL de base pour construire les URLs courtes complètes
}

// DatabaseConfig contient les paramètres de la base de données
type DatabaseConfig struct {
	Name string `mapstructure:"name"` // Nom du fichier SQLite (ex: "url_shortener.db")
}

// AnalyticsConfig contient les paramètres pour le système d'analytics asynchrone
type AnalyticsConfig struct {
	BufferSize  int `mapstructure:"buffer_size"`  // Taille du buffer du channel de clics
	WorkerCount int `mapstructure:"worker_count"` // Nombre de goroutines workers
}

// MonitorConfig contient les paramètres pour le moniteur d'URLs
type MonitorConfig struct {
	IntervalMinutes int `mapstructure:"interval_minutes"` // Intervalle de vérification en minutes
}

// LoadConfig charge la configuration de l'application en utilisant Viper.
// Elle recherche un fichier 'config.yaml' dans le dossier 'configs/'.
// Elle définit également des valeurs par défaut si le fichier de config est absent ou incomplet.
func LoadConfig() (*Config, error) {
	// Étape 1: Spécifier où Viper doit chercher les fichiers de configuration
	// On cherche dans le dossier 'configs' relatif au répertoire d'exécution.
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("../configs") // Au cas où on exécute depuis un sous-dossier
	viper.AddConfigPath(".")          // Fallback : répertoire courant

	// Étape 2: Spécifier le nom du fichier de config (sans l'extension)
	viper.SetConfigName("config")

	// Étape 3: Spécifier le type de fichier de config
	viper.SetConfigType("yaml")

	// Étape 4: Définir les valeurs par défaut pour toutes les options de configuration.
	// Ces valeurs seront utilisées si les clés correspondantes ne sont pas trouvées dans le fichier
	// ou si le fichier n'existe pas. C'est une bonne pratique pour la robustesse.
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.base_url", "http://localhost:8080")
	viper.SetDefault("database.name", "url_shortener.db")
	viper.SetDefault("analytics.buffer_size", 1000)
	viper.SetDefault("analytics.worker_count", 5)
	viper.SetDefault("monitor.interval_minutes", 5)

	// Étape 5: Lire le fichier de configuration
	// ReadInConfig() cherche et lit le fichier config.yaml
	if err := viper.ReadInConfig(); err != nil {
		// Si le fichier n'est pas trouvé, ce n'est pas grave grâce aux valeurs par défaut
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("Aucun fichier de configuration trouvé. Utilisation des valeurs par défaut.")
		} else {
			// Si c'est une autre erreur (ex: YAML mal formé), on la retourne
			return nil, fmt.Errorf("erreur lors de la lecture du fichier de configuration : %w", err)
		}
	} else {
		log.Printf("Fichier de configuration chargé : %s", viper.ConfigFileUsed())
	}

	// Étape 6: Démapper (unmarshal) la configuration lue (ou les valeurs par défaut) dans la structure Config
	// Viper va utiliser les tags `mapstructure` pour savoir où mettre chaque valeur
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("impossible de démapper la configuration dans la structure : %w", err)
	}

	// Log pour vérifier la config chargée (utile pour le debug)
	log.Printf("Configuration loaded: Server Port=%d, DB Name=%s, Analytics Buffer=%d, Monitor Interval=%dmin",
		cfg.Server.Port, cfg.Database.Name, cfg.Analytics.BufferSize, cfg.Monitor.IntervalMinutes)

	return &cfg, nil // Retourne la configuration chargée
}
