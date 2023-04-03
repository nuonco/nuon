import "dotenv/config";
import { cleanEnv, json, port, str } from "envalid";

export const env = cleanEnv(process.env, {
  AUTH_AUDIENCE: str({ default: "localhost:3000" }),
  AUTH_ISSUER: str({ default: "https://nuon.us.auth0.com" }),
  HOST: str({ default: "localhost" }),
  LOG_LEVEL: str({
    choices: ["fatal", "error", "warn", "info", "debug", "trace", "silent"],
    default: "info",
  }),
  NODE_ENV: str({
    choices: ["development", "production", "test"],
    default: "production",
  }),
  PORT: port({ default: 3000 }),
  SERVICES: json({
    default: [{ name: "noop", url: "http://localhost:8080" }],
    desc: "List of gRPC services",
  }),
});
