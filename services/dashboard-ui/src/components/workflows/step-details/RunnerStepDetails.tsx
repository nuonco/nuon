"use client";

import { Icon } from "@/components/common/Icon";
import { Link } from "@/components/common/Link";
import { Text } from "@/components/common/Text";
import { RunnerDetailsCard } from "@/components/runners/RunnerDetailsCard";
import { RunnerDetailsCardSkeleton } from "@/components/runners/RunnerDetailsCardSkeleton";
import { RunnerHealthCard } from "@/components/runners/RunnerHealthCard";
import { RunnerHealthCardSkeleton } from "@/components/runners/RunnerHealthCardSkeleton";
import { useOrg } from "@/hooks/use-org";
import { useQuery } from "@/hooks/use-query";
import type { IStepDetails } from "./types";

interface IRunnerStepDetails extends IStepDetails {}

export const RunnerStepDetails = ({ step }: IRunnerStepDetails) => {
  const { org } = useOrg();
  const { data: runner, isLoading: isRunnerLoading } = useQuery({
    path: `/api/orgs/${org.id}/runners/${step.step_target_id}`,
  });
  const { data: runnerHeartbeat, isLoading: isHeartbeatLoading } = useQuery({
    path: `/api/orgs/${org.id}/runners/${step.step_target_id}/heartbeat`,
  });
  const { data: runnerHealthCheck, isLoading: isHealthCheckLoading } = useQuery(
    {
      path: `/api/orgs/${org.id}/runners/${step.step_target_id}/health-checks`,
    },
  );

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center gap-4">
        <Text variant="base" weight="strong">
          Install runner
        </Text>

        <Text variant="subtext">
          <Link href={`/${org.id}/installs/${step.owner_id}/runner`}>
            View runner <Icon variant="CaretRight" />
          </Link>
        </Text>
      </div>
      <div className="flex flex-col @min-4xl:flex-row gap-6">
        {isRunnerLoading || isHeartbeatLoading || !runner ? (
          <RunnerDetailsCardSkeleton />
        ) : (
          <RunnerDetailsCard
            runner={runner}
            initHeartbeat={runnerHeartbeat}
            runnerGroup={{ platform: "local" }}
            shouldPoll
          />
        )}

        {isHealthCheckLoading || !runnerHealthCheck || !runner ? (
          <RunnerHealthCardSkeleton />
        ) : (
          <RunnerHealthCard
            runnerId={runner?.id}
            initHealthchecks={runnerHealthCheck}
            shouldPoll
          />
        )}
      </div>
    </div>
  );
};
