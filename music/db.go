package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"regexp"
	"time"
)

// MongoDB DOC struct
type MusicDoc struct {
	ID       string    `bson:"_id"`
	ArtistsB []ArtistJ `bson:"artists"`
}

// Inserts var ArtistsJ to database
func UpdateIndex(path string) {
	ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)
	col := Client.Database("home_server").Collection("music")

	if count, err := col.CountDocuments(ctx, bson.M{"_id": "0"}); err == nil && count == 0 {
		doc := MusicDoc{"0", ArtistsJ}
		_, err := col.InsertOne(ctx, doc)

		if err != nil {
			fmt.Println(err)
		}
	} else {
		fmt.Println("Document exists")

		reg := regexp.QuoteMeta(path) + ".*"

		ctx2, _ := context.WithTimeout(context.Background(), 15*time.Second)
		_, err := col.UpdateOne(ctx2, bson.M{"_id": "0"}, bson.M{"$pull": bson.M{"artists.$[].albums.$[].tracks": bson.M{"path": bson.M{"$regex": reg}}}})
		if err != nil {
			fmt.Println(err)
		}

		ctx2, _ = context.WithTimeout(context.Background(), 15*time.Second)
		_, err = col.UpdateOne(ctx2, bson.M{"_id": "0"}, bson.M{"$pull": bson.M{"artists": bson.M{"albums": bson.A{}}}})
		if err != nil {
			fmt.Println(err)
		}

		doc := MusicDoc{"0", ArtistsJ}
		_, err = col.InsertOne(ctx, doc)

		if err != nil {
			fmt.Println(err)
		}
	}
}
