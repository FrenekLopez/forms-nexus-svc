# Forms Nexus - Serverless Notification API

High-performance backend service built with **Go (Hexagonal Architecture)** and deployed on AWS using **Infrastructure as Code (AWS CDK)**.

This service receives JSON payloads from web forms, strictly validates the data, sends notifications through Amazon SES, and maintains an immutable record in Amazon DynamoDB.

## 🏗️ Cloud Architecture
* **AWS API Gateway:** Public entry point and request router.
* **AWS Lambda (Go/ARM64):** Processing and validation engine.
* **Amazon DynamoDB:** Persistent storage for processed records.
* **Amazon SES:** Email delivery engine.

## 🛠️ Prerequisites
* Go 1.21+
* AWS CLI configured (`aws configure`)
* AWS CDK v2 installed (`npm install -g aws-cdk`)
* A verified email address in Amazon SES (within the deployment region).

## 🚀 Deployment Guide (Windows / PowerShell)

To deploy this infrastructure to your AWS account, execute the following commands in order.

**Important:** You must inject the verified Amazon SES email into your terminal before building the infrastructure.

```powershell
# 1. Configure the source/destination email (Secret Variable)
$env:SECRET_APP_EMAIL="your_verified_email@example.com"

# 2. Build the Go executable for AWS Linux ARM64 servers
$env:GOOS="linux"
$env:GOARCH="arm64"
go build -tags lambda.norpc -o bin/bootstrap cmd/forms-nexus-svc/main.go

# 3. Restore the local environment back to Windows for CDK compilation
$env:GOOS="windows"
$env:GOARCH="amd64"

# 4. Deploy the infrastructure
cd deployments
cdk deploy