package utils

import "log"

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Println("Package utils loaded for some support staff.")
}
