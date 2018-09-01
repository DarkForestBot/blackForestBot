package database

import "log"

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Println("Package database loaded for connect to database and redis.")
}
