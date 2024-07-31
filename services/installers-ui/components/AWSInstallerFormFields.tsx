import React, { type FC } from "react";
import { Select } from "@/components";
import { getFlagEmoji } from "@/common";

export const AWSRegionSelect: FC<{
  defaultValue?: string;
  regions: Array<any>;
}> = ({ defaultValue = "", regions = [] }) => {
  const options = regions.map((o) => ({
    value: o.value,
    label: `${getFlagEmoji(o.icon.substring(5))} ${o.display_name}`,
  }));

  return (
    <Select
      defaultValue={defaultValue}
      name="region"
      options={options}
      required
    />
  );
};

export const AWSInstallerFormFields: FC<{
  searchParams?: Record<string, string>;
  regions: Array<Object>;
  aws_account: any | null;
}> = ({ searchParams = {}, regions, aws_account }) => {
  return (
    <fieldset className="p-4 w-full">
      <label className="mb-2 flex flex-col flex-auto gap-2">
        <span className="text-sm font-medium">AWS IAM role ARN</span>
        <input
          className="border bg-inherit rounded px-4 py-1.5 shadow-inner"
          defaultValue={
            aws_account
              ? aws_account.iam_role_arn
              : Object.hasOwn(searchParams, "iam_role_arn")
                ? searchParams?.iam_role_arn
                : ""
          }
          name="iam_role_arn"
          type="text"
          required
        />
      </label>

      <label className="mb-2 flex flex-col flex-auto gap-2">
        <span className="text-sm font-medium">AWS Region</span>
        <AWSRegionSelect
          defaultValue={
            aws_account
              ? aws_account.region
              : Object.hasOwn(searchParams, "region")
                ? searchParams?.region
                : ""
          }
          regions={regions}
        />
      </label>
    </fieldset>
  );
};
