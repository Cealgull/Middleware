from smoke.utils import *
from smoke.topic import *
import unittest
import time


class PostTestCase(unittest.TestCase):
    def setUp(self):
        self.credentials = test_auth_login()
        self.request = get_request_handler(self.credentials)
        self.num = TopicTestCase.create_plugs(self.request)
        time.sleep(0.5)
        self.topic_hash = TopicTestCase.create_topic(self.request, self.num)
        time.sleep(0.5)

    @classmethod
    def create_post(
        cls, request: Callable[[str, dict], dict], topic_hash: str, post_hash: str = ""
    ):
        return request(
            "/api/post/invoke/create",
            {
                "content": "this is a testing post",
                "images": [],
                "belongTo": topic_hash,
                "replyTo": post_hash,
            },
        )["hash"]

    @classmethod
    def query_posts(
        cls, request: Callable[[str, dict], list | dict], topic_hash: str, wallet: str
    ):
        return request(
            "/api/post/query/list",
            {
                "pageOrdinal": 1,
                "pageSize": 10,
                "belongTo": topic_hash,
                "creator": wallet,
            },
        )

    def test_0001_create_post(self):
        self.create_post(self.request, self.topic_hash)

    def test_0002_query_posts(self):
        for _ in range(20):
            self.create_post(self.request, self.topic_hash)
        time.sleep(0.5)
        res = self.query_posts(self.request, self.topic_hash, self.credentials.wallet)
        self.assertEqual(len(res), 10)
        self.assertEqual(res[0]["creator"]["wallet"], self.credentials.wallet)
        self.assertEqual(res[0]["content"], "this is a testing post")

    def test_0003_reply_to_posts(self):
        hash = self.create_post(self.request, self.topic_hash)
        time.sleep(0.5)
        self.create_post(self.request, self.topic_hash, hash)
        time.sleep(0.5)
        res = self.query_posts(self.request, self.topic_hash, self.credentials.wallet)[
            1
        ]
        self.assertEqual(res["replyTo"]["hash"], hash)

    def test_0004_delete_posts(self):
        hash = self.create_post(self.request, self.topic_hash)
        time.sleep(0.5)
        self.request(
            "/api/post/invoke/delete",
            {"hash": hash, "creator": self.credentials.wallet},
        )
        time.sleep(0.5)
        res = self.query_posts(self.request, self.topic_hash, self.credentials.wallet)
        self.assertEqual(len(res), 0)

    def test_0005_upvote_posts(self):
        hash = self.create_post(self.request, self.topic_hash)
        time.sleep(0.5)
        self.request("/api/post/invoke/upvote", {"hash": hash})
        time.sleep(0.5)
        res = self.query_posts(self.request, self.topic_hash, self.credentials.wallet)[
            0
        ]
        self.assertIn(self.credentials.wallet, res["upvotes"])
        self.request("/api/post/invoke/upvote", {"hash": hash})
        time.sleep(0.5)
        res = self.query_posts(self.request, self.topic_hash, self.credentials.wallet)[
            0
        ]
        self.assertNotIn(self.credentials.wallet, res["upvotes"])

    def test_0005_downvote_posts(self):
        hash = self.create_post(self.request, self.topic_hash)
        time.sleep(0.5)
        self.request("/api/post/invoke/downvote", {"hash": hash})
        time.sleep(0.5)
        res = self.query_posts(self.request, self.topic_hash, self.credentials.wallet)[
            0
        ]
        self.assertIn(self.credentials.wallet, res["downvotes"])
        self.request("/api/post/invoke/downvote", {"hash": hash})
        time.sleep(0.5)
        res = self.query_posts(self.request, self.topic_hash, self.credentials.wallet)[
            0
        ]
        self.assertNotIn(self.credentials.wallet, res["downvotes"])

    def test_0006_upvote_downvote_posts(self):
        hash = self.create_post(self.request, self.topic_hash)

        time.sleep(0.5)

        self.request("/api/post/invoke/upvote", {"hash": hash})

        time.sleep(0.5)

        res = self.query_posts(self.request, self.topic_hash, self.credentials.wallet)[
            0
        ]
        self.assertIn(self.credentials.wallet, res["upvotes"])
        self.request("/api/post/invoke/downvote", {"hash": hash})

        time.sleep(0.5)

        res = self.query_posts(self.request, self.topic_hash, self.credentials.wallet)[
            0
        ]
        self.assertNotIn(self.credentials.wallet, res["upvotes"])
        self.assertIn(self.credentials.wallet, res["downvotes"])
        self.request("/api/post/invoke/upvote", {"hash": hash})

        time.sleep(0.5)

        res = self.query_posts(self.request, self.topic_hash, self.credentials.wallet)[
            0
        ]
        self.assertNotIn(self.credentials.wallet, res["downvotes"])
        self.assertIn(self.credentials.wallet, res["upvotes"])

    def test_0007_update_posts(self):
        hash = self.create_post(self.request, self.topic_hash)
        time.sleep(0.5)
        self.request("/api/post/invoke/update", {"content": "hello world", "hash": hash})
        time.sleep(0.5)
        res = self.query_posts(self.request, self.topic_hash, self.credentials.wallet)[0]
        self.assertEqual(res["content"], "hello world")
