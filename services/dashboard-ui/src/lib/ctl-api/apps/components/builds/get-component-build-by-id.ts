import { api } from "@/lib/api";
import type { TBuild } from "@/types";

export const getComponentBuildById = ({
  componentId,
  buildId,
  orgId,
}: {
  componentId: string;
  buildId: string;
  orgId: string;
}) =>
  api<TBuild>({
    path: `components/${componentId}/builds/${buildId}`,
    orgId,
  });

