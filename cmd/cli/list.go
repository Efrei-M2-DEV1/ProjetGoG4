package cli

import (
	"fmt"
	"log"

	cmd2 "github.com/axellelanca/urlshortener/cmd"
	"github.com/axellelanca/urlshortener/internal/repository"
	"github.com/spf13/cobra"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// ListCmd représente la commande 'list'
var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "Affiche la liste de tous les liens raccourcis.",
	Long: `Cette commande affiche tous les liens raccourcis enregistrés dans la base de données
avec leur code court, leur URL longue et leur date de création.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Charger la configuration
		cfg := cmd2.Cfg
		if cfg == nil {
			log.Fatalf("FATAL: Configuration non chargée")
		}

		// Initialiser la connexion à la BDD
		db, err := gorm.Open(sqlite.Open(cfg.Database.Name), &gorm.Config{})
		if err != nil {
			log.Fatalf("FATAL: Échec de la connexion à la base de données: %v", err)
		}

		sqlDB, err := db.DB()
		if err != nil {
			log.Fatalf("FATAL: Échec de l'obtention de la base de données SQL sous-jacente: %v", err)
		}

		defer func() {
			if err := sqlDB.Close(); err != nil {
				log.Printf("Erreur lors de la fermeture de la connexion: %v", err)
			}
		}()

		// Initialiser le repository
		linkRepo := repository.NewLinkRepository(db)

		// Récupérer tous les liens
		links, err := linkRepo.GetAllLinks()
		if err != nil {
			log.Fatalf("FATAL: Erreur lors de la récupération des liens: %v", err)
		}

		// Afficher les résultats
		if len(links) == 0 {
			fmt.Println("Aucun lien trouvé dans la base de données.")
			return
		}

		fmt.Printf("Liste des liens (%d total):\n\n", len(links))
		for i, link := range links {
			fmt.Printf("%d. Code: %s\n", i+1, link.ShortCode)
			fmt.Printf("   URL longue: %s\n", link.LongURL)
			fmt.Printf("   URL courte: %s/%s\n", cfg.Server.BaseURL, link.ShortCode)
			fmt.Printf("   Créé le: %s\n\n", link.CreatedAt.Format("2006-01-02 15:04:05"))
		}
	},
}

func init() {
	// Ajouter la commande à RootCmd
	cmd2.RootCmd.AddCommand(ListCmd)
}

