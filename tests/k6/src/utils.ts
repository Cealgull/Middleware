import { Buffer } from "buffer";
import crypto from "k6/crypto";
import nacl from "tweetnacl";
import http from "k6/http";
import { JSONObject } from "k6";
import BN from "bn.js";

const verifyHost = "http://localhost:7999";
const middlewareHost = "http://localhost:8080";

export interface Headers {
  [key: string]: string;
}

export interface Payload {
  [key: string]: string | string[];
}

export const login = (): ReturnType<typeof http.post> => {
  const keypair = nacl.sign.keyPair.fromSeed(
    new Uint8Array(crypto.randomBytes(32))
  );

  const pub = Buffer.from(keypair.publicKey).toString("base64");

  const headers = {
    "content-type": "application/json",
    signature: "HACK",
  };

  const payload = JSON.stringify({
    pub: pub,
  });

  const certResponse = http
    .post(`${verifyHost}/cert/sign`, payload, {
      headers,
    })
    .json();

  const cert = (certResponse as JSONObject).cert as string;
  const signature = nacl.sign.detached(Buffer.from(cert), keypair.secretKey);
  headers.signature = Buffer.from(signature).toString("base64");

  const resp = http.post(
    `${middlewareHost}/auth/login`,
    JSON.stringify(certResponse),
    {
      headers,
    }
  );
  return resp;
};

export const random = (): BN => {
  return new BN(Buffer.from(crypto.randomBytes(32)));
};

export const get = (
  endpoint: string,
  headers: Headers
): ReturnType<typeof http.get> => {
  return http.get(`${middlewareHost}/${endpoint}`, { headers });
};

export const post = (
  endpoint: string,
  payload?: Payload,
  headers?: Headers
): ReturnType<typeof http.post> => {
  return http.post(`${middlewareHost}/${endpoint}`, JSON.stringify(payload), {
    headers,
  });
};
