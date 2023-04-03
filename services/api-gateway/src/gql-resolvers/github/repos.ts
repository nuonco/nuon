import { GetReposRequest } from "@buf/nuon_apis.grpc_node/github/v1/messages_pb";
import { GraphQLError } from "graphql";
import {
  IConnectionResolver,
  TConnection,
  TRepo,
  TResolverFn,
} from "../../types";

interface IReposResolver extends IConnectionResolver {
  githubInstallId: string;
}

export const repos: TResolverFn<IReposResolver, TConnection<TRepo>> = (
  _,
  { githubInstallId },
  { clients }
) =>
  new Promise((resolve, reject) => {
    if (clients.github) {
      const request = new GetReposRequest().setGithubInstallId(githubInstallId);

      clients.github.getRepos(request, (err, res) => {
        if (err) {
          reject(new GraphQLError(err?.message));
        } else {
          const { reposList } = res.toObject();

          resolve({
            edges:
              reposList?.map((github) => ({
                cursor: github?.url,
                node: github,
              })) || [],
            pageInfo: {
              endCursor: null,
              hasNextPage: false,
              hasPreviousPage: false,
              startCursor: null,
            },
            totalCount: reposList?.length || 0,
          });
        }
      });
    } else {
      reject(new GraphQLError("Service isn't available"));
    }
  });
