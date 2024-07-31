/*
 *
 * Catch All API Proxy that injects the Hosted Installer Service Account Static Token
 *
 */
import { type NextRequest, NextResponse } from "next/server";

const NUON_API_URL = process?.env?.NUON_API_URL || "https://ctl.prod.nuon.co";

export async function GET(req: NextRequest) {
  // This endpoint SHOULD be a pure proxy
  console.log(`[get] method=${req.method} url=${req.url}`);
  let url = new URL(req.url);
  let path = url.pathname.replace("/api/v1/", "");
  let newUrl = `${NUON_API_URL}/v1/${path}`;
  const [installId] = req.url.split("/");
  let headers = {
    "Content-Type": "application/json",
    // NOTE: disabled for the time being
    // Authorization: `Bearer ${process?.env?.NUON_API_TOKEN}`,
    "X-Nuon-Org-Id": "", // we have to grab this somewhere and pass it
  };
  let result = await fetch(newUrl, {
    method: "get",
    cache: "no-store",
    headers: headers,
  });

  if (299 < result.status) {
    console.error(result);
  }
  return NextResponse.json(await result.json());
}

export async function POST(req: NextRequest) {
  console.log(`[post] method=${req.method} url=${req.url}`);
  let url = new URL(req.url);
  let path = url.pathname.replace("/api/v1/", "");
  let newUrl = `${NUON_API_URL}/v1/${path}`;
  const [installId] = req.url.split("/");
  let headers = {
    "Content-Type": "application/json",
    // NOTE: disabled for the time being
    // Authorization: `Bearer ${process?.env?.NUON_API_TOKEN}`,
    "X-Nuon-Org-Id": "",
  };
  let result = await fetch(newUrl, {
    method: "post",
    cache: "no-store",
    headers: headers,
  });

  if (299 < result.status) {
    console.error(result);
  }
  return NextResponse.json(await result.json());
}
