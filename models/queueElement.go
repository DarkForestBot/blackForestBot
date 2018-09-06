package models

//QueueElement is
type QueueElement struct {
	UserID    int64 `json:"userId"`
	MessageID int   `json:"messageId"`
}

//NewQueueElement is
func NewQueueElement(userID int64, messageID int) QueueElement {
	return QueueElement{
		UserID:    userID,
		MessageID: messageID,
	}
}

//Is is
func (q QueueElement) Is(q2 QueueElement) bool {
	return q.UserID == q2.UserID && q.MessageID == q2.MessageID
}
