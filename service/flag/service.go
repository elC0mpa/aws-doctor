// Package flag provides a service for parsing CLI flags.
package flag

import (
	"flag"

	"github.com/elC0mpa/aws-doctor/model"
)

// NewService creates a new Flag service.
func NewService() Service {
	return &service{}
}

func (s *service) GetParsedFlags(args []string) (model.Flags, error) {
	fs := flag.NewFlagSet("aws-doctor", flag.ContinueOnError)

	region := fs.String("region", "", "AWS region (defaults to AWS_REGION, AWS_DEFAULT_REGION, or ~/.aws/config)")
	profile := fs.String("profile", "", "AWS profile configuration")
	trend := fs.Bool("trend", false, "Display a trend report for the last 6 months")
	waste := fs.Bool("waste", false, "Display AWS waste report")
	output := fs.String("output", "table", "Output format: table or json")
	version := fs.Bool("version", false, "Display version information")
	update := fs.Bool("update", false, "Update aws-doctor to the latest version")

	if err := fs.Parse(args); err != nil {
		return model.Flags{}, err
	}

	return model.Flags{
		Region:  *region,
		Profile: *profile,
		Trend:   *trend,
		Waste:   *waste,
		Output:  *output,
		Version: *version,
		Update:  *update,
	}, nil
}
