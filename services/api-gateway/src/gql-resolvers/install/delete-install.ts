import { GraphQLError } from "graphql";
import { DeleteInstallRequest } from "../../build/api/install/v1/messages_pb";
import type {
  Mutation,
  MutationDeleteInstallArgs,
  TResolverFn,
} from "../../types";

export const deleteInstall: TResolverFn<
  MutationDeleteInstallArgs,
  Mutation["deleteInstall"]
> = (_, { id }, { clients }) =>
  new Promise((resolve, reject) => {
    if (clients.install) {
      const request = new DeleteInstallRequest().setId(id);

      clients.install.deleteInstall(request, (err, res) => {
        if (err) {
          reject(new GraphQLError(err?.message));
        } else {
          resolve(res.toObject().deleted);
        }
      });
    } else {
      reject(new GraphQLError("Service isn't available"));
    }
  });
