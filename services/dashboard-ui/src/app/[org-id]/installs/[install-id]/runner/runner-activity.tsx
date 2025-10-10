import { RunnerRecentActivity } from "@/components/runners/RunnerRecentActivity";
import { Text } from "@/components/common/Text";
import { getRunnerJobs } from "@/lib";

interface ILoadRunnerRecentActivity {
  offset: string;
  orgId: string;
  runnerId: string;
}

export async function RunnerActivity({
  orgId,
  offset,
  runnerId,
}: ILoadRunnerRecentActivity) {
  const {
    data: jobs,
    error,
    headers,
  } = await getRunnerJobs({ orgId, runnerId, offset });

  const pagination = {
    hasNext: headers?.['x-nuon-page-next'] === 'true',
    offset: Number(headers?.['x-nuon-page-offset'] ?? '0'),
  }

  return jobs && !error ? (
    <>
      <Text variant="base" weight="strong">
        Recent activity
      </Text>
      <RunnerRecentActivity
        initJobs={jobs}
        pagination={pagination}
        runnerId={runnerId}
        shouldPoll
      />
    </>
  ) : (
    <RunnerActivityError />
  );
}

export const RunnerActivityError = () => (
  <div className="w-full">
    <Text>Error fetching recenty runner activity </Text>
  </div>
);
