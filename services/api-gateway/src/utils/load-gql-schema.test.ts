import { readFileSync, readdirSync } from "fs";
import { extname, join } from "path";
import { _loadGQLFile, _parseGQL, loadGQLSchema } from "./load-gql-schema";

jest.mock("fs");
jest.mock("path");

afterAll(() => {
  jest.restoreAllMocks();
});

beforeEach(() => {
  (extname as jest.Mock).mockReturnValue(".graphql");
  (join as jest.Mock).mockImplementation((path, file) =>
    file ? `schema/path/${file}` : "schema/path"
  );
  (readdirSync as jest.Mock).mockReturnValue([
    "test-1.graphql",
    "test-2.graphql",
  ]);
});

test("_loadGQLFile should load a file", () => {
  _loadGQLFile("test.graphql");
  expect(readFileSync).toBeCalledWith("schema/path/test.graphql", "utf-8");
});

test("_parseGQL should turn a loaded gql file into a document node", () => {
  const spec = _parseGQL(["type Query { \n example: String! \n }"]);
  expect(spec).toMatchInlineSnapshot(`
    "
        type Query { 
     example: String! 
     }
      "
  `);
});

test("loadGQLSchema should load all gql files in the schema directory", () => {
  (readFileSync as jest.Mock)
    .mockReturnValueOnce("type Query { \n test: String! \n }")
    .mockReturnValueOnce("extend type Query { \n another: String! \n }");
  const spec = loadGQLSchema();
  expect(spec).toMatchInlineSnapshot(`
    "
        type Query { 
     test: String! 
     },extend type Query { 
     another: String! 
     }
      "
  `);
});
