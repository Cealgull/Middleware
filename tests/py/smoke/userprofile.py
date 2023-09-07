from .config import *
from .utils import *

import unittest
import time


class UserProfileSmokeTest(unittest.TestCase):
    def setUp(self):
        self.credential = test_auth_login()
        self.request = get_request_handler(self.credential)

    def test_0001_update_user(self):
        time.sleep(0.5)

        res = self.request(
            "/api/user/invoke/update",
            {
                "username": "alice_exp",
                "avatar": "0xsaadfwadf",
                "wallet": self.credential.wallet,
                "signature": "Genshin Impact is a good game",
            },
        )

        time.sleep(0.5)

        res = self.request(
            "/api/user/query/profile",
            {"wallet": self.credential.wallet},
        )

        self.assertEqual(res["username"], "alice_exp")
        self.assertEqual(res["wallet"], self.credential.wallet)
        self.assertEqual(res["avatar"], "0xsaadfwadf")
        self.assertEqual(res["signature"], "Genshin Impact is a good game")

    def test_0002_query_statitics(self):
        res = self.request(
            "/api/user/query/statistics", {"wallet": self.credential.wallet}
        )

        print(res)


if __name__ == "__main__":
    unittest.main()
