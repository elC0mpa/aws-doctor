---
title: "Automation & CI/CD"
description: "Integrate AWS Doctor into your automated workflows using JSON output and CI/CD pipelines like GitHub Actions or Jenkins."
weight: 50
type: docs
prev: /docs/waste-detection
---

**AWS Doctor** is designed to be part of a larger ecosystem. With native JSON support, you can integrate it into your automated workflows.

## JSON Output

To get machine-readable data, use the `--output json` flag:

```bash
aws-doctor --waste --output json > report.json
```

### Schema Example
The JSON output provides a structured list of every resource identified as waste, including its ID, creation date, and size.

```json
{
  "account_id": "123456789012",
  "generated_at": "2026-02-09T12:00:00Z",
  "unused_ebs_volumes": [
    {
      "volume_id": "vol-0abcd1234",
      "size": 50,
      "status": "available"
    }
  ],
  "has_waste": true
}
```

---

## Use Cases

### 1. Build Failure on Waste
In your CI/CD pipeline (GitHub Actions, Jenkins, etc.), you can use `jq` to fail the build if the tool detects any waste:

```bash
# Example logic
if aws-doctor --waste --output json | jq -e '.has_waste == true'; then
  echo "Waste detected! Clean up before proceeding."
  exit 1
fi
```

### 2. Custom Dashboards
Pipe the JSON output to an ELK stack, CloudWatch Logs, or a custom database to track your infrastructure health over time.

---

## Zero Configuration
Because the tool relies on standard AWS credentials, it works out-of-the-box in environments like:
- **GitHub Actions Runners** (using `aws-actions/configure-aws-credentials`).
- **GitLab CI** (using pre-configured runner variables).
- **Terraform Post-Apply** hooks to verify deployment hygiene.
