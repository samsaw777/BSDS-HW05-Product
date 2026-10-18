## Part IV: Load Testing Using Locust

For this part of the assignment, I used Locust to load test my Product API deployed on AWS ECS Fargate.

The main goal was to see how the API performs when multiple users send requests at the same time, how much traffic a single ECS task can handle, and what happens when the number of users increases.

I also compared Locust's `HttpUser` and `FastHttpUser` to see whether the type of HTTP client affects the test results.

### Setting Up Locust

I created a separate `locust/` directory for the load testing scripts.

To install Locust, I created a Python virtual environment and installed the required package.

```bash
cd locust
python3 -m venv .venv
source .venv/bin/activate
pip install locust
```

The load testing logic is written in `locust/locustfile.py`.

I defined the API requests in a common `TaskSet` and reused them across four different user classes. This allowed me to change the HTTP client and wait time without changing the actual requests being sent.

| User Class            | Client Type  | Wait Time     | Purpose                              |
| --------------------- | ------------ | ------------- | ------------------------------------ |
| `ProductHttpUser`     | HttpUser     | 0.5–2 seconds | Normal load testing                  |
| `ProductFastHttpUser` | FastHttpUser | 0.5–2 seconds | Normal load testing                  |
| `StressHttpUser`      | HttpUser     | No wait time  | HTTP client comparison               |
| `StressFastHttpUser`  | FastHttpUser | No wait time  | Stress testing and client comparison |

### Request Distribution

I used three types of requests during load testing.

| Request                           | Weight | Percentage |
| --------------------------------- | ------ | ---------- |
| GET an existing product           | 8      | 80%        |
| POST to update product details    | 1      | 10%        |
| GET a product that does not exist | 1      | 10%        |

I gave more weight to GET requests because, in a typical online store, users are more likely to view products than update their details.

I also included requests for products that do not exist to simulate situations such as incorrect product IDs or broken links.

For these requests, a 404 response is expected, so I configured Locust to count them as successful requests rather than failures.

This way, the failure percentage represents unexpected errors instead of valid 404 responses.

### Running the Load Tests

I used the following command to start Locust:

```bash
locust -f locustfile.py StressFastHttpUser \
  --host http://$IP:8080 \
  --processes 4
```

After starting Locust, I opened the web interface at:

`http://localhost:8089`

I used four Locust processes so that load generation could use multiple CPU cores on my laptop instead of depending on a single process.

All the tests were performed from my laptop in Boston against a single ECS Fargate task deployed in the AWS Oregon region.

The ECS task was configured with:

| Configuration    | Value                 |
| ---------------- | --------------------- |
| AWS Region       | us-west-2 (Oregon)    |
| ECS Tasks        | 1                     |
| CPU              | 0.25 vCPU             |
| Memory           | 512 MiB               |
| Load Generator   | Local laptop (Boston) |
| Locust Processes | 4                     |

Since my laptop and the ECS task were in different regions, network latency also affected the response times.

The network round-trip time was approximately 75–80 ms, which contributed to the minimum response time observed during testing.

## Test 1: Baseline Load Testing

For the first test, I used a wait time between 0.5 and 2 seconds to simulate users making requests with small pauses between them.

I gradually increased the number of users and monitored the requests per second (RPS), response time, and failure rate.

### Results

| Metric               | Observation         |
| -------------------- | ------------------- |
| Concurrent Users     | Approximately 1,000 |
| Requests per Second  | Approximately 785   |
| Median Response Time | 98 ms               |
| Failure Rate         | 0%                  |
| ECS CPU Usage        | Below 100%          |


### All baseline screenshots

**10 users, 0.5 to 2 s wait time**

<img width="1470" height="800" alt="Screenshot 2026-10-17 at 8 17 33 PM" src="https://github.com/user-attachments/assets/1b8c7036-c9ab-4dae-bb7b-a9e8d9aaece6" />


**50 users, 0.5 to 2 s wait time**

<img width="1470" height="800" alt="Screenshot 2026-10-17 at 8 21 25 PM" src="https://github.com/user-attachments/assets/8fc712db-d21b-4ec5-8eba-0ff01cf9539d" />


**200 users, 0.5 to 2 s wait time**

<img width="1470" height="805" alt="Screenshot 2026-10-17 at 8 30 13 PM" src="https://github.com/user-attachments/assets/48520531-e57c-40e3-9a72-84b96d1ec066" />



**500 users, 0.5 to 2 s wait time**

<img width="1470" height="802" alt="Screenshot 2026-10-17 at 8 34 14 PM" src="https://github.com/user-attachments/assets/92f56728-9dec-4b22-a2b5-54c5eccc5446" />


**1,000 users, 0.5 to 2 s wait time**

<img width="1470" height="802" alt="Screenshot 2026-10-17 at 8 43 04 PM" src="https://github.com/user-attachments/assets/b15cf78d-d21a-4a6c-8826-0f3ebce7f607" />


### Observations

As I increased the number of users, the RPS also increased. However, after reaching approximately 1,000 users, the throughput started to level off at around 785 RPS.

The failure rate remained at 0%, and the median response time was around 98 ms.

Initially, I expected the RPS to keep increasing with more users. However, the wait time between requests was limiting how frequently each simulated user could send requests.

The average wait time was approximately:

```text
(0.5 + 2) / 2 = 1.25 seconds
```

If we also consider approximately 0.12 seconds for each request, one user takes around:

```text
1.25 + 0.12 = 1.37 seconds
```

Therefore, the approximate request rate per user is:

```text
1 / 1.37 = 0.73 requests/second
```

For 1,000 users:

```text
1000 × 0.73 = 730 RPS
```

This is reasonably close to the observed throughput of approximately 785 RPS.

This relationship is explained by Little's Law, which relates the number of concurrent requests, response time, and throughput.

From this test, I understood that the throughput was mainly limited by the wait time configured in Locust rather than the server's processing capacity.

## Test 2: Stress Testing Without Wait Time

For the second test, I removed the wait time between requests.

The goal was to push the API harder and identify the point where the ECS task could no longer handle additional traffic efficiently.

I gradually increased the number of users and monitored the RPS along with CPU and memory utilization in AWS CloudWatch.

For CPU utilization, I used the **Maximum** statistic with a **1-minute period**.

### Results

| Concurrent Users | Requests/Second | ECS CPU Utilization |
| ---------------- | --------------- | ------------------- |
| 50               | ~480            | ~25%                |
| 100              | ~960            | ~40%                |
| 200              | ~1,900          | ~70%                |
| 400              | ~3,200          | ~100%               |
| 800              | ~3,450          | ~100%               |

<!-- ![Stress Test RPS](docs/images/locust-stress-rps.png) -->

<!-- ![CloudWatch CPU and Memory](docs/images/cloudwatch-cpu-memory.png) -->

### Observations

At lower user counts, the throughput increased almost proportionally with the number of users.

For example, increasing the users from 50 to 100 increased the throughput from approximately 480 to 960 RPS.

A similar increase was observed when going from 100 to 200 users.

However, when I reached 400 users, the ECS CPU utilization reached approximately 100%.

At this point, the throughput was around 3,200 RPS.

When I doubled the number of users from 400 to 800, the throughput only increased to approximately 3,450 RPS.

This was an increase of only about 8%, even though the number of users doubled.

The throughput eventually stabilized around **3,300–3,500 RPS**.

Another observation was that the failure rate remained at 0%, even when the CPU utilization reached 100%.

Memory utilization also remained low throughout the experiment.

### What I Learned

From these results, I concluded that CPU was the main bottleneck for this deployment.

The ECS task had only 0.25 vCPU, so once CPU utilization reached its limit, adding more users did not significantly increase the number of requests the server could process.

One way to improve throughput would be to increase the CPU allocated to the ECS task, for example from 0.25 vCPU to 1 vCPU.

Another option would be to run multiple ECS tasks and distribute incoming requests between them using a load balancer.

However, since my API stores products in memory, each ECS task would maintain its own separate product data. A shared database would be needed if all tasks were expected to access the same updated product information.

## Test 3: HttpUser vs FastHttpUser

For the third test, I compared the performance of Locust's `HttpUser` and `FastHttpUser`.

Both tests used the same request distribution and configuration so that the main difference was the HTTP client being used.

### Test Configuration

| Configuration    | Value           |
| ---------------- | --------------- |
| Concurrent Users | 100             |
| Spawn Rate       | 20 users/second |
| Wait Time        | None            |
| Locust Processes | 4               |
| ECS Tasks        | 1               |

### Results

| Metric                | HttpUser | FastHttpUser |
| --------------------- | -------- | ------------ |
| Steady RPS            | ~950     | ~950         |
| Median Response Time  | 96 ms    | 95 ms        |
| 95th Percentile (p95) | 140 ms   | 150 ms       |
| 99th Percentile (p99) | 210 ms   | 180 ms       |
| Failure Rate          | 0%       | 0%           |

<!-- ![HttpUser Results](docs/images/locust-httpuser.png) -->

<!-- ![FastHttpUser Results](docs/images/locust-fasthttpuser.png) -->

### Observations

I expected `FastHttpUser` to generate more requests because it is designed to use fewer CPU resources compared to `HttpUser`.

However, both clients produced almost the same throughput of approximately 950 RPS.

Their response times were also very similar, and neither test produced any failures.

Since there was no wait time between requests, each user sent another request as soon as the previous one finished.

We can estimate the throughput using:

```text
RPS ≈ Number of Users / Average Response Time
```

For 100 users and an average response time of approximately 105 ms:

```text
RPS ≈ 100 / 0.105

RPS ≈ 952
```

This is close to the observed throughput of approximately 950 RPS.

The main reason both clients performed similarly was that network latency was a significant part of the response time.

Since requests were traveling from Boston to Oregon, switching HTTP clients could not reduce the network round-trip time.

### Difference Between HttpUser and FastHttpUser

The main difference between the two clients is how they generate HTTP requests.

**HttpUser:**

- Uses Python's `requests` library.
- Is simple to use and works well for normal load testing.
- Generally requires more CPU resources on the load-generating machine.

**FastHttpUser:**

- Uses `geventhttpclient`.
- Is designed to generate HTTP requests with lower CPU overhead.
- Can be useful when generating a very large number of requests.

In my experiment, the load-generating machine was not reaching its CPU limit at 100 users.

Because of this, the lower CPU overhead of `FastHttpUser` did not provide a noticeable improvement in throughput.

From this comparison, I learned that using a faster HTTP client does not always mean higher RPS. It depends on whether the load generator itself is the bottleneck.

## Tradeoffs and Final Observations

One of the important design choices in my API was using a Go map with `sync.RWMutex` to store and access product information.

Since 80% of my test requests were GET requests, this design worked well for the workload.

A map provides fast lookups using the product ID, while `RWMutex` allows multiple requests to read the data at the same time.

However, write operations require an exclusive lock.

This means that if the workload had more POST requests, the performance could be different because writes would temporarily block other reads and writes.

Overall, the load tests helped me understand how the number of users, request wait time, network latency, and available CPU resources affect API performance.

The main observations from my experiments were:

- With a wait time between requests, throughput was mainly limited by how frequently each simulated user sent requests.
- Without a wait time, throughput increased until the ECS task reached its CPU limit.
- The single ECS task handled approximately 3,300–3,500 RPS before throughput started to level off.
- Memory usage remained low, while CPU utilization reached 100% during stress testing.
- `HttpUser` and `FastHttpUser` showed similar performance because the load generator was not the bottleneck in the comparison test.
- All three tests completed with 0% unexpected request failures.

These experiments also showed that increasing the number of users does not always increase throughput. Once a system reaches its resource limits, additional users mainly increase the load rather than the amount of work the server can complete.
