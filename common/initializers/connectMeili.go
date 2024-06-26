package initializers

import (
	"github.com/Arxtect/Einstein/config"
	"github.com/meilisearch/meilisearch-go"
)

var MeiliClient *meilisearch.Client

func InitMeiliClient(config *config.Config) {
	MeiliClient = meilisearch.NewClient(meilisearch.ClientConfig{
		Host:   "http://10.10.101.123:7700",
		APIKey: "HmPEKZhcoANzRt3DTPhRQVRxPEVZw7m2TymKwKhs_6xdnky",
	})
}
