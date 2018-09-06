package controllers

import (
	"encoding/json"
	"fmt"

	"git.wetofu.top/tonychee7000/blackForestBot/consts"
	"git.wetofu.top/tonychee7000/blackForestBot/database"
	"git.wetofu.top/tonychee7000/blackForestBot/models"
)

//GetGameQueue is
func GetGameQueue(ID int64) ([]models.QueueElement, error) {
	var (
		gameQueue []models.QueueElement
		jsonByte  []byte
	)
	if err := database.Redis.Get(
		fmt.Sprintf(consts.GameQueueFormatString, ID),
	).Scan(&jsonByte); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(jsonByte, &gameQueue); err != nil {
		return nil, err
	}
	return gameQueue, nil
}

//UpdateGameQueue is
func UpdateGameQueue(ID int64, gameQueue []models.QueueElement) error {
	jsonByte, err := json.Marshal(gameQueue)
	if err != nil {
		return err
	}
	if err := database.Redis.Set(
		fmt.Sprintf(consts.GameQueueFormatString, ID),
		jsonByte, -1,
	).Err(); err != nil {
		return err
	}
	return nil
}

//ClearGameQueue is
func ClearGameQueue(ID int64) error {
	if err := database.Redis.Del(
		fmt.Sprintf(consts.GameQueueFormatString, ID)).Err(); err != nil {
		return err
	}
	return nil
}

//AddGameQueue is
func AddGameQueue(ID int64, element models.QueueElement) error {
	gameQueue, err := GetGameQueue(ID)
	if err != nil {
		if gameQueue == nil {
			gameQueue = make([]models.QueueElement, 0)
		} else {
			return err
		}
	}
	gameQueue = append(gameQueue, element)
	return UpdateGameQueue(ID, gameQueue)
}

//DelGameQueue is
func DelGameQueue(ID int64, element models.QueueElement) error {
	defer func() { recover() }()
	gameQueue, err := GetGameQueue(ID)
	if err != nil {
		return err
	}
	for i, e := range gameQueue {
		if e.Is(element) {
			gameQueue = append(gameQueue[:i], gameQueue[i+1:]...)
		}
	}
	return UpdateGameQueue(ID, gameQueue)
}
