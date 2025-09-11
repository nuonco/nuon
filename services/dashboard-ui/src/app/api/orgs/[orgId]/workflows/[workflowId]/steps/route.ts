import { type NextRequest, NextResponse } from "next/server";
import { getWorkflowSteps } from "@/lib";
import type { TRouteProps } from "@/types";

export async function GET(
  request: NextRequest,
  { params }: TRouteProps<"orgId" | "workflowId">,
) {
  const { orgId, workflowId } = await params;
  const { searchParams } = new URL(request.url);
  const limit = searchParams.get("limit") || undefined;
  const offset = searchParams.get("offset") || undefined;
  
  const response = await getWorkflowSteps({ orgId, workflowId, limit, offset });
  return NextResponse.json(response);
}
