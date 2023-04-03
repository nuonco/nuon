import { echo } from "./echo";

test("echo resolver should return provided string", () => {
  const spec = echo(undefined, { word: "test" });

  expect(spec).toBe("test");
});
