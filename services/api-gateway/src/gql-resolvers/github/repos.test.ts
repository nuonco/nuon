import { repos } from "./repos";

const mockRepo = {
  defaultBranch: "main",
  fullName: "some-github-repo",
  name: "repo",
  owner: "me",
  url: "github.com/me/repo",
};

const mockRepoServiceClient = {
  getRepos: jest
    .fn()
    .mockImplementationOnce((req, cb) => {
      cb(undefined, {
        toObject: jest.fn().mockReturnValue({
          reposList: [mockRepo],
        }),
      });
    })
    .mockImplementationOnce((req, cb) => {
      cb(undefined, {
        toObject: jest.fn().mockReturnValue({
          reposList: [],
        }),
      });
    })
    .mockImplementationOnce((req, cb) => {
      const err = new Error("some error");
      cb(err);
    }),
};

const mockClients = {
  github: mockRepoServiceClient,
};

test("repos resolver should return Repo connection", async () => {
  const spec = await repos(
    undefined,
    { githubInstallId: "test-id" },
    { clients: mockClients }
  );

  expect(spec).toEqual({
    edges: [
      {
        cursor: "github.com/me/repo",
        node: {
          defaultBranch: "main",
          fullName: "some-github-repo",
          name: "repo",
          owner: "me",
          url: "github.com/me/repo",
        },
      },
    ],
    pageInfo: {
      endCursor: null,
      hasNextPage: false,
      hasPreviousPage: false,
      startCursor: null,
    },
    totalCount: 1,
  });
});

test("repos resolver should return empty connection", async () => {
  const spec = await repos(
    undefined,
    { githubInstallId: "test-id" },
    { clients: mockClients }
  );

  expect(spec).toEqual({
    edges: [],
    pageInfo: {
      endCursor: null,
      hasNextPage: false,
      hasPreviousPage: false,
      startCursor: null,
    },
    totalCount: 0,
  });
});

test("repos resolver should return error on failed query", async () => {
  await expect(
    repos(undefined, { githubInstallId: "test-id" }, { clients: mockClients })
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"some error"`);
});

test("repos resolver should return error if service client doesn't exist", async () => {
  await expect(
    repos(undefined, { githubInstallId: "test-id" }, { clients: {} })
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"Service isn't available"`);
});
