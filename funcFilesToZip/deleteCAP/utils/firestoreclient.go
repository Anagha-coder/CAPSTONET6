package utils

import (
	"context"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/option"
)

func CreateFirestoreClient() (*firestore.Client, error) {
	ctx := context.Background()

	client, err := firestore.NewClient(ctx, "capstone-408907", option.WithCredentialsFile("./capstone.json"))
	if err != nil {
		return nil, err
	}

	return client, nil
}
