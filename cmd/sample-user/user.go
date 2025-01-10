package main

type User struct {
	ID                 int64  `json:"user_id"`
	Flags              int64  `json:"flags"`
	Name               string `json:"name" crud:"lenmin:0 lenmax:50"`
	Email              string `json:"email" crud:"req"`
	Password           string `json:"password" crud:"hidden password"`
	EmailActivationKey string `json:"email_activation_key" crud:"hidden"`
	CreatedAt          int64  `json:"created_at"`
	CreatedBy          int64  `json:"created_by"`
	LastModifiedAt     int64  `json:"last_modified_at"`
	LastModifiedBy     int64  `json:"last_modified_by"`
}

type User_Create struct {
	ID       int    `json:"user_id"`
	Name     string `json:"name" crud:"req lenmin:2 lenmax:50"`
	Email    string `json:"email" crud:"req"`
	Password string `json:"password" crud:"req password"`
}

type User_Update struct {
	ID    int    `json:"user_id"`
	Name  string `json:"name" crud:"req lenmin:2 lenmax:50"`
	Email string `json:"email" crud:"req"`
}

type User_UpdatePassword struct {
	ID       int    `json:"user_id"`
	Password string `json:"password" crud:"req password"`
}

type User_List struct {
	ID   int    `json:"user_id"`
	Name string `json:"name"`
}
