import { type NextRequest, NextResponse } from "next/server";

const NUON_API_URL = process?.env?.NUON_API_URL || "https://ctl.prod.nuon.co";

const handler = async (req: NextRequest, res: NextResponse) => {
  console.log(req.url);
  const [installId] = req.url.split("/");
  let result = await fetch(`${NUON_API_URL}/v1/`, {
    headers: {
      Authorization: `Bearer ${process?.env?.NUON_API_TOKEN}`,
      // "X-Nuon-Org-ID": `${process?.env?.NUON_ORG_ID}`,
    },
  });

  return NextResponse.json(await result.json());
};
export default handler;
