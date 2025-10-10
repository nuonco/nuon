import { Card } from "@/components/common/Card";
import { RunnerDetailsCard } from "@/components/runners/RunnerDetailsCard";
import { Text } from "@/components/common/Text";
import { getRunnerById, getRunnerLatestHeartbeat } from "@/lib";
import type { TRunnerGroup } from "@/types";

export async function RunnerDetails({
  orgId,
  runnerId,
}: {
  orgId: string;
  runnerId: string;
}) {
  const [
    { data: runnerHeartbeat, error: runnerHeartbeatError },
    { data: runner, error: runnerError },
  ] = await Promise.all([
    getRunnerLatestHeartbeat({
      orgId,
      runnerId,
    }),
    getRunnerById({
      orgId,
      runnerId,
    }),
  ]);

  const error = runnerError || runnerHeartbeatError || null;

  return runner && !error ? (
    <RunnerDetailsCard
      className="md:flex-initial"
      initHeartbeat={runnerHeartbeat}
      runner={runner}
      runnerGroup={{ platform: "local" } as TRunnerGroup}
      shouldPoll
    />
  ) : (
    <RunnerDetailsError />
  );
}

export const RunnerDetailsError = () => (
  <Card className="flex-auto">
    <Text>Unable to load build runner</Text>
  </Card>
);
