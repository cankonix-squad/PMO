"use client";

import type { CreateProjectRequest, Priority, UpdateProjectRequest } from "@/types/project";
import type { FormEvent } from "react";
import type { ProjectFormValues } from "./schemas";

export function sanitizeRichText(html: string): string {
  if (typeof window === "undefined") return "";
  const doc = new DOMParser().parseFromString(html, "text/html");
  const allowed = new Set(["P", "BR", "STRONG", "EM", "UL", "OL", "LI", "HR"]);
  doc.body.querySelectorAll("*").forEach((node) => {
    if (!allowed.has(node.tagName)) { node.replaceWith(...Array.from(node.childNodes)); return; }
    Array.from(node.attributes).forEach((attr) => node.removeAttribute(attr.name));
  });
  return doc.body.innerHTML;
}

export function formatBytes(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

export function formValues(form: HTMLFormElement): ProjectFormValues {
  const data = new FormData(form);
  return {
    code: getString(data, "code"),
    name: getString(data, "name"),
    description: getString(data, "description"),
    objectives: getString(data, "objectives"),
    priority: getString(data, "priority") as Priority,
    category: getString(data, "category"),
    start_date: getString(data, "start_date"),
    end_date: getString(data, "end_date"),
    budget_total: parseRupiahInput(getString(data, "budget_total")),
    currency: getString(data, "currency") || "IDR",
    progress_pct: Number(getString(data, "progress_pct") || 0),
    org_unit_id: getString(data, "org_unit_id"),
    program_id: getString(data, "program_id"),
    sector_id: getString(data, "sector_id"),
    region_id: getString(data, "region_id"),
    river_basin_id: getString(data, "river_basin_id"),
    latitude: getString(data, "latitude"),
    longitude: getString(data, "longitude"),
    province: getString(data, "province"),
    location_name: getString(data, "location_name"),
  };
}

export function getString(data: FormData, key: string) {
  const value = data.get(key);
  return typeof value === "string" ? value : "";
}

export function toCreatePayload(values: ProjectFormValues): CreateProjectRequest {
  return {
    code: values.code,
    name: values.name,
    description: optionalString(values.description),
    objectives: optionalString(values.objectives),
    priority: values.priority,
    category: optionalString(values.category),
    start_date: optionalString(values.start_date),
    end_date: optionalString(values.end_date),
    budget_total: values.budget_total,
    currency: values.currency,
    progress_pct: values.progress_pct,
    org_unit_id: optionalString(values.org_unit_id),
    program_id: optionalString(values.program_id),
    sector_id: optionalString(values.sector_id),
    region_id: optionalString(values.region_id),
    river_basin_id: optionalString(values.river_basin_id),
    latitude: optionalNumber(values.latitude),
    longitude: optionalNumber(values.longitude),
    province: optionalString(values.province),
    location_name: optionalString(values.location_name),
  };
}

export function toUpdatePayload(values: ProjectFormValues): UpdateProjectRequest {
  return {
    name: values.name,
    description: optionalString(values.description),
    objectives: optionalString(values.objectives),
    priority: values.priority,
    category: optionalString(values.category),
    start_date: optionalString(values.start_date),
    end_date: optionalString(values.end_date),
    budget_total: values.budget_total,
    currency: values.currency,
    progress_pct: values.progress_pct,
    org_unit_id: optionalString(values.org_unit_id),
    program_id: optionalString(values.program_id),
    sector_id: optionalString(values.sector_id),
    region_id: optionalString(values.region_id),
    river_basin_id: optionalString(values.river_basin_id),
    latitude: optionalNumber(values.latitude),
    longitude: optionalNumber(values.longitude),
    province: optionalString(values.province),
    location_name: optionalString(values.location_name),
  };
}

export function optionalString(value: string | undefined) {
  const trimmed = value?.trim();
  return trimmed ? trimmed : undefined;
}

export function optionalNumber(value: string | undefined) {
  const trimmed = value?.trim();
  if (!trimmed) return undefined;
  const n = Number(trimmed);
  return Number.isFinite(n) ? n : undefined;
}

export function parseRupiahInput(value: string) {
  return Number(value.replace(/[^\d]/g, "") || 0);
}

export function formatRupiahInput(value: number) {
  if (!Number.isFinite(value) || value <= 0) {
    return "";
  }
  return `Rp. ${Math.round(value).toLocaleString("en-US")},-`;
}

export function formatRupiahInputEvent(event: FormEvent<HTMLInputElement>) {
  const input = event.currentTarget;
  input.value = formatRupiahInput(parseRupiahInput(input.value));
}

export function toDateInput(value: string | null | undefined) {
  if (!value) return "";
  return value.slice(0, 10);
}
