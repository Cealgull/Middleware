from .config import *
from .utils import *
import unittest
import random
import time


class TopicTestCase(unittest.TestCase):
    def setUp(self):
        num = random.randint(0, 10000000)

        self.credential = test_auth_login()

        res = request_with_credential(
            self.credential,
            "/api/categoryGroup/invoke/create",
            {"name": "testGroup" + str(num), "color": 123},
        )

        print(res)

        request_with_credential(
            self.credential,
            "/api/category/invoke/create",
            {
                "categoryGroup": "testGroup" + str(num),
                "name": "testCategory" + str(num),
                "color": 123,
            },
        )

        print(res)

        request_with_credential(
            self.credential,
            "/api/tag/invoke/create",
            {"name": "testTag" + str(num), "color": 123},
        )

        self.num = num

        print(res)

    def test_0001_create_topic(self):
        res = request_with_credential(
            self.credential,
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

        print(res)
