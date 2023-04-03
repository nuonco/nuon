import { me } from "./me";

test("me resolver should return test user from gql ctx", () => {
  const spec = me(undefined, undefined, { user: { id: "test" } });

  expect(spec).toEqual({ id: "test" });
});
