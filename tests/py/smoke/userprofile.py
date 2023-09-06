from .config import *
from .utils import *

import requests
import unittest
import time


class UserProfileSmokeTest(unittest.TestCase):
    def setUp(self):
        self.credential = test_auth_login()

    def test_0001_update_user(self):
        res = request_with_credential(
            self.credential,
            "/api/user/invoke/update",
            {
                "username": "alice_exp",
                "avatar": "0xsaadfwadf",
                "wallet": self.credential.wallet,
                "signature": "Genshin Impact is a good game",
            },
        )

        time.sleep(0.1)

        res = request_with_credential(
            self.credential,
            "/api/user/query/profile",
            {"wallet": self.credential.wallet},
        )

        self.assertEqual(res["username"], "alice_exp")
        self.assertEqual(res["wallet"], self.credential.wallet)
        self.assertEqual(res["avatar"], "0xsaadfwadf")
        self.assertEqual(res["signature"], "Genshin Impact is a good game")
