package model

// ECRNoLifecyclePolicyInfo represents an ECR repository missing a lifecycle policy.
type ECRNoLifecyclePolicyInfo struct {
	RepositoryName string `json:"repository_name"`
}

// ECREmptyRepositoryInfo represents an ECR repository with zero images.
type ECREmptyRepositoryInfo struct {
	RepositoryName string `json:"repository_name"`
}

// ECRUntaggedImageInfo represents an ECR repository containing untagged (dangling) images.
type ECRUntaggedImageInfo struct {
	RepositoryName       string  `json:"repository_name"`
	UntaggedImageCount   int     `json:"untagged_image_count"`
	UntaggedSizeBytes    int64   `json:"untagged_size_bytes"`
	EstimatedMonthlyCost float64 `json:"estimated_monthly_cost"`
}
