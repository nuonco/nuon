import { Card } from "@/components/common/Card";
import { RunnerHealthCard } from "@/components/runners/RunnerHealthCard";
import { EmptyState } from "@/components/common/EmptyState";
import { getRunnerRecentHealthChecks } from "@/lib";

export async function RunnerHealth({
  orgId,
  runnerId,
}: {
  orgId: string;
  runnerId: string;
}) {
  const { data: healthchecks, error } = await getRunnerRecentHealthChecks({
    orgId,
    runnerId,
  });

  return !error ? (
    <RunnerHealthCard
      className="flex-auto"
      initHealthchecks={healthchecks}
      runnerId={runnerId}
      shouldPoll
    />
  ) : (
    <RunnerHealthError />
  );
}

export const RunnerHealthError = () => (
  <Card className="flex-auto">
    <EmptyState
      emptyMessage="Runner health checks will display here once available."
      emptyTitle="No health check data"
    />
  </Card>
);
