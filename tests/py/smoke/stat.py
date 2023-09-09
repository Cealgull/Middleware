from smoke.post import *
from smoke.topic import *
from .utils import *
import unittest
import time


class StatisticTestCase(unittest.TestCase):
    def setUp(self):
        self.credentials = test_auth_login()
        self.request = register_POST_handler(self.credentials)
        self.num = TopicTestCase.create_plugs(self.request)
        time.sleep(0.5) 
        self.topic_hash = TopicTestCase.create_topic(self.request, self.num)
        time.sleep(0.5)
        self.post_hash = PostTestCase.create_post(self.request, self.topic_hash)
        time.sleep(0.5)
        self.request("/api/post/invoke/upvote", {"hash": self.post_hash})
        time.sleep(0.5)
        self.request("/api/topic/invoke/upvote", {"hash": self.topic_hash})
        time.sleep(0.5)

    def test_0001_query_statistics(self):
        res = self.request(
            "/api/user/query/statistics", {"wallet": self.credentials.wallet}
        )
        print(res)
