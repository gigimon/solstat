package web

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gigimon/solstat/pkg/processor"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type App struct {
	DB *mongo.Database
}

func (app *App) GetBlocks(w http.ResponseWriter, r *http.Request) {
	countQ := r.URL.Query().Get("count")
	if countQ == "" {
		countQ = "20"
	}
	count, err := strconv.Atoi(countQ)
	if err != nil {
		count = 20
	}

	if count > 100 {
		count = 100
	} else if count < 1 {
		count = 1
	}

	filter := bson.D{}
	opts := options.Find().SetSort(bson.D{{"slot", -1}}).SetLimit(int64(count))
	ctx := r.Context()
	cursor, err := app.DB.Collection("blocks").Find(ctx, filter, opts)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)
	var blocks []processor.ServerProcessedBlock
	if err := cursor.All(ctx, &blocks); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	b, err := json.Marshal(blocks)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(b)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
