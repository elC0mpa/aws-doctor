package model

import "time"

// S3BucketWasteInfo represents information about an S3 bucket that is considered waste.
type S3BucketWasteInfo struct {
	BucketName   string
	CreationDate time.Time
	Reason       string
}
