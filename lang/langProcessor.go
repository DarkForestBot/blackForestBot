package lang

import (
	"bytes"
	"fmt"
	"log"
	"text/template"

	"git.wetofu.top/tonychee7000/blackForestBot/basis"
	"git.wetofu.top/tonychee7000/blackForestBot/consts"
	"git.wetofu.top/tonychee7000/blackForestBot/database"
	"git.wetofu.top/tonychee7000/blackForestBot/models"
)

func init() {
	var users []models.User
	var groups []models.TgGroup
	if err := database.DB.Find(&users).Error; err != nil {
		panic(err)
	}
	if err := database.DB.Find(&groups).Error; err != nil {
		panic(err)
	}
	for i, user := range users {
		if err := database.Redis.Set(
			fmt.Sprintf(consts.LangSetFormatString, user.TgUserID),
			user.Language, -1,
		).Err(); err != nil {
			panic(err)
		}
		log.Printf("Load user language set: %d/%d", i+1, len(users))
	}
	for i, group := range groups {
		if err := database.Redis.Set(
			fmt.Sprintf(consts.LangSetFormatString, group.TgGroupID),
			group.Lang, -1,
		).Err(); err != nil {
			panic(err)
		}
		log.Printf("Load group language set: %d/%d", i+1, len(groups))
	}
}

// T set lang and return real info
func T(langset, key string, args interface{}) string {
	language, ok := basis.GlobalLanguageList[langset]
	if !ok {
		log.Printf("WARNING: no language set `%s` found.", langset)
		return fmt.Sprintf(consts.LangMissingFormatString, key)
	}

	if args == nil {
		str, ok := language[key]
		if !ok {
			log.Printf("WARNING: no language key `%s` found.", key)
			return fmt.Sprintf(consts.LangMissingFormatString, key)
		}
		return str
	}

	t, err := template.New(key).Parse(language[key])
	if err != nil {
		log.Printf("ERROR: %v\n", err)
		return fmt.Sprintf(consts.LangMissingFormatString, key)
	}
	s := new(bytes.Buffer)
	err = t.Execute(s, args)
	if err != nil {
		log.Printf("ERROR: %v\n", err)
		return fmt.Sprintf(consts.LangMissingFormatString, key)
	}
	return s.String()

}
