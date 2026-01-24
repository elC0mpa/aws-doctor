package flag

import (
	"flag"

	"github.com/elC0mpa/aws-doctor/model"
)

// NewService creates a new Flag service.
func NewService() *service {
	return &service{}
}

func (s *service) GetParsedFlags() (model.Flags, error) {
	region := flag.String("region", "", "AWS region (defaults to AWS_REGION, AWS_DEFAULT_REGION, or ~/.aws/config)")
	profile := flag.String("profile", "", "AWS profile configuration")
	trend := flag.Bool("trend", false, "Display a trend report for the last 6 months")
	waste := flag.Bool("waste", false, "Display AWS waste report")
	output := flag.String("output", "table", "Output format: table or json")
	flag.Bool("version", false, "Display version information")

	flag.Parse()

	return model.Flags{
		Region:  *region,
		Profile: *profile,
		Trend:   *trend,
		Waste:   *waste,
		Output:  *output,
	}, nil
}
