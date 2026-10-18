## Part V: Observability Using Prometheus and Grafana

For the last part of the assignment, I added monitoring and logging to my Product API using Prometheus, Grafana, and Go's built-in logging package.

The goal was to understand how the API behaves while handling requests and to monitor important information such as request rate, errors, and response time.

For this, I followed the **RED method**, which focuses on three metrics:

- **Rate:** How many requests the server receives per second.
- **Errors:** How many requests result in errors.
- **Duration:** How long the server takes to process requests.

I implemented these using Prometheus metrics and structured JSON logs.

### Implementing Prometheus Metrics

I added a middleware in `src/metrics.go` to collect metrics for incoming API requests.

The metrics are exposed through the `/metrics` endpoint, which Prometheus uses to collect monitoring data.

I implemented the following metrics:

| RED Signal | Metric Name                     | Type      | Description                                 |
| ---------- | ------------------------------- | --------- | ------------------------------------------- |
| Rate       | `http_requests_total`           | Counter   | Counts the total number of HTTP requests    |
| Errors     | `http_errors_total`             | Counter   | Counts requests with non-2xx response codes |
| Duration   | `http_request_duration_seconds` | Histogram | Records how long each request takes         |

The metrics also contain labels such as the HTTP method, endpoint, and response status.

### Implementation Decisions

I made a few decisions while implementing the monitoring middleware.

**1. Using route patterns instead of actual URLs**

For the route label, I used the endpoint pattern instead of the actual product ID.

For example:

```text
/products/:productId
```

Instead of:

```text
/products/1100
/products/1200
/products/1300
```

This prevents Prometheus from creating a separate time series for every product ID, which would increase the number of metrics being stored.

**2. Excluding the metrics endpoint**

I excluded requests to `/metrics` from the collected metrics.

Since Prometheus regularly calls this endpoint to collect data, including those requests would increase the request count even when no actual users were accessing the API.

**3. Using smaller histogram buckets**

I configured smaller histogram buckets starting at 0.5 milliseconds.

The default Prometheus histogram buckets start at 5 milliseconds, but my API processes requests much faster when running locally.

Using smaller buckets helps capture the response times more accurately.

**4. Handling errors in the middleware**

I placed the metrics and logging middleware outside the recovery middleware.

This was important because I implemented simulated 500 errors using panics.

If the metrics middleware was placed inside the recovery middleware, the panic could skip the remaining metrics code.

By placing it outside, I made sure that even requests resulting in a 500 error were included in the metrics.

### Implementing Structured Logging

I also added structured logging in `src/logging.go` using Go's built-in `log/slog` package.

Instead of using Gin's default text-based logger, I configured the application to generate logs in JSON format.

Each request produces a log entry containing information such as:

- Requested endpoint
- HTTP response status
- Request duration

This makes it easier to understand what happened during a request and investigate errors.

### Running Prometheus and Grafana

For this part, I ran the monitoring stack locally using Docker Compose instead of deploying it to AWS.

The configuration files are available in the `observability/` directory.

To start the stack, I used:

```bash
cd observability
docker compose up --build
```

This starts the Product API, Prometheus, and Grafana.

Once the containers are running, the services can be accessed at:

| Service     | URL                   |
| ----------- | --------------------- |
| Product API | http://localhost:8080 |
| Prometheus  | http://localhost:9090 |
| Grafana     | http://localhost:3000 |

Prometheus is configured to collect metrics from the Product API every **5 seconds**.

This configuration is defined in:

```text
observability/prometheus.yml
```

### Configuring Grafana

After starting the containers, I opened Grafana at:

```text
http://localhost:3000
```

I logged in using the default credentials:

```text
Username: admin
Password: admin
```

Next, I connected Grafana to Prometheus.

I navigated to:

**Connections → Data Sources → Add Data Source → Prometheus**

I used the following URL:

```text
http://prometheus:9090
```

After clicking **Save & Test**, Grafana was able to connect to Prometheus.

I then imported my dashboard configuration from:

```text
observability/grafana-dashboard.json
```

The dashboard displays the RED metrics collected from the Product API.

I also verified that Prometheus was successfully collecting metrics by opening:

```text
http://localhost:9090/targets
```

The `product-service` target showed the status `UP`.

### Running Load Tests with Monitoring

Once Prometheus and Grafana were running, I used Locust to generate traffic against the locally running Product API.

I ran the following command:

```bash
cd locust

locust -f locustfile.py StressFastHttpUser \
  --host http://localhost:8080 \
  --processes 4
```

For this experiment, I used:

| Configuration    | Value                   |
| ---------------- | ----------------------- |
| Concurrent Users | 50                      |
| Test Duration    | Approximately 5 minutes |
| Wait Time        | None                    |
| Locust Processes | 4                       |
| Target           | Local Product API       |

The purpose of this experiment was to observe how the metrics changed when the server went from being idle to handling multiple requests.

### Grafana Dashboard Results

I used the Grafana dashboard to compare the API's behavior before, during, and after the load test.

The server was initially idle, then handled requests generated by Locust, and returned to an idle state after the test finished.

<!-- Add Grafana dashboard screenshot here -->

<!-- ![Grafana Dashboard](docs/images/grafana-idle-vs-loaded.png) -->

#### 1. Request Rate

Before starting Locust, the request rate was close to zero.

Once the load test started, the request rate increased significantly.

The dashboard showed approximately:

| Request Type | Requests per Second |
| ------------ | ------------------- |
| GET          | ~3,100              |
| POST         | ~350                |

This was expected because I configured Locust to generate more GET requests than POST requests.

The request distribution was approximately 90% GET and 10% POST, including both successful and intentionally unsuccessful GET requests.

After stopping Locust, the request rate returned to almost zero.

#### 2. Error Rate

The dashboard showed an error rate of approximately **11.1% for GET requests**.

These errors were mainly caused by the intentionally generated requests for products that did not exist.

In my Locust configuration, I assigned:

- Weight 8 to GET requests for existing products.
- Weight 1 to GET requests for missing products.
- Weight 1 to POST requests.

This means that out of every 9 GET requests, approximately 1 request is expected to return 404.

Therefore:

```text
GET Error Rate = 1 / 9 × 100

GET Error Rate ≈ 11.1%
```

This matched the error percentage displayed in Grafana.

The POST endpoint did not show any errors during this experiment.

These 404 responses were expected and did not indicate that the server was failing.

#### 3. Request Duration

I also monitored how long the server took to process requests.

During the load test, the dashboard showed approximately:

| Metric       | Response Time |
| ------------ | ------------- |
| p50 (Median) | 250 µs        |
| p95          | 480 µs        |

The response times were much lower compared to my AWS ECS load tests.

During the ECS experiments, the median response time was approximately 95 ms because the requests were traveling from Boston to the AWS Oregon region.

In this experiment, both Locust and the Product API were running locally, so there was no long-distance network round trip.

One thing to consider is that the histogram buckets started at 0.5 ms.

Since many requests completed faster than this, the percentile values shown by Grafana were estimates based on the histogram buckets.

Using smaller buckets would provide more precise measurements for these very short response times.

### Request Logs

Along with the Grafana dashboard, I also collected structured JSON logs generated by the Product API during the load test.

Each log entry contains information about the request, including the endpoint, response status, and duration.

<!-- Add sample request logs here later -->

```json

```

### Comparing Grafana Metrics and Logs

Both Grafana and structured logging helped me understand the behavior of the API, but they were useful for different purposes.

Grafana was useful for monitoring the overall performance of the application. It allowed me to see how the request rate, error rate, and response time changed during the load test.

On the other hand, structured logs were useful for checking individual requests. If a request failed, I could look at the logs to find which endpoint was called, what response status was returned, and how long the request took.

From this experiment, I understood that metrics are useful for identifying performance trends, while logs provide more detailed information when investigating specific requests.

### Dashboard vs logs

I'd use the dashboard to see what's happening right now, like a sudden jump in traffic, errors, or latency, and when it started. I'd use the logs to dig into specific requests after the fact, like which ones failed, on which endpoint, and how long they took, even if nobody was watching Grafana at the time. This run showed why I'd want both: the latency panel could only tell me requests were under half a millisecond, but the logs showed they actually took a few microseconds.
