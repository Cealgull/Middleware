from cryptography.hazmat.primitives.asymmetric import ed25519
import base64
import requests
import random


def get_cert():
    private_key = ed25519.Ed25519PrivateKey.generate()
    public_key = private_key.public_key()
    pubb64 = base64.b64encode(public_key.public_bytes_raw()).decode()
    res = requests.post(
        "http://localhost:7999/cert/sign",
        headers={"content-type": "application/json", "signature": "HACK"},
        json={"pub": pubb64},
    ).json()
    print(pubb64)
    return private_key, res["cert"]


def register(private_key: ed25519.Ed25519PrivateKey, cert: str):
    signature = base64.b64encode(private_key.sign(cert.encode())).decode()
    result = requests.post(
        "http://localhost:8080/auth/login",
        headers={"content-type": "application/json", "signature": signature},
        json={"cert": cert},
    )
    print(result.json())
    return result.cookies


def create_topic(cookies):
    result = requests.post(
        "http://localhost:8080/api/create/topic",
        headers={"content-type": "application/json"},
        cookies=cookies,
        json={
            "title": "testabc",
            "content": "test" + str(random.randint(0, 100000)),
            "category": "genshin-impact",
            "tags": [],
            "images": [],
        },
    )

    print(result.json())

    result = requests.get(
        "http://localhost:8080/api/list/topics",
        headers={"content-type": "application/json"},
        cookies=cookies,
    )

    print(result.json())

    return result.json()[0]["id"]


def create_post(cookies, postId):
    result = requests.post(
        "http://localhost:8080/api/create/post",
        headers={"content-type": "application/json"},
        cookies=cookies,
        json={
            "content": "test" + str(random.randint(0, 100000)),
            "images": [],
            "replyTo": "",
            "belongTo": postId,
        },
    )

    print(result.json())


if __name__ == "__main__":
    priv, cert = get_cert()
    cookies = register(priv, cert)
    # postId = create_topic(cookies)
    # create_post(cookies, postId)
