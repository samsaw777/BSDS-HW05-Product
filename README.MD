# CS 6650 HW05 - Product API

For this assignment, I built a Product API for an online store using Go and the Gin framework, following the provided OpenAPI specification (`api.yaml`).

The project covers different stages of developing and running an API, starting with implementing the endpoints, deploying the application to AWS using Terraform, performing load testing with Locust, and finally monitoring its performance using Prometheus and Grafana.

## Repository Structure

```text
.
├── src/              Go server implementation and Dockerfile
├── terraform/        Terraform configuration for AWS deployment
├── locust/           Locust load testing scripts
├── observability/    Prometheus and Grafana configuration
└── docs/images/      Screenshots of test results and monitoring
```

## Running the API Locally

To run the Product API locally:

```bash
cd src
go run .
```

Once the server starts, it can be accessed at `http://localhost:8080`.

For example, to fetch product details:

```bash
curl -i http://localhost:8080/products/1100
```

## Project Overview

### Part II: Product API

**[Product API Implementation](Project_Overview.md)**

I implemented the GET and POST endpoints defined in the OpenAPI specification using Go and Gin.

For storing products, I used an in-memory map protected by `sync.RWMutex` to handle concurrent requests safely. I also added four sample products at startup and implemented the required response codes, including a way to simulate internal server errors.

The documentation explains the implementation decisions and includes `curl` commands to test each response code.

### Part III: AWS Deployment Using Terraform

**[AWS Deployment](AWS_Deployment.md)**

For deployment, I modified the provided Terraform starter repository to run my Product API on AWS ECS Fargate.

Terraform creates the required AWS resources, builds the Docker image, pushes it to Amazon ECR, and deploys the application to ECS.

I also made a few changes to handle Docker image compatibility, rebuild the image when the code changes, and manage dependencies between resources.

The documentation includes the steps to deploy the application, retrieve the ECS task's public IP, test the deployed API, and remove the AWS resources.

### Part IV: Load Testing Using Locust

**[Load Testing](Load_Testing.md)**

I used Locust to test how the deployed API performs when multiple users send requests simultaneously.

The tests included a mix of 80% successful GET requests, 10% POST requests, and 10% requests for products that do not exist.

During stress testing, I observed that throughput increased as the number of users grew. However, once the ECS task reached approximately 100% CPU utilization, the throughput started to level off at around **3,300–3,500 requests per second**, with no unexpected failures.

I also compared `HttpUser` and `FastHttpUser` to understand how different HTTP clients affect load testing.

The documentation contains the test results, comparisons, and observations from these experiments.

### Part V: Observability Using Prometheus and Grafana

**[Observability](Observability.md)**

For monitoring, I added Prometheus metrics and structured JSON logging to the Product API.

I followed the RED method to track request rate, errors, and response duration.

I also configured Prometheus and Grafana using Docker Compose and created a dashboard to monitor the API while running load tests locally.

The documentation explains the monitoring setup, dashboard results, and how metrics and logs can be used to understand the application's behavior.

## Cleaning Up AWS Resources

After completing the deployment and testing, the AWS resources can be removed using Terraform.

```bash
cd terraform
terraform destroy
```

This removes the infrastructure created for the assignment and helps avoid unnecessary AWS resource usage.
