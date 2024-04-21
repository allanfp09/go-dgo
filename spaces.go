package go_dgo

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gin-gonic/gin"
	"log"
	"mime/multipart"
	"net/http"
)

// newSpacesConnection initialize a new connection with digitalocean
func newSpacesConnection(accessKey, secretKey, endpoint, region string) *s3.S3 {
	// Initialize a session using DigitalOcean Spaces credentials and the desired region
	sess := session.Must(session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
		Endpoint:    aws.String(endpoint),
	}))

	// Create an S3 service client

	return s3.New(sess)
}

// NewSpacesUploader connects to the bucket, keys and region to upload
func NewSpacesUploader(endpoint, region, accessKey, secretKey string, data *multipart.FileHeader) (*s3manager.UploadOutput, error) {

	// new client session
	clientSess := newSpacesConnection(accessKey, secretKey, endpoint, region)

	// Specify the bucket name and the file you want to upload
	bucketName := "my.posts"
	key := data.Filename + ".jpg"

	// Open the file
	file, err := data.Open()
	if err != nil {
		fmt.Println("Failed to open file", err)
		return nil, err
	}
	defer file.Close()

	// Upload the file to the specified bucket
	uploader := s3manager.NewUploaderWithClient(clientSess)
	info, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
		Body:   file,
	})
	if err != nil {
		fmt.Println("Failed to upload file", err)
		return nil, err
	}

	return info, nil
}

// DeleteFromBucket deletes files from bucket
func DeleteFromBucket(accessKey, secretKey, endpoint, region, bucket, object string) error {

	clientSess := newSpacesConnection(accessKey, secretKey, endpoint, region)

	// Prepare input parameters for the DeleteObject API
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	}

	// Delete the object
	_, err := clientSess.DeleteObject(input)
	if err != nil {
		log.Fatalf("Failed to delete object %s from bucket %s: %v", object, bucket, err)
		return err
	}

	log.Printf("object %s deleted successfully from bucket!", object)

	return nil

}

// uploadMultipleFiles uploads files to the corresponding endpoint
func uploadMultipleFiles(c *gin.Context, endpoint, region, accessKey, secretKey string) ([]string, error) {
	form, err := c.MultipartForm()
	if err != nil {
		return nil, err
	}

	files := form.File["files"]

	var imgLoc []string

	if len(files) != 0 {
		for _, file := range files {
			src, err := file.Open()
			if err != nil {
				c.String(http.StatusInternalServerError, "error opening file!")
				return nil, err
			}
			defer src.Close()

			info, err := NewSpacesUploader(endpoint, region, accessKey, secretKey, file)
			if err != nil {
				return nil, err
			}

			imgLoc = append(imgLoc, info.Location)

		}
	} else {
		return nil, err
	}

	log.Printf("files uploaded successfully!")

	return imgLoc, nil
}
