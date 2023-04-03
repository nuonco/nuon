import type { CodegenConfig } from "@graphql-codegen/cli";

const config: CodegenConfig = {
  overwrite: true,
  schema: "./src/gql-schema/*.graphql",
  generates: {
    "src/types/gql.ts": {
      plugins: [
        {
          add: {
            content: "/* eslint-disable */",
          },
        },
        "typescript",
      ],
    },
  },
};

export default config;
