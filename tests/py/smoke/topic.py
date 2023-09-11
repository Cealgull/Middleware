from .config import *
from .utils import *
import unittest
import random
import time

lipsum_text = "There are many variations of passages of Lorem Ipsum available, but the majority have suffered alteration in some form, by injected humour, or randomised words which don't look even slightly believable. If you are going to use a passage of Lorem Ipsum, you need to be sure there isn't anything embarrassing hidden in the middle of text. All the Lorem Ipsum generators on the Internet tend to repeat predefined chunks as necessary, making this the first true generator on the Internet. It uses a dictionary of over 200 Latin words, combined with a handful of model sentence structures, to generate Lorem Ipsum which looks reasonable. The generated Lorem Ipsum is therefore always free from repetition, injected humour, or non-characteristic words etc."

class TopicTestCase(unittest.TestCase):
    @classmethod
    def create_plugs(cls, request: Callable[[str, dict], dict]) -> int:
        num = random.randint(0, 10000000)

        request(
            "/api/categoryGroup/invoke/create",
            {"name": "testGroup" + str(num), "color": "123"},
        )

        time.sleep(0.5)

        request(
            "/api/category/invoke/create",
            {
                "categoryGroup": "testGroup" + str(num),
                "name": "testCategory" + str(num),
                "color": "123",
            },
        )

        time.sleep(0.5)

        request(
            "/api/tag/invoke/create",
            {"name": "testTag" + str(num)},
        )

        time.sleep(0.5)

        return num

    @classmethod
    def create_topic(cls, request: Callable[[str, dict], dict], num: int) -> str:
        rand_int = random.randint(0, 1000000)
        return request(
            "/api/topic/invoke/create",
            {
                "title": "test" + str(rand_int),
                "content": lipsum_text[rand_int % 100 : rand_int % 100 + 50],
                "images": [],
                "category": "testCategory" + str(num),
                "tags": ["testTag" + str(num)],
            },
        )["hash"]

    @classmethod
    def query_topics(cls, request: Callable[[str, dict], list | dict], num: int):
        return request(
            "/api/topic/query/list",
            {
                "pageOrdinal": 1,
                "pageSize": 10,
                "category": "testCategory" + str(num),
                "tags": ["testTag" + str(num)],
            },
        )

    def setUp(self):
        self.credential = test_auth_login()
        self.request = register_POST_handler(self.credential)
        self.num = self.create_plugs(self.request)

    def test_0001_create_topic(self):
        self.create_topic(self.request, self.num)

    def test_0002_query_topics(self):
        for _ in range(20):
            self.create_topic(self.request, self.num)
        res = self.query_topics(self.request, self.num)
        topic = res[0]
        self.assertEqual(len(res), 10)
        self.assertEqual(
            topic["categoryAssigned"],
            {"name": "testCategory" + str(self.num), "color": "123"},
        )
        self.assertEqual(topic["creator"]["wallet"], self.credential.wallet)
        self.assertEqual(topic["title"], "test")
        self.assertEqual(topic["content"], "test")

    def test_0003_invoke_upvotes(self):
        hash = self.create_topic(self.request, self.num)
        time.sleep(0.5)

        self.request("/api/topic/invoke/upvote", {"hash": hash})
        time.sleep(0.5)

        res = self.query_topics(self.request, self.num)[0]
        self.assertIn(self.credential.wallet, res["upvotes"])

        self.request("/api/topic/invoke/upvote", {"hash": hash})
        time.sleep(0.5)

        res = self.query_topics(self.request, self.num)[0]
        self.assertNotIn(self.credential.wallet, res["upvotes"])

    def test_0004_invoke_downvotes(self):
        hash = self.create_topic(self.request, self.num)
        time.sleep(0.5)

        self.request("/api/topic/invoke/downvote", {"hash": hash})
        time.sleep(0.5)

        res = self.query_topics(self.request, self.num)[0]
        self.assertIn(self.credential.wallet, res["downvotes"])

        self.request("/api/topic/invoke/downvote", {"hash": hash})
        time.sleep(0.5)

        res = self.query_topics(self.request, self.num)[0]
        self.assertNotIn(self.credential.wallet, res["downvotes"])

    def test_0005_invoke_upvote_downvote(self):
        hash = self.create_topic(self.request, self.num)
        time.sleep(0.5)

        self.request("/api/topic/invoke/upvote", {"hash": hash})
        time.sleep(0.5)

        self.request("/api/topic/invoke/downvote", {"hash": hash})
        time.sleep(0.5)
        res = self.query_topics(self.request, self.num)[0]
        self.assertNotIn(self.credential.wallet, res["upvotes"])
        self.assertIn(self.credential.wallet, res["downvotes"])

        self.request("/api/topic/invoke/upvote", {"hash": hash})
        time.sleep(0.5)
        res = self.query_topics(self.request, self.num)[0]
        self.assertIn(self.credential.wallet, res["upvotes"])
        self.assertNotIn(self.credential.wallet, res["downvotes"])

    def test_0006_invoke_update_topics(self):
        num_2 = self.create_plugs(self.request)
        hash = self.create_topic(self.request, self.num)

        time.sleep(0.5)

        self.request(
            "/api/topic/invoke/update",
            {
                "hash": hash,
                "title": "test2",
                "content": "test2",
                "category": "testCategory" + str(num_2),
                "tags": ["testTag" + str(num_2)],
            },
        )

        time.sleep(0.5)

    def test_0007_invoke_delete_topics(self):
        hash = self.create_topic(self.request, self.num)
        time.sleep(0.5)
        self.request(
            "/api/topic/invoke/delete",
            {"hash": hash, "creator": self.credential.wallet},
        )
        time.sleep(0.5)
        self.assertEqual(0, len(self.query_topics(self.request, self.num)))

    def test_0008_query_categories(self):
        self.request = register_GET_handler(self.credential)
        self.assertGreater(len(self.request("/api/topic/query/categories")),0)

    def test_0009_query_tags(self):
        self.request = register_GET_handler(self.credential)
        self.assertGreater(len(self.request("/api/topic/query/tags")), 0)
