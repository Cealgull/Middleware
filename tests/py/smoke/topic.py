from .config import *
from .utils import *
import unittest
import random
import time


class TopicTestCase(unittest.TestCase):
    def setUp(self):
        num = random.randint(0, 10000000)

        self.credential = test_auth_login()
        self.request = get_request_handler(self.credential)

        self.request(
            "/api/categoryGroup/invoke/create",
            {"name": "testGroup" + str(num), "color": "123"},
        )

        time.sleep(0.5)

        self.request(
            "/api/category/invoke/create",
            {
                "categoryGroup": "testGroup" + str(num),
                "name": "testCategory" + str(num),
                "color": "123",
            },
        )

        time.sleep(0.5)

        self.request(
            "/api/tag/invoke/create",
            {"name": "testTag" + str(num), "color": "123"},
        )

        time.sleep(0.5)

        self.num = num

    def test_0001_create_topic(self):
        res = self.request(
            "/api/topic/invoke/create",
            {
                "title": "test",
                "content": "test",
                "images": [],
                "category": "testCategory" + str(self.num),
                "tags": ["testTag" + str(self.num)],
            },
        )
        time.sleep(0.5)

        hash = res["hash"]

        res = self.request(
            "/api/topic/invoke/upvote",
            {"hash": hash},
        )

        time.sleep(0.5)

        res = self.request(
            "/api/topic/query/list",
            {
                "pageOrdinal": 1,
                "pageSize": 10,
                "category": "testCategory" + str(self.num),
                "tags": ["testTag" + str(self.num)],
                "creator": self.credential.wallet,
            },
        )

        self.assertEqual(len(res), 1)

        time.sleep(0.5)

        res = self.request(
            "/api/topic/invoke/update",
            {
                "hash": hash,
                "title": "test2",
                "images": [],
                "content": "test2",
            },
        )
