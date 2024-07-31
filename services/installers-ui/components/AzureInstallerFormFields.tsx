import React, { type FC } from "react";
import { Select } from "@/components";
import { getFlagEmoji } from "@/common";

export const AzureLocationSelect: FC<{
  defaultValue?: string;
  regions: Array<Record<string, any>>;
}> = ({ defaultValue = "", regions = [] }) => {
  const options = regions.map((o) => ({
    value: o.value,
    label: `${getFlagEmoji(o.icon.substring(5))} ${o.display_name}`,
  }));

  return (
    <Select
      defaultValue={defaultValue}
      name="location"
      options={options}
      required
    />
  );
};

export const AzureInstallerFormFields: FC<{
  searchParams?: Record<string, string>;
  regions: Array<Record<string, any>>;
  azure_account: any | null;
}> = ({ searchParams = {}, regions, azure_account }) => {
  return (
    <fieldset className="p-4 w-full">
      <label className="mb-2 flex flex-col flex-auto gap-2">
        <span className="text-sm font-medium">Location</span>
        <AzureLocationSelect
          defaultValue={
            Object.hasOwn(searchParams, "location")
              ? searchParams?.location
              : ""
          }
          regions={regions}
        />
      </label>

      <label className="mb-2 flex flex-col flex-auto gap-2">
        <span className="text-sm font-medium">Service principal app ID</span>
        <input
          className="border bg-inherit rounded px-4 py-1.5 shadow-inner"
          defaultValue={
            azure_account
              ? azure_account.service_principal_app_id
              : Object.hasOwn(searchParams, "service_principal_app_id")
                ? searchParams?.service_principal_app_id
                : ""
          }
          name="service_principal_app_id"
          type="text"
          required
        />
      </label>

      <label className="mb-2 flex flex-col flex-auto gap-2">
        <span className="text-sm font-medium">Service principal password</span>
        <input
          className="border bg-inherit rounded px-4 py-1.5 shadow-inner"
          defaultValue={
            azure_account
              ? azure_account.service_principal_password
              : Object.hasOwn(searchParams, "service_principal_password")
                ? searchParams?.service_principal_password
                : ""
          }
          name="service_principal_password"
          type="password"
          required
        />
      </label>

      <label className="mb-2 flex flex-col flex-auto gap-2">
        <span className="text-sm font-medium">Subscription ID</span>
        <input
          className="border bg-inherit rounded px-4 py-1.5 shadow-inner"
          defaultValue={
            azure_account
              ? azure_account.subscription_id
              : Object.hasOwn(searchParams, "subscription_id")
                ? searchParams?.subscription_id
                : ""
          }
          name="subscription_id"
          type="text"
          required
        />
      </label>

      <label className="mb-2 flex flex-col flex-auto gap-2">
        <span className="text-sm font-medium">Subscription tenant ID</span>
        <input
          className="border bg-inherit rounded px-4 py-1.5 shadow-inner"
          defaultValue={
            azure_account
              ? azure_account.subscription_tenant_id
              : Object.hasOwn(searchParams, "subscription_tenant_id")
                ? searchParams?.subscription_tenant_id
                : ""
          }
          name="subscription_tenant_id"
          type="text"
          required
        />
      </label>
    </fieldset>
  );
};
