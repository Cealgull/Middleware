import http from "k6/http";
import { Payload, post, login, random } from "./utils";
import { check } from "k6";

const createTopic = (): ReturnType<typeof http.post> => {
  const topic: Payload = {
    title: "hello world",
    images: [],
    tags: ["abc"],
    category: "testing",
    content: `hello world + ${random().toString("hex")}`,
  };

  return post("api/topic/create", topic, {
    "content-type": "application/json",
  });
};

export default function () {
  const account = login();
  check(account, {
    "resp is 200": (a) => a.status == 200,
  });
  const topic = createTopic();
  check(topic, {
    "resp is 200": (a) => a.status == 200,
  });
}
