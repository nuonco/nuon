"use client";

import Link from "next/link";
import { useEffect } from "react";

export default function Error({
  error,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    console.error(error);
  }, [error]);

  return (
    <div className="flex flex-col flex-auto items-center justify-start py-24">
      <h1 className="text-2xl">Not found</h1>
      <p className="text-base">{error.message}</p>
      <p className="text-xs">
        If this issue persist conntact us at{" "}
        <Link href="mailto:support@nuon.co">support@nuon.co</Link>
      </p>
    </div>
  );
}
