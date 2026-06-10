# Forms Nexus - Notification Service

Forms Nexus is a serverless backend service developed in **Go**, designed to process, validate, and route web form notifications to multiple channels (Telegram, Email) in a highly available and scalable manner.

In its current version (v4), the system implements an asynchronous architecture based on the **Producer-Consumer Pattern** to guarantee millisecond response times to the end client and provide fault tolerance against third-party service outages or latency.

## 🏗️ AWS Architecture

The service is deployed using **AWS CDK (Cloud Development Kit)** and interacts with the following components:

1. **API Gateway (HTTP API v2):** The entry point that receives the JSON payload from the client or frontend.
2. **Lambda Producer (Ingestion):** A lightweight function responsible for initial request reception. It validates the environment configuration, publishes the payload to the main queue, and immediately returns a `200 OK` code to release the client connection.
3. **Amazon SQS (Main Queue):** Temporarily stores messages (in configurable batches, currently set to 5) to decouple the HTTP reception from the actual message delivery.
4. **Lambda Consumer (Worker):** A function invoked asynchronously by SQS events. It extracts messages from the batch, decodes and validates the JSON, and executes the domain logic to route the notification to the target channel.
5. **Amazon SQS DLQ (Dead-Letter Queue):** If message processing fails after the configured retries in the Consumer, SQS automatically isolates it in this queue for future auditing, preventing data loss.
6. **Integrations:**
   * **Telegram API:** Real-time push notifications.
   * **Amazon SES:** Transactional email delivery.
   * **Amazon DynamoDB:** Historical record of interactions for traceability.
   * **Amazon CloudWatch:** Monitoring and custom metrics logging (Successes and Errors).

## 📁 Project Structure

The source code follows standard Go conventions, clearly separating infrastructure from executable binaries:

```text
/forms-nexus-svc
 ├── /bin                 # Compiled binaries ready for AWS (Ignored in git)
 ├── /cmd
 │    ├── /consumer       # Source code for the Consumer Lambda (Worker / Core logic)
 │    └── /producer       # Source code for the Producer Lambda (API Handler / Ingestion)
 ├── /deployments         # Infrastructure as Code (AWS CDK)
 │    └── deployments.go
 ├── /internal            # Domain logic, validators, and clients (SES, Telegram, Dynamo)
 ├── go.mod               # Dependency management
 └── README.md