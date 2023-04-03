import { GetOrgsByMemberRequest } from "@buf/nuon_apis.grpc_node/org/v1/messages_pb";
import { GraphQLError } from "graphql";
import {
  IConnectionResolver,
  TConnection,
  TOrg,
  TResolverFn,
} from "../../types";
import { getNodeFields } from "../../utils";

interface IOrgsResolver extends IConnectionResolver {
  memberId: string;
}

export const orgs: TResolverFn<IOrgsResolver, TConnection<TOrg>> = (
  _,
  { memberId },
  { clients }
) =>
  new Promise((resolve, reject) => {
    if (clients.org) {
      const request = new GetOrgsByMemberRequest().setMemberId(memberId);

      clients.org.getOrgsByMember(request, (err, res) => {
        if (err) {
          reject(new GraphQLError(err?.message));
        } else {
          const { orgsList } = res.toObject();

          resolve({
            edges:
              orgsList?.map((org) => ({
                cursor: org?.id,
                node: getNodeFields(org),
              })) || [],
            pageInfo: {
              endCursor: null,
              hasNextPage: false,
              hasPreviousPage: false,
              startCursor: null,
            },
            totalCount: orgsList.length || 0,
          });
        }
      });
    } else {
      reject(new GraphQLError("Service isn't available"));
    }
  });
