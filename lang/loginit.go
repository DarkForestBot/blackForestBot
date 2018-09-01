package lang

import "log"

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Println("Package lang loaded for multilanguage support.")
}
