import React, { type FC } from "react";
import { Link } from "@/components";
import { NuonLogotype } from "@/components/NuonLogotype";

export const PoweredByNuon: FC = ({}) => {
  return (
    <span className="text-xs">
      Powered by{" "}
      <Link
        href="https://www.nuon.co/"
        className="text-xs"
        target="_blank"
        rel="noreferrer"
      >
        <NuonLogotype className="inline w-auto h-6 ml-1 -mt-1.5 fill-black dark:fill-white" />
      </Link>
    </span>
  );
};
