package main

type Message struct {
	Id        int32  `json:"id"`
	From      int32  `json:"from"`
	Timestamp int32  `json:"ts"`
	Message   string `json:"msg"`
}

type User struct {
	Id        int32  `json:"user_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	Phone     string `json:"phone"`
}

type Dialogue struct {
	Count    int32     `json:"message_count"`
	Messages []Message `json:"messages"`
	Users    []User    `json:"users"`
}
