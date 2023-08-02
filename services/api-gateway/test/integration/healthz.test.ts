import supertest from "supertest";
import { initServer } from "../../src/server";

const request = supertest(initServer());

test("livez endpoint should return health status", async () => {
  const spec = await request.get("/livez");
  expect(spec.status).toBe(200);
  expect(spec.body.status).toBe("ok");
  expect(spec.body).toHaveProperty("version");
});

test("readyz endpoint should return health status", async () => {
  const spec = await request.get("/readyz");
  expect(spec.status).toBe(200);
  expect(spec.body.status).toBe("ok");
  expect(spec.body).toHaveProperty("version");
});

test("health endpoint should return health status", async () => {
  const spec = await request.get("/health");
  expect(spec.status).toBe(200);
  expect(spec.body.status).toBe("ok");
  expect(spec.body).toHaveProperty("version");
});
