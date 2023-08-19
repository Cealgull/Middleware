import { check } from "k6";
import { login, post } from "./utils";
import { Options } from "k6/options";

export const options: Options = {
  hosts: { "gateway.cealgull.xyz": "localhost:8080" },
  thresholds: {
    http_req_duration: ["p(95)<500"],
  },
  scenarios: {
    basic_scenerio: {
      executor: "constant-arrival-rate",
      preAllocatedVUs: 50,
      duration: "15s",
      timeUnit: "1s",
      rate: 30,
    },
    // failed_scenerio: {
    //   executor: "constant-arrival-rate",
    //   preAllocatedVUs: 20,
    //   duration: "30s",
    //   exec: "login_without_cert",
    //   rate: 300,
    // },
  },
};

export function login_without_cert() {
  const resp = post(
    "http://localhost:8080/auth/login",
    {},
    {
      "content-type": "application/json",
    }
  );
  check(resp, {
    "is status 400": (r) => r.status == 400,
  });
}

export default function () {
  const resp = login();
  check(resp, {
    "is status 200": (r) => r.status == 200,
  });
}
