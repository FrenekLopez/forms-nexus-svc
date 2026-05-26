# Forms Nexus Service 🚀

Forms Nexus Service is a Serverless API built in **Go** that acts as a smart notification router. Initially designed to process messages from web forms (such as portfolios), it captures HTTP requests and asynchronously dispatches them to messaging channels like Telegram.

## 🏗️ Architecture and Technologies

The entire infrastructure is defined and managed using the **AWS Cloud Development Kit (CDK)**, ensuring consistent and reproducible deployments without the need to manually configure resources in the AWS Console.

* **Language:** Go (Golang)
* **Infrastructure as Code (IaC):** AWS CDK v2
* **Compute:** AWS Lambda (ARM64 / AL2023 Architecture)
* **Gateway:** Amazon API Gateway (HTTP API)
* **Database:** Amazon DynamoDB (Pay-Per-Request billing)
* **Permissions:** Integrated IAM Policies for Amazon SES and DynamoDB
* **Design Patterns:** Interface implementation (Polymorphism) for seamless scalability of notification channels.

## ⚙️ Prerequisites

To deploy this project to your own AWS account, you need the following installed:
* [Go](https://golang.org/doc/install) (v1.21 or higher)
* [Node.js](https://nodejs.org/) (Required by AWS CDK)
* [AWS CLI](https://aws.amazon.com/cli/) configured with your credentials
* [AWS CDK CLI](https://docs.aws.amazon.com/cdk/v2/guide/cli.html) (`npm install -g aws-cdk`)

## 🔑 Environment Variables

The project requires you to inject certain variables into your local environment (terminal) prior to deployment. AWS CDK will fetch these values and securely inject them into the Lambda function:

| Local Variable | Description |
| :--- | :--- |
| `TELEGRAM_BOT_TOKEN` | Access token provided by BotFather on Telegram. |
| `TELEGRAM_CHAT_ID` | Numeric ID of the chat or group where alerts will be delivered. |
| `SECRET_APP_EMAIL` | Verified email address in Amazon SES that will act as the sender and recipient for email routing. |
| `CDK_DEFAULT_ACCOUNT` | (Optional if using AWS profiles) Your AWS Account ID. |
| `CDK_DEFAULT_REGION` | (Optional if using AWS profiles) The region where the stack will be deployed (e.g., `us-east-2`). |

**Note on auto-managed resources:**
Critical infrastructure variables, such as `DYNAMODB_TABLE_NAME`, do not require manual configuration. AWS CDK dynamically resolves these references during deployment (via `formsTable.TableName()`), ensuring exact coupling between the generated resources.

## 🚀 Deployment

1. Clone the repository and navigate to the `deployments` directory:
   ```bash
   cd deployments