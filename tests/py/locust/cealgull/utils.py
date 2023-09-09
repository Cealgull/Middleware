from .config import *

from cryptography.hazmat.primitives.asymmetric import ed25519
from dataclasses import dataclass
from typing import Any
import base64


@dataclass
class Credential:
    wallet: str
    cert: str
    cookies: Any


def cealgull_auth_login(user):
    priv = ed25519.Ed25519PrivateKey.generate()
    pub = priv.public_key()
    pubb64 = base64.b64encode(pub.public_bytes_raw()).decode()

    cert = user.client.post(
        CEALGULL_CA_HOST + "/cert/sign",
        headers={"signature": "HACK"},
        json={"pub": pubb64},
    ).json()["cert"]
    sig = base64.b64encode(priv.sign(cert.encode())).decode()

    res = user.client.post(
        CEALGULL_MIDDLEWARE_HOST + "/auth/login",
        headers={"signature": sig},
        json={"cert": cert},
    )

    wallet = res.json()["wallet"]
    cookies = res.cookies

    return Credential(wallet=wallet, cookies=cookies, cert=cert)


def register_post(user, credential: Credential):
    def r(url: str, data: dict):
        return user.client.post(
            CEALGULL_MIDDLEWARE_HOST + url, json=data, cookies=credential.cookies
        )
    return r

def register_get(user, credential: Credential):
    def r(url: str):
        return user.client.get(
            CEALGULL_MIDDLEWARE_HOST + url, cookies=credential.cookies
        )
    return r
