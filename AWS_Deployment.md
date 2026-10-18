## Part III: Deploying to AWS Using Terraform

For this part of the assignment, I deployed my Product API to AWS using Terraform.

I used the [CS6650_2b_demo](https://github.com/RuidiH/CS6650_2b_demo) repository provided as a starting point and modified it to deploy my Product API instead of the original Albums API.

Terraform handles creating the required AWS resources, building the Docker image, pushing it to Amazon ECR, and running the application on ECS Fargate.

### Terraform File Structure

The Terraform configuration is divided into multiple files and modules to keep the infrastructure organized.

| File / Module     | Description                                                                                         |
| ----------------- | --------------------------------------------------------------------------------------------------- |
| `provider.tf`     | Configures the AWS and Docker providers                                                             |
| `variables.tf`    | Defines configurable values such as AWS region, resource names, container port, and number of tasks |
| `main.tf`         | Connects the Terraform modules and handles building and pushing the Docker image                    |
| `output.tf`       | Displays useful information after deployment, such as the ECS cluster and service names             |
| `modules/network` | Creates the networking resources, including subnets and a security group                            |
| `modules/ecr`     | Creates an ECR repository to store the Docker image                                                 |
| `modules/logging` | Creates a CloudWatch log group to store application logs                                            |
| `modules/ecs`     | Creates the ECS cluster, task definition, and service                                               |

For the ECS task, I used **0.25 vCPU and 512 MiB of memory**, which was sufficient for running the Product API.

### Changes Made to the Starter Repository

I made a few changes to the starter Terraform configuration to deploy my application.

**1. Replaced the Albums API**

I replaced the original Albums API with my Product API implementation inside the `src/` directory.

I also updated the Dockerfile to build and run the Product API.

**2. Fixed the Docker image architecture**

Since I was building the Docker image on a Mac with Apple Silicon, Docker was building an ARM-based image by default.

However, my ECS Fargate task was configured to use the AMD64 architecture.

To avoid compatibility issues, I added the following configuration in `main.tf`:

```hcl
platform = "linux/amd64"
```

This makes sure Docker builds an image that can run on the configured ECS task.

**3. Rebuilding the image when code changes**

I added `triggers` to the Docker image resource so that Terraform can detect changes made to files inside the `src/` directory.

Without this, Terraform might not rebuild the image because the image tag remains `latest`, even when the application code changes.

**4. Managing deployment dependencies**

I added `depends_on` to the ECS module so that Terraform waits for the Docker image to be pushed to ECR before creating the ECS service.

This helps avoid starting an ECS task before the required image is available.

### Prerequisites

Before deploying the application, I installed and configured the following tools:

- [Terraform](https://developer.hashicorp.com/terraform/install)
- [Docker Desktop](https://www.docker.com/products/docker-desktop/)
- [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html)
- AWS Learner Lab credentials

Since I was using AWS Learner Lab, I started the lab and copied the AWS CLI credentials from the **AWS Details** section.

I added these credentials to:

```bash
~/.aws/credentials
```

To verify that the credentials were working, I ran:

```bash
aws sts get-caller-identity
```

One thing to keep in mind is that AWS Learner Lab credentials expire after each session. If I received an `ExpiredToken` error, I had to update the credentials before running Terraform again.

I used the `us-west-2` (Oregon) region for deployment, which is configured in `terraform/variables.tf`.

### Deploying the Application

Once the AWS credentials and Docker were configured, I deployed the application using Terraform.

First, clone the repository and navigate to the Terraform directory:

```bash
git clone <your-fork-url>
cd <repo-name>/terraform
```

Initialize Terraform:

```bash
terraform init
```

This downloads the required providers and initializes the Terraform configuration.

Next, deploy the infrastructure:

```bash
terraform apply
```

Terraform displays the resources it plans to create. I entered `yes` to confirm the deployment.

During deployment, Terraform creates the required networking resources, ECR repository, CloudWatch log group, and ECS resources.

It also builds the Docker image from `src/`, pushes it to ECR, and starts the Product API on ECS Fargate.

After Terraform finishes, I waited for the ECS task to reach the `RUNNING` state.

### Getting the Public IP of the ECS Task

Since the application runs on ECS Fargate, I did not need to manually create or manage an EC2 instance.

Fargate runs the container using the CPU and memory defined in the ECS task definition.

The deployed task receives a public IP address, which I used to send requests to the Product API.

I used the following command from the `terraform/` directory to retrieve the public IP:

```bash
IP=$(aws ec2 describe-network-interfaces \
  --network-interface-ids $(
    aws ecs describe-tasks \
      --cluster $(terraform output -raw ecs_cluster_name) \
      --tasks $(
        aws ecs list-tasks \
          --cluster $(terraform output -raw ecs_cluster_name) \
          --service-name $(terraform output -raw ecs_service_name) \
          --query 'taskArns[0]' --output text
      ) \
      --query "tasks[0].attachments[0].details[?name=='networkInterfaceId'].value" \
      --output text
  ) \
  --query 'NetworkInterfaces[0].Association.PublicIp' \
  --output text)

echo $IP
```

This command finds the running ECS task, gets its network interface, and retrieves the public IP associated with it.

The public IP can change when a task is replaced or restarted, so I needed to retrieve it again after a new deployment.

### Testing the Deployed API

Once the ECS task was running, I tested the API using `curl`.

For example, to fetch product details:

```bash
curl -i http://$IP:8080/products/1100
```

The API listens on port `8080`, which is also allowed through the security group created by Terraform.

I could also run all the test cases from Part II by updating the host:

```bash
HOST=http://$IP:8080
```

This allowed me to verify that the API was working correctly after deployment to AWS.

### Deploying Code Changes

If I made changes to the Product API, I needed to rebuild the Docker image and update the running ECS task.

After modifying files inside `src/`, I ran:

```bash
terraform apply
```

Since I added triggers to the Docker image resource, Terraform detects the changes, rebuilds the image, and pushes the updated version to ECR.

However, because the image still uses the `latest` tag, the running ECS task does not automatically pull the updated image.

To start a new deployment, I used:

```bash
aws ecs update-service \
  --cluster $(terraform output -raw ecs_cluster_name) \
  --service $(terraform output -raw ecs_service_name) \
  --force-new-deployment
```

This tells ECS to start a new task using the updated Docker image.

After the deployment finishes, I retrieve the new public IP and test the API again.

### Cleaning Up AWS Resources

After completing the deployment and testing, I used Terraform to remove the resources created for the assignment.

```bash
terraform destroy
```

Terraform displays the resources that will be deleted and asks for confirmation.

This is important when using AWS Learner Lab because leaving unnecessary resources running can consume the available AWS credits.

### Issues and Troubleshooting

I also came across a few things that are important to check while deploying the application.

| Issue                      | Cause / Solution                                                                                                 |
| -------------------------- | ---------------------------------------------------------------------------------------------------------------- |
| `ExpiredToken` error       | AWS Learner Lab credentials have expired. Start the lab and update the credentials in `~/.aws/credentials`.      |
| AWS resources not visible  | Check that the AWS Console region is set to `us-west-2` (Oregon).                                                |
| ECS task keeps stopping    | Check the stopped task reason and CloudWatch logs in the ECS Console.                                            |
| `exec format error`        | The Docker image was built for the wrong CPU architecture. Build the image using `linux/amd64`.                  |
| API still returns old data | The request might be going to an older ECS task. Wait for the new deployment and retrieve the updated public IP. |

### Deployment Screenshots

The following screenshots show the ECS task running successfully and the API responding to requests after deployment.

<!-- Add screenshots here -->

<!-- Example:
![ECS Task Running](docs/images/ecs-running.png)
![API Response on ECS](docs/images/ecs-api-response.png)
-->
<img width="1027" height="505" alt="Screenshot 2026-10-17 at 1 20 51 PM" src="https://github.com/user-attachments/assets/890a66bf-6380-40a8-ae4b-418c290d4026" />

<img width="1032" height="614" alt="Screenshot 2026-10-17 at 1 26 33 PM" src="https://github.com/user-attachments/assets/24751644-23b5-4acb-8bea-e0860396a34c" />

<img width="1470" height="799" alt="Screenshot 2026-10-17 at 8 45 52 PM" src="https://github.com/user-attachments/assets/00eda01d-791a-477b-ba7f-163ab0a4d37a" />


