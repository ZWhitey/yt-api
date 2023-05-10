package types

import "go.mongodb.org/mongo-driver/mongo"

type Handler struct {
	db *mongo.Database
}
