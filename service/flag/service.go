// Package flag provides a service for parsing CLI flags.
package flag

import (
	"flag"
	"strings"

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
	waste := fs.Bool("waste", false, "Display AWS waste report (e.g., --waste ec2,s3)")
	output := fs.String("output", "table", "Output format: table or json")
	version := fs.Bool("version", false, "Display version information")
	update := fs.Bool("update", false, "Update aws-doctor to the latest version")

	var wasteChecks []string
	filteredArgs := make([]string, 0, len(args))

	for i := 0; i < len(args); i++ {
		arg := args[i]
		filteredArgs = append(filteredArgs, arg)

		// If the current argument is the waste flag and the next argument is not a flag,
		// we treat the next argument as the list of specific checks to run.
		if (arg == "--waste" || arg == "-waste") && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
			wasteChecks = strings.Split(args[i+1], ",")
			i++ // Consume the next argument so it's not parsed as a positional argument
		}
	}

	if err := fs.Parse(filteredArgs); err != nil {
		return model.Flags{}, err
	}

	return model.Flags{
		Region:      *region,
		Profile:     *profile,
		Trend:       *trend,
		Waste:       *waste,
		WasteChecks: wasteChecks,
		Output:      *output,
		Version:     *version,
		Update:      *update,
	}, nil
}
