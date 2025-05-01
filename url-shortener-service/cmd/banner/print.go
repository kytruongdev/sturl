package banner

import (
	"log"
	"os"
)

// Print prints the banner
func Print() {
	log.New(os.Stdout, "", log.LstdFlags).Printf("Starting project: %s\n", os.Getenv("SERVICE_NAME"))
}
