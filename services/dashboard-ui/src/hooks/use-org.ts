"use client";

import { useContext } from "react";
import { OrgContext } from "@/providers/org-provider";

export function useOrg() {
  const ctx = useContext(OrgContext);
  if (!ctx) {
    throw new Error("useOrg must be used within an OrgProvider");
  }
  return ctx;
}
