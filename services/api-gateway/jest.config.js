process.env.TZ = 'UTC';

module.exports = {
  roots: ["<rootDir>"],
  transform: {
    "^.+\\.(ts|tsx)$": ["ts-jest", {
      isolatedModules: true,
    }],
  },
  testPathIgnorePatterns: ["/node_modules/"],
  setupFilesAfterEnv: [],
  testEnvironment: "node",
  coveragePathIgnorePatterns: ["/node_modules/"],
};
