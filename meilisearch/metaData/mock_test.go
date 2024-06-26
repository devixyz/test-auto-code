package metaData

import (
	"github.com/meilisearch/meilisearch-go"
	"log"
	"testing"
)

var Mclient *meilisearch.Client

func init() {
	Mclient = meilisearch.NewClient(meilisearch.ClientConfig{
		Host:   "http://10.10.101.123:7700",
		APIKey: "HmPEKZhcoANzRt3DTPhRQVRxPEVZw7m2TymKwKhs_6xdnky",
	})
}

func Test_Delete_All_Data(t *testing.T) {
	resp, err := Mclient.DeleteIndex("documents")
	if err != nil {
		t.Error(err)
	}
	log.Println(resp)
}
