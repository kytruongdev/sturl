package banner

import (
	"log"
	"os"

	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/env"
)

// Print prints the banner
func Print() {
	log.New(os.Stdout, "", log.LstdFlags).Printf("Started %v\n", env.GetAndValidateF("SERVICE_NAME"))
}
