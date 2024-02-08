package s3Bucket

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func DocUpload(file multipart.File, fileInfo *multipart.FileHeader, filePath string) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
		Credentials: credentials.NewStaticCredentials(
			"AKIAWG57OGN3Z2HRJK6T",                     // Replace with your access key ID
			"toCr1wqXQYejDwhhW+Prr6llKu4Ca9nV0vDVPFg7", // Replace with your secret access key
			""),
	})
	if err != nil {
		fmt.Println("Error creating session:", err)
		return
	}

	// Create a buffer to hold the file contents
	buffer := make([]byte, fileInfo.Size)

	// Read the file contents into the buffer
	_, err = file.Read(buffer)
	if err != nil {
		return
	}

	// Upload the file to S3

	svc := s3.New(sess)
	// fileName := generateFileName(fileInfo)
	fileName := fileInfo.Filename
	ext := filepath.Ext(fileName)
	name := strings.TrimSuffix(fileName, ext)
	newName := fmt.Sprintf("%s_%d%s", name, time.Now().Unix(), ext)
	directoryPath := filePath
	// Upload the file to the S3 bucket
	_, err = svc.PutObject(&s3.PutObjectInput{
		Body:   bytes.NewReader(buffer),
		Bucket: aws.String("data-export-import-zinnext"),
		Key:    aws.String(directoryPath + newName),
	})
	fmt.Println("uploaderror", err)
	if err != nil {
		return
	}
}

func GetFileFromS3(filepath string) (string, error) {
	// Create an AWS session
	// fmt.Println("filename", filename)
	// filename := "Navatha_06-06-2023.pdf"
	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String("us-east-1"),        // Replace with your desired AWS region
		Endpoint:         aws.String("s3.amazonaws.com"), // Replace with the appropriate S3 endpoint for your region
		S3ForcePathStyle: aws.Bool(true),                 // Required when using a custom endpoint
		Credentials: credentials.NewStaticCredentials(
			"AKIAWG57OGN3Z2HRJK6T",                     // Replace with your access key ID
			"toCr1wqXQYejDwhhW+Prr6llKu4Ca9nV0vDVPFg7", // Replace with your secret access key
			""),
	})
	if err != nil {
		fmt.Println("region error")
		return "", err
	}

	// Create an S3 service client
	svc := s3.New(sess)

	// Specify the bucket and key
	bucket := "data-export-import-zinnext" // Replace with your S3 bucket name
	// key := "s3resume/Navatha_06-06-2023.pdf" // Replace with the key (file name) of the uploaded file

	// Generate a pre-signed URL for the object
	req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(filepath),
	})
	url, err := req.Presign(24 * time.Hour) // URL expiration time: 24 hours
	if err != nil {
		fmt.Println("making url error")
		return "", err
	}

	fmt.Println("File URL:", url)
	return url, nil
}

func UploadResumeToS3(resumeData string, filepath string) (string, error) {
	// Create an AWS session
	file, err := os.Open(resumeData)
	if err != nil {
		fmt.Println("open error")
		return "error", err
	}
	defer file.Close()
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"), // Replace with your desired AWS region
		// Add any other necessary AWS configuration options
		Credentials: credentials.NewStaticCredentials(
			"AKIAWG57OGN3Z2HRJK6T",                     // Replace with your access key ID
			"toCr1wqXQYejDwhhW+Prr6llKu4Ca9nV0vDVPFg7", // Replace with your secret access key
			""),
	})
	if err != nil {
		fmt.Println("cred error")
		return "error", err
	}

	// Create an S3 client
	s3Client := s3.New(sess)
	// Prepare the input parameters for uploading the resume to S3
	params := &s3.PutObjectInput{
		Bucket: aws.String("data-export-import-zinnext"),
		Key:    aws.String(filepath),
		Body:   file,
	}

	// Upload the resume to S3
	_, err = s3Client.PutObject(params)
	if err != nil {
		fmt.Println("put error")
		return "error", err
	}
	err = file.Close()
	if err != nil {
		return "close error", err
	}

	fmt.Println("Resume uploaded to S3 successfully")
	return "Resume uploaded to S3 successfully", err
}

func MultiDocsUpload(fileInfos []*multipart.FileHeader, directoryPath string, uniqueName string) error {
	fmt.Println("inside document upload")
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
		Credentials: credentials.NewStaticCredentials(
			"AKIAWG57OGN3Z2HRJK6T",                     // Replace with your access key ID
			"toCr1wqXQYejDwhhW+Prr6llKu4Ca9nV0vDVPFg7", // Replace with your secret access key
			""),
	})
	if err != nil {
		// Handle error creating session
		return err
	}

	svc := s3.New(sess)

	for _, fileHeader := range fileInfos {
		fmt.Println("inside for loop")
		file, err := fileHeader.Open()
		if err != nil {
			// Handle error opening file
			return err
		}

		// Create a buffer to hold the file contents
		buffer := make([]byte, fileHeader.Size)

		// Read the file contents into the buffer
		_, err = file.Read(buffer)
		if err != nil {
			// Handle error reading file contents
			return err
		}
		// Upload the file to the S3 bucket
		_, err = svc.PutObject(&s3.PutObjectInput{
			Body:   bytes.NewReader(buffer),
			Bucket: aws.String("data-export-import-zinnext"),
			Key:    aws.String(directoryPath + uniqueName),
		})
		//fmt.Println("upload error", err)
		if err != nil {
			// Handle error uploading file
			return err
		}

		// Handle successful upload for the file
		fmt.Printf("File %s uploaded successfully\n", uniqueName)
	}
	return err
}
