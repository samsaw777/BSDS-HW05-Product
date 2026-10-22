import random

from locust import FastHttpUser, task, constant

SEARCH_TERMS = [
     "electro", "book", "home", "sports", "toys",   
    "alpha", "beta", "gamma", "nova", "product",
]

class SearchUser(FastHttpUser):
    wait_time = constant(0)
 
    @task
    def search(self):
        q = random.choice(SEARCH_TERMS)
        self.client.get(f"/products/search?q={q}", name="/products/search")
 