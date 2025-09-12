import { type NextRequest, NextResponse } from "next/server";
import { getInstallSandboxRunById } from "@/lib";
import type { TRouteProps } from "@/types";

export async function GET(
  request: NextRequest,
  { params }: TRouteProps<"orgId" | "installId" | "runId">,
) {
  const { runId, orgId } = await params;

  const response = await getInstallSandboxRunById({
    runId,
    orgId,
  });
  return NextResponse.json(response);
}
