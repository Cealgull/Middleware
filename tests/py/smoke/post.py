from smoke.utils import *
import unittest
import random


class PostTestCase(unittest.TestCase):
    def setUp(self):
        self.credentials = test_auth_login()
        self.request = get_request_handler(self.credentials)

        num = random.randint(0, 10000000)
        self.num = num

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

        res = self.request(
            "/api/topic/invoke/create",
            {
                "title": "test",
                "content": "test",
                "images": [],
                "category": "testCategory" + str(num),
                "tags": ["testTag" + str(num)],
            },
        )

        self.topic_hash = res["hash"]

    def test_0001_create_post(self):
        time.sleep(0.5)
        res = self.request(
            "/api/post/invoke/create",
            {
                "content": "this is a testing post",
                "images": [],
                "belongTo": self.topic_hash,
                "replyTo": "",
            },
        )

        return res

    def test_0002_get_query_list(self):
        self.test_0001_create_post()
        time.sleep(0.5)
        res = self.request(
            "/api/post/query/list",
            {
                "pageOrdinal": 1,
                "pageSize": 10,
                "belongTo": self.topic_hash,
                "creator": self.credential.wallet,
            },
        )
        print(res)

    def test_003_update_post(self):
        res = self.test_0001_create_post()
        time.sleep(0.5)
        res = self.request(
            "/api/post/invoke/update",
            {
                "hash": res["hash"],
                "content": "this is another testing post",
                "images": [],
            },
        )

        print(res)
