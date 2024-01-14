package utils

import (
	"context"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

func CreateStorageClient() (*storage.Client, error) {
	ctx := context.Background()

	storageClient, err := storage.NewClient(ctx, option.WithCredentialsFile("./capstone.json"))
	if err != nil {
		return nil, err
	}

	return storageClient, nil
}
