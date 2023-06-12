import { AppsServiceClient } from "@buf/nuon_apis.grpc_node/app/v1/service_grpc_pb";
import { BuildsServiceClient } from "@buf/nuon_apis.grpc_node/build/v1/service_grpc_pb";
import { ComponentsServiceClient } from "@buf/nuon_apis.grpc_node/component/v1/service_grpc_pb";
import { DeployServiceClient } from "@buf/nuon_apis.grpc_node/deploy/v1/service_grpc_pb";
import { DeploymentsServiceClient } from "@buf/nuon_apis.grpc_node/deployment/v1/service_grpc_pb";
import { GithubServiceClient } from "@buf/nuon_apis.grpc_node/github/v1/service_grpc_pb";
import { InstallsServiceClient } from "@buf/nuon_apis.grpc_node/install/v1/service_grpc_pb";
import { OrgsServiceClient } from "@buf/nuon_apis.grpc_node/org/v1/service_grpc_pb";
import { BuildsServiceClient as BuildsStatusServiceClient } from "@buf/nuon_orgs-api.grpc_node/builds/v1/service_grpc_pb";
import { DeploymentsServiceClient as DeploymentsStatusServiceClient } from "@buf/nuon_orgs-api.grpc_node/deployments/v1/service_grpc_pb";
import { InstallsServiceClient as InstallsStatusServiceClient } from "@buf/nuon_orgs-api.grpc_node/installs/v1/service_grpc_pb";
import { InstancesServiceClient as InstancesStatusServiceClient } from "@buf/nuon_orgs-api.grpc_node/instances/v1/service_grpc_pb";
import { OrgsServiceClient as OrgsStatusServiceClient } from "@buf/nuon_orgs-api.grpc_node/orgs/v1/service_grpc_pb";
import { StatusServiceClient } from "@buf/nuon_shared.grpc_node/status/v1/service_grpc_pb";
import { credentials } from "@grpc/grpc-js";
import { env, logger } from "./utils";

// TODO(nnnnat): needs correct grpc clients typing
type TClientClasses = Record<string, any>;
export type TServiceClients = Record<string, any>;

const CLIENT_CLASSES: TClientClasses = {
  app: AppsServiceClient,
  build: BuildsServiceClient,
  buildStatus: BuildsStatusServiceClient,
  component: ComponentsServiceClient,
  deploy: DeployServiceClient,
  deployment: DeploymentsServiceClient,
  deploymentStatus: DeploymentsStatusServiceClient,
  github: GithubServiceClient,
  install: InstallsServiceClient,
  installStatus: InstallsStatusServiceClient,
  instanceStatus: InstancesStatusServiceClient,
  org: OrgsServiceClient,
  orgStatus: OrgsStatusServiceClient,
  status: StatusServiceClient,
};

export function initServiceClients(
  services = env.SERVICES,
  clients = CLIENT_CLASSES
): TServiceClients {
  logger.debug("Initializing service clients");

  return services.reduce((acc, { name, url }) => {
    const client = clients[name];
    if (client) {
      acc[name] = new client(url, credentials.createInsecure());
    }

    return acc;
  }, {});
}
