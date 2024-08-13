package database

import (
	"encoding/json"
	"fmt"

	"github.com/arpitpandey992/go-mpd/internal/config"
	"github.com/meilisearch/meilisearch-go"
)

type AudioMeilisearchClient struct {
	client *meilisearch.Client
	index  *meilisearch.Index
}

func GetNewAudioMeiliSearchClient(config *config.Config) *AudioMeilisearchClient {
	client := meilisearch.NewClient(meilisearch.ClientConfig{
		Host: fmt.Sprintf("http://%s:%d", config.Database.Meilisearch.Host, config.Database.Meilisearch.Port),
	})
	index := client.Index(config.Database.Meilisearch.IndexName)

	return &AudioMeilisearchClient{
		client: client,
		index:  index,
	}
}

func (amc *AudioMeilisearchClient) SearchAudioFiles(searchKey string, limit int64) ([]AudioFileMetadata, error) {
	searchRes, err := amc.index.Search(searchKey,
		&meilisearch.SearchRequest{
			Limit: limit,
		})

	if err != nil {
		return nil, err
	}

	var audioMetadataList []AudioFileMetadata
	for _, hit := range searchRes.Hits {
		var metadata AudioFileMetadata
		hitBytes, err := json.Marshal(hit)
		if err != nil {
			fmt.Println("Error marshalling hit:", err)
			continue
		}
		err = json.Unmarshal(hitBytes, &metadata)
		if err != nil {
			fmt.Println("Error unmarshalling to AudioFileMetadata:", err)
			continue
		}
		audioMetadataList = append(audioMetadataList, metadata)
	}

	return audioMetadataList, nil
}
