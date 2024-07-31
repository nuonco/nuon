import Link from "next/link";
import React, { type FC } from "react";

export const StepOneAzure: FC = () => {
  return (
    <div className="flex flex-col gap-12">
      <div className="flex flex-col gap-4">
        <h2 className="text-xl font-semibold">
          Step 1: create your service principals
        </h2>
        <p className="text-base">
          Installs are created using service principals that nuon can use to
          provision the sandbox. Read more about how to set these principals up{" "}
          <Link
            className="text-primary-600 hover:text-primary-900 focus:text-primary-900"
            href="https://docs.nuon.co/guides/azure-support#create-service-principal"
            target="_blank"
            rel="noreferrer"
          >
            here
          </Link>
          .
        </p>
      </div>
    </div>
  );
};
