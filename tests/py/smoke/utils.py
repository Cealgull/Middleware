from collections.abc import Callable
from .config import *

from cryptography.hazmat.primitives.asymmetric import ed25519
from requests.sessions import RequestsCookieJar
from dataclasses import dataclass
import requests
import base64
import time


@dataclass
class Credential:
    wallet: str
    cert: str
    cookies: RequestsCookieJar


def test_auth_login() -> Credential:
    priv = ed25519.Ed25519PrivateKey.generate()
    pub = priv.public_key()
    pubb64 = base64.b64encode(pub.public_bytes_raw()).decode()

    cert = requests.post(
        CEALGULL_CA_HOST + "/cert/sign",
        headers={"signature": "HACK"},
        json={"pub": pubb64},
    ).json()["cert"]
    sig = base64.b64encode(priv.sign(cert.encode())).decode()

    res = requests.post(
        CEALGULL_MIDDLEWARE_HOST + "/auth/login",
        headers={"signature": sig},
        json={"cert": cert},
    )

    wallet = res.json()["wallet"]
    cookies = res.cookies

    return Credential(wallet=wallet, cookies=cookies, cert=cert)


def register_POST_handler(credential: Credential) -> Callable[[str, dict], dict]:
    def f(endpoint: str, payload: dict):
        return requests.post(
            CEALGULL_MIDDLEWARE_HOST + endpoint,
            cookies=credential.cookies,
            json=payload,
        ).json()

    return f


def register_GET_handler(credential: Credential) -> Callable[[str, dict], dict]:
    def f(endpoint: str):
        return requests.get(
            CEALGULL_MIDDLEWARE_HOST + endpoint, cookies=credential.cookies
        ).json()
    return f
