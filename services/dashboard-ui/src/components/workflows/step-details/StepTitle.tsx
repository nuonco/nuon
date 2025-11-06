"use client";

import { Status } from "@/components/common/Status";
import { Text } from "@/components/common/Text";
import { toSentenceCase } from "@/utils/string-utils";
import { IStepDetails } from "./types";

export const StepTitle = ({ step }: IStepDetails) => {
  return (
    <span className="flex items-center gap-4 overflow-hidden w-72 md:w-fit">
      <Status
        isWithoutText
        status={step?.retried ? "retried" : step.status?.status || "unknown"}
        variant="timeline"
      />
      <Text className="!inline-block !text-nowrap truncate" variant="base">
        {toSentenceCase(step.name)}
      </Text>
    </span>
  );
};
