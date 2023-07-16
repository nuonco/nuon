import {
  status as ESTATUS,
  InterceptingCall,
  credentials,
} from "@grpc/grpc-js";
import { AppsServiceClient } from "./build/api/app/v1/service_grpc_pb";
import { BuildsServiceClient } from "./build/api/build/v1/service_grpc_pb";
import { ComponentsServiceClient } from "./build/api/component/v1/service_grpc_pb";
import { DeployServiceClient } from "./build/api/deploy/v1/service_grpc_pb";
import { DeploymentsServiceClient } from "./build/api/deployment/v1/service_grpc_pb";
import { GithubServiceClient } from "./build/api/github/v1/service_grpc_pb";
import { InstallsServiceClient } from "./build/api/install/v1/service_grpc_pb";
import { InstancesServiceClient } from "./build/api/instance/v1/service_grpc_pb";
import { OrgsServiceClient } from "./build/api/org/v1/service_grpc_pb";
import { BuildsServiceClient as BuildsStatusServiceClient } from "./build/orgs-api/builds/v1/service_grpc_pb";
import { DeploymentsServiceClient as DeploymentsStatusServiceClient } from "./build/orgs-api/deployments/v1/service_grpc_pb";
import { InstallsServiceClient as InstallsStatusServiceClient } from "./build/orgs-api/installs/v1/service_grpc_pb";
import { InstancesServiceClient as InstancesStatusServiceClient } from "./build/orgs-api/instances/v1/service_grpc_pb";
import { OrgsServiceClient as OrgsStatusServiceClient } from "./build/orgs-api/orgs/v1/service_grpc_pb";
import { StatusServiceClient } from "./build/shared/status/v1/service_grpc_pb";
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
  instance: InstancesServiceClient,
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
      acc[name] = new client(url, credentials.createInsecure(), {
        interceptors: [
          (options, nextCall) =>
            new InterceptingCall(nextCall(options), {
              cancel(next) {
                next();
              },
              halfClose(next) {
                next();
              },
              sendMessage(message, next) {
                logger.debug(`${name} client sent message: ${message}`);
                next(message);
              },
              start(meta, listener, next) {
                next(meta, {
                  onReceiveMessage(message, next) {
                    logger.debug(`${name} client received message: ${message}`);
                    next(message);
                  },
                  onReceiveMetadata(metadata, next) {
                    logger.debug(
                      `${name} client received metadata: ${JSON.stringify(
                        metadata?.toJSON()
                      )}`
                    );
                    next(metadata);
                  },
                  onReceiveStatus(status, next) {
                    if (status?.code !== ESTATUS.OK) {
                      logger.error(`${name} client: ${status.details}`);
                    } else {
                      logger.debug(
                        `${name} client received status: ${status.code}`
                      );
                    }
                    next(status);
                  },
                });
              },
            }),
        ],
      });
    }

    return acc;
  }, {});
}
