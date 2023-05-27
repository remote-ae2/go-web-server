package main

type Item struct {
	Id    int    `json:"i"` //id
	Count int    `json:"c"` //count
	Name  string `json:"n"` //name
}
type StorageAE2 struct {
	ID    string `json:"id"`
	Items []Item `json:"items"`
}

//--------------------------------------------------------
type Message struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

func createSessionStartMessage() interface{} {
	msg := &Message{}
	msg.Type = "SessionStart"
	return msg
}
func createWaitingMessage() interface{} {
	msg := &Message{}
	msg.Type = "Waiting"
	return msg
}
