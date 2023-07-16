import { GraphQLError } from "graphql";
import { GetOrgsByMemberRequest } from "../../build/api/org/v1/messages_pb";
import type { Org, Query, QueryOrgsArgs, TResolverFn } from "../../types";
import { getNodeFields } from "../../utils";

export const orgs: TResolverFn<QueryOrgsArgs, Query["orgs"]> = (
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
                node: getNodeFields<Org>(org),
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
