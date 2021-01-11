package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type MusicDoc struct {
	ID       primitive.ObjectID `bson:"_id"`
	ArtistsB []ArtistJ          `bson:"artists"`
}

func InsertIntoDB() {
	ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)
	col := Client.Database("home_server").Collection("music")

	_, err := col.InsertOne(ctx, map[string][]ArtistJ{"artists": ArtistsJ})

	if err != nil {
		fmt.Println(err)
	}
}
