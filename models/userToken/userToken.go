package userToken

import "time"

type Token struct {
	ID        string    `bson:"_id,omitempty"`
	UserID    string    `bson:"userId"`
	Token     string    `bson:"token"`
	CreatedAt time.Time `bson:"createdAt"`
}
