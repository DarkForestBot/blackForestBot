package consts

import "log"

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Println("Package consts loaded for global consts.")
}
