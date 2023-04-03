import { readFileSync, readdirSync } from "fs";
import { extname, join } from "path";

const SCHEMA_PATH = join(__dirname, "..", "gql-schema");

export function _loadGQLFile(fileName) {
  return readFileSync(join(SCHEMA_PATH, fileName), "utf-8");
}

export function _parseGQL(...files) {
  return `
    ${files}
  `;
}

export function loadGQLSchema(schemaPath = SCHEMA_PATH) {
  const schemaFiles = readdirSync(schemaPath);
  return _parseGQL(
    schemaFiles.filter((file) => extname(file) === ".graphql").map(_loadGQLFile)
  );
}
