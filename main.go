package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/officialevsty/rabbitmq_producer/rabbitmq"
	"github.com/officialevsty/rabbitmq_producer/s3"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type PublishScreenshot struct {
	ID     uuid.UUID `json:"id"`
	S3Data S3Data    `json:"s3"`
}

type S3Data struct {
	Key    string `json:"key"`
	Bucket string `json:"bucket"`
	S3Name string `json:"s3_name"`
}

func readFiles(dir string) ([]string, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	extensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".bmp":  true,
		".gif":  true,
		".webp": true,
		".tiff": true,
	}
	var imageFiles []string
	for _, file := range files {
		if !file.IsDir() && extensions[strings.ToLower(filepath.Ext(file.Name()))] {
			imageFiles = append(imageFiles, filepath.Join(dir, file.Name()))
		}
	}
	if len(imageFiles) == 0 {
		return nil, errors.New("no files found in directory")
	}
	return imageFiles, nil
}

func getPreSign(ctx context.Context, c *s3.S3Client, filename string) (string, string, error) {
	url, key, err := c.GeneratePreSignUrl(ctx, os.Getenv("AWS_BUCKET"), filepath.Ext(filename))
	if err != nil {
		return "", "", err
	}
	return url, key, nil
}

func uploadImg(ctx context.Context, url string, filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return err
	}
	size := stat.Size()

	// Build the request
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, io.LimitReader(f, size))
	if err != nil {
		return err
	}
	req.ContentLength = size

	client := &http.Client{
		Transport: &http.Transport{
			DisableCompression:    true,
			ExpectContinueTimeout: 0,
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed: %s\n%s", resp.Status, string(body))
	}
	return nil
}

func publishInRabbitMQ(ctx context.Context, mq *rabbitmq.RabbitMQ, key string) (uuid.UUID, error) {
	data := S3Data{
		Key:    key,
		Bucket: os.Getenv("AWS_BUCKET"),
		S3Name: "selectel",
	}
	publish := &PublishScreenshot{
		ID:     uuid.New(),
		S3Data: data,
	}
	err := mq.Send(ctx, publish, "to_proceed_images")
	if err != nil {
		return uuid.Nil, err
	}
	return publish.ID, nil
}

func main() {
	forever := make(chan bool)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	rabbit, err := rabbitmq.NewRabbitMQ(logger)
	if err != nil {
		panic(err)
	}
	s3Client, err := s3.NewS3Client(logger)
	if err != nil {
		panic(err)
	}
	filenames, err := readFiles("/root/media")
	if err != nil {
		panic(err)
	}
	logger.Info(fmt.Sprintf("%d files found in /media", len(filenames)))
	_, _ = rabbit, s3Client
	go func() {
		for {
			ctx := context.Background()
			randIndex := rand.Intn(len(filenames))
			url, key, err := getPreSign(ctx, s3Client, filenames[randIndex])
			if err != nil {
				logger.Error(err.Error())
				return
			}
			err = uploadImg(ctx, url, filenames[randIndex])
			if err != nil {
				logger.Error(err.Error())
				return
			}
			_, err = publishInRabbitMQ(ctx, rabbit, key)
			if err != nil {
				logger.Error(err.Error())
				return
			}
			logger.Info(fmt.Sprintf("uploaded %s to %s", filenames[randIndex], url))
			logger.Info(fmt.Sprintf("%ds timeout", 15))
			time.Sleep(15 * time.Second)
		}
	}()
	<-forever
}
