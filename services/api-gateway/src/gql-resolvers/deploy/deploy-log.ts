import { Metadata } from "@grpc/grpc-js";
import { GraphQLError } from "graphql";
import { GetJobStreamRequest } from "../../build/waypoint/server/waypoint/main_pb";
import type { Query, QueryDeployLogArgs, TResolverFn } from "../../types";
import { getOrgWaypointToken, initOrgWaypointClient } from "../utils";

export const deployLog: TResolverFn<QueryDeployLogArgs, Query["deployLog"]> = (
  _,
  { deployId, orgId },
  { clients }
) =>
  new Promise((resolve, reject) => {
    if (clients.orgStatus) {
      getOrgWaypointToken(orgId, clients.orgStatus)
        .then(({ address, token }) => {
          // init org waypoint client
          const wpClient = initOrgWaypointClient(address);

          // init metadata with access token
          const metadata = new Metadata();
          metadata.add("authorization", token);
          metadata.add("client-api-protocol", "1,1");

          // init job stream request
          const request = new GetJobStreamRequest().setJobId(
            `sync-${deployId}`
          );
          const jobStreamCall = wpClient.getJobStream(request, metadata);

          // handle job stream request on error, data & end
          jobStreamCall.on("error", (err) => {
            reject(new GraphQLError(err?.message));
          });

          let logs = [];
          jobStreamCall.on("data", (res) => {
            const data = res.toObject();
            if (data?.terminal?.eventsList) {
              logs = data?.terminal?.eventsList;
            }
          });

          jobStreamCall.on("end", () => {
            resolve({
              deployId,
              logs,
            });
          });
        })
        .catch((err) => {
          reject(new GraphQLError(err?.message));
        });
    } else {
      reject(new GraphQLError("Service isn't available"));
    }
  });
