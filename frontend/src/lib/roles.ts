import type { UserInfo } from "@/types/auth";

/** Role code of external contractor accounts (VENDOR-PORTAL-001). */
export const ROLE_VENDOR = "VENDOR";

/** True when the user is a contractor account that may only use the contractor portal. */
export function isVendorUser(user: Pick<UserInfo, "roles"> | null | undefined): boolean {
  const roles = user?.roles ?? [];
  return roles.includes(ROLE_VENDOR) && roles.every((r) => r === ROLE_VENDOR);
}

export const VENDOR_HOME = "/vendor-portal";

/** Internal routes a contractor account may open besides the portal. */
export const VENDOR_ALLOWED_PATHS = [VENDOR_HOME, "/notifications"];
