---
title: "Machine Learning"
description: "Detect idle AWS SageMaker endpoints that are billing but not being used."
weight: 60
type: docs
prev: /docs/waste-detection/networking
next: /docs/reporting
---

Machine Learning resources, particularly **Amazon SageMaker** endpoints, are often among the most expensive components of an AWS architecture. It is common for data scientists and developers to deploy endpoints for testing or validation and forget to decommission them once the work is complete.

**AWS Doctor** helps you identify these idle endpoints to prevent unnecessary costs.

## SageMaker Endpoints

The tool currently audits SageMaker real-time endpoints across your account.

### Detection Logic
An endpoint is flagged as **idle** if it meets the following criteria:
1.  **Status**: The endpoint must be in the `InService` state.
2.  **Invocations**: The CloudWatch metric `SageMakerVariantInvocations` (Sum) must be **0** for all variants associated with the endpoint over the last **14 days**.

### Why 14 days?
We use a 14-day lookback window to account for models that might be used for bi-weekly processing or scheduled tasks, ensuring we don't flag resources that are still part of a regular (though infrequent) production workflow.

## How to Run
To run the SageMaker waste detection individually:

```bash
aws-doctor waste sagemaker
```

## Remediation
If an endpoint is flagged as idle:
1.  **Verify**: Check if the model is still needed for future requests or if it was part of a completed experiment.
2.  **Delete**: If the endpoint is no longer required, delete it via the AWS Console or CLI. Remember that SageMaker billing is based on the instance type and the duration the endpoint is active, regardless of whether it processes requests.
3.  **Update**: If you need the model occasionally but not 24/7, consider using **SageMaker Serverless Inference** or **Inference Pipelines** to reduce costs.
