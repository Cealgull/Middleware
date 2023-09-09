from cealgull import *
from locust import HttpUser, TaskSet, task
import random


class UserProfileSection(TaskSet):
    def on_start(self):
        self.credential = cealgull_auth_login(self)
        self.post = register_post(self, self.credential)

    @task(20)
    def view_profile(self):
        self.post(
            "/api/user/query/profile",
            {"wallet": self.credential.wallet},
        )

    @task(5)
    def update_profile(self):
        num = random.randint(0, 1000000)
        self.post(
            "/api/user/invoke/update",
            {
                "username": f"user{num}",
                "signature": f"this is a random signature {num}",
            },
        )

    def on_stop(self):
        self.post("/auth/logout", {})


class TopicSection(TaskSet):
    def on_start(self):
        self.credential = cealgull_auth_login(self)
        self.post = register_post(self, self.credential)
        self.get = register_get(self, self.credential)
        self.tags = [t["name"] for t in self.get("/api/topic/query/tags").json()]
        self.categories = [
            c["name"] for c in self.get("/api/topic/query/categories").json()
        ]

    @task(5)
    def create_topic(self):
        self.post(
            "/api/topic/invoke/create",
            {
                "title": "this is a random title",
                "content": "this is test content",
                "tags": random.sample(self.tags, random.randint(0, len(self.tags))),
                "category": random.choice(self.categories),
            },
        )

    @task(20)
    def list_topic(self):
        self.post(
            "/api/topic/query/list",
            {
                "pageOrdinal": 1,
                "pageSize": 10,
                "category": random.choice(self.categories),
                "tags": random.sample(self.tags, random.randint(0, len(self.tags))),
            },
        )

    @task(2)
    def upvote_topic(self):
        topics = self.post(
            "/api/topic/query/list",
            {
                "pageOrdinal": 1,
                "pageSize": 10,
                "category": random.choice(self.categories),
                "tags": random.sample(self.tags, random.randint(0, len(self.tags))),
            },
        ).json()

        if len(topics) > 0:
            self.post(
                "/api/topic/invoke/upvote",
                {
                    "hash": random.choice(topics)["hash"],
                },
            )

    @task(2)
    def downvote_topic(self):
        topics = self.post(
            "/api/topic/query/list",
            {
                "pageOrdinal": 1,
                "pageSize": 10,
                "category": random.choice(self.categories),
                "tags": random.sample(self.tags, random.randint(0, len(self.tags))),
            },
        ).json()

        if len(topics) > 0:
            self.post(
                "/api/topic/invoke/downvote",
                {
                    "hash": random.choice(topics)["hash"],
                },
            )


class LoginUser(HttpUser):
    tasks = {TopicSection: 5}
