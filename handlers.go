package main

type modifiedUser struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
	Token string `json:"token"`
}

type expectedStruct struct {
	Id   int    `json:"id"`
	Body string `json:"body"`
}
