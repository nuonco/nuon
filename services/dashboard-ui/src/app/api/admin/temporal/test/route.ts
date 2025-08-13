
import { NextRequest, NextResponse } from "next/server";

export async function GET(req: NextRequest, context: any) {
  return NextResponse.json({
    ok: true,
    params: context.params,
    url: req.url,
  });
}
