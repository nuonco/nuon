import { type NextRequest, NextResponse } from "next/server";
import { getInstallActionRunById } from "@/lib";
import type { TRouteProps } from "@/types";

export async function GET(
  request: NextRequest,
  { params }: TRouteProps<"orgId" | "installId" | "runId">,
) {
  const { installId, runId, orgId } = await params;

  const response = await getInstallActionRunById({
    installId,
    runId,
    orgId,
  });
  return NextResponse.json(response);
}
