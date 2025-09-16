import { type NextRequest, NextResponse } from "next/server";
import { getRunnerRecentHealthChecks } from "@/lib";
import type { TRouteProps } from "@/types";

export async function GET(
  request: NextRequest,
  { params }: TRouteProps<"orgId" | "runnerId">,
) {
  const { runnerId, orgId } = await params;
  const { searchParams } = new URL(request.url);
  const limit = searchParams.get("limit") || undefined;
  const offset = searchParams.get("offset") || undefined;
  const window = searchParams.get("window") || undefined;
  
  const response = await getRunnerRecentHealthChecks({ 
    runnerId, 
    orgId, 
    limit, 
    offset,
    window
  });
  return NextResponse.json(response);
}
