# Cealgull Middleware

Cealgull Middleware is a middleware which receive requests from Cealgull app and handle them with Hyperledger Firefly api and IPFS storage system.

You need a `config.yaml` file in the root directory of the project with the following structure:

```yaml
port: <port>
ipfs:
  url: <host>:<port>
firefly:
  url:
    - http://<host>:<port>
    - http://<host>:<port>
ca:
  url: http://<host>:<port>
```
