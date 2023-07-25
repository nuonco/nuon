import { credentials } from "@grpc/grpc-js";
import { GetTokenRequest } from "../build/orgs-api/orgs/v1/token_pb";
import { WaypointClient } from "../build/waypoint/server/waypoint/main_grpc_pb";

type TOrgWaypointInfo = {
  address: string;
  token: string;
};

export function getOrgWaypointToken(
  orgId: string,
  orgsAPIClient: any
): Promise<TOrgWaypointInfo> {
  return new Promise((resolve, reject) => {
    const request = new GetTokenRequest().setOrgId(orgId);

    orgsAPIClient.getToken(request, (err, res) => {
      if (err) {
        reject(err);
      } else {
        resolve(res?.toObject());
      }
    });
  });
}

export function initOrgWaypointClient(url: string): any {
  return new WaypointClient(url, credentials.createSsl());
}
