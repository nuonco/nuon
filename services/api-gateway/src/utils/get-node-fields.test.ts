import type { Node } from "../types";
import { getNodeFields } from "./get-node-fields";

test("getNodeFields should return node object with formated dates", () => {
  const mockDateTimeObject = {
    day: 31,
    hours: 8,
    minutes: 15,
    month: 12,
    seconds: 30,
    year: 1999,
  };
  const spec = getNodeFields<Node>({
    createdAt: mockDateTimeObject,
    id: "test-id",
    name: "test-node",
    updatedAt: mockDateTimeObject,
  });

  expect(spec).toMatchInlineSnapshot(`
    {
      "createdAt": "1999-12-31T08:15:30.000",
      "id": "test-id",
      "name": "test-node",
      "updatedAt": "1999-12-31T08:15:30.000",
    }
  `);
});
