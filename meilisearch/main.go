package main

import (
	"encoding/json"
	"github.com/meilisearch/meilisearch-go"
	"io"
	"log"
	"os"
)

func main() {
	client := meilisearch.NewClient(meilisearch.ClientConfig{
		Host:   "http://10.10.101.123:7700",
		APIKey: "HmPEKZhcoANzRt3DTPhRQVRxPEVZw7m2TymKwKhs_6xdnky",
	})
	//curl -X GET 'http://localhost:7700/tasks' -H 'Authorization: Bearer HmPEKZhcoANzRt3DTPhRQVRxPEVZw7m2TymKwKhs_6xdnky'
	//client.DeleteIndex("movies")

	jsonFile, _ := os.Open("./movies.json")
	defer jsonFile.Close()

	byteValue, _ := io.ReadAll(jsonFile)
	var movies []map[string]interface{}
	json.Unmarshal(byteValue, &movies)

	_, err := client.Index("movies").AddDocuments(movies)
	if err != nil {
		panic(err)
	}
	log.Println("success")
}
