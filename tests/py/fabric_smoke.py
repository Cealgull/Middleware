from cryptography.hazmat.primitives.asymmetric import ed25519
import requests
import base64


def main():
    priv = ed25519.Ed25519PrivateKey.generate()
    pub = priv.public_key()
    pubb64 = base64.b64encode(pub.public_bytes_raw()).decode()

    cert = requests.post(
        "http://localhost:7999/cert/sign",
        headers={"content-type": "application/json", "signature": "HACK"},
        json={"pub": pubb64},
    ).json()

    sig = base64.b64encode(priv.sign(cert["cert"].encode())).decode()

    res = requests.post(
        "http://localhost:8080/api/user/create",
        headers={"content-type": "application/json", "signature": sig},
        json=cert,
    )

    print(res.text)


if __name__ == "__main__":
    main()
