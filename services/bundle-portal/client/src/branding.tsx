import { useEffect, type ReactNode } from "react";
import { useQuery } from "@tanstack/react-query";
import { Loading } from "@/components/common/Loading";
import { getPortalBranding } from "./api";

const mix = (color: string, target: "#ffffff" | "#000000", weight: number) => {
  const channels = [1, 3, 5].map((offset) =>
    Number.parseInt(color.slice(offset, offset + 2), 16),
  );
  const targetValue = target === "#ffffff" ? 255 : 0;
  return `#${channels
    .map((channel) =>
      Math.round(channel + (targetValue - channel) * weight)
        .toString(16)
        .padStart(2, "0"),
    )
    .join("")}`;
};

const applyPrimaryColor = (color: string) => {
  const root = document.documentElement;
  const colors: Record<string, string> = {
    "50": mix(color, "#ffffff", 0.96),
    "100": mix(color, "#ffffff", 0.9),
    "200": mix(color, "#ffffff", 0.8),
    "300": mix(color, "#ffffff", 0.65),
    "400": mix(color, "#ffffff", 0.35),
    "500": mix(color, "#ffffff", 0.15),
    "600": color,
    "700": mix(color, "#000000", 0.1),
    "800": mix(color, "#000000", 0.2),
    "900": mix(color, "#000000", 0.35),
    "950": mix(color, "#000000", 0.55),
  };
  for (const [shade, value] of Object.entries(colors)) {
    root.style.setProperty(`--color-primary-${shade}`, value);
    root.style.setProperty(`--primary-${shade}`, value);
  }
};

export const usePortalBranding = () =>
  useQuery({
    queryKey: ["portal-branding"],
    queryFn: getPortalBranding,
    staleTime: Infinity,
  });

export const BrandingBoundary = ({ children }: { children: ReactNode }) => {
  const branding = usePortalBranding();

  useEffect(() => {
    if (!branding.data) return;
    document.title = branding.data.name;
    applyPrimaryColor(branding.data.primary_color);
    if (branding.data.favicon_url) {
      let favicon = document.querySelector<HTMLLinkElement>('link[rel="icon"]');
      if (!favicon) {
        favicon = document.createElement("link");
        favicon.rel = "icon";
        document.head.append(favicon);
      }
      favicon.href = branding.data.favicon_url;
    }
  }, [branding.data]);

  if (branding.isLoading)
    return (
      <div className="flex min-h-screen items-center justify-center">
        <Loading variant="large" />
      </div>
    );
  return children;
};
