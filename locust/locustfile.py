import random
from locust import HttpUser, TaskSet, task, between, constant
from locust.contrib.fasthttp import FastHttpUser

PRODUCT_IDS = [1100, 1200, 1300, 1400]


class ProductTasks(TaskSet):
    @task(8)
    def get_product(self):
        pid = random.choice(PRODUCT_IDS)
        self.client.get(f"/products/{pid}", name="GET /products/[id]")

    @task(1)
    def update_details(self):
        pid = random.choice(PRODUCT_IDS)
        self.client.post(
            f"/products/{pid}/details",
            json={
                "product_id": pid,
                "sku": f"SKU-{pid}-{random.randint(1, 1000)}",
                "manufacturer": "Load Test Inc",
                "category_id": 1,
                "weight": random.randint(0, 5000),
                "some_other_id": 1,
            },
            name="POST /products/[id]/details",
        )

    @task(1)
    def get_missing_product(self):
        with self.client.get("/products/99999", name="GET /products/[missing]",
                             catch_response=True) as resp:
            if resp.status_code == 404:
                resp.success()


class ProductHttpUser(HttpUser):
    tasks = [ProductTasks]
    wait_time = between(0.5, 2)


class ProductFastHttpUser(FastHttpUser):
    tasks = [ProductTasks]
    wait_time = between(0.5, 2)

class StressFastHttpUser(FastHttpUser):
    tasks = [ProductTasks]
    wait_time = constant(0)

class StressHttpUser(HttpUser):
    tasks = [ProductTasks]
    wait_time = constant(0)