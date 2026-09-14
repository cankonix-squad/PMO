"use client";

import { cn } from "@/lib/utils";
import type { Priority, ProjectStatus } from "@/types/project";
import { z } from "zod";

export const STATUS_COLORS: Record<ProjectStatus, string> = {
  DRAFT: "bg-gray-100 text-gray-700",
  PLANNING: "bg-blue-100 text-blue-700",
  ACTIVE: "bg-green-100 text-green-700",
  ON_HOLD: "bg-yellow-100 text-yellow-700",
  COMPLETED: "bg-emerald-100 text-emerald-700",
  CANCELLED: "bg-red-100 text-red-700",
};

export const PRIORITIES: Priority[] = ["LOW", "MEDIUM", "HIGH", "CRITICAL"];

export const STATUS_LABELS: Record<ProjectStatus, string> = {
  DRAFT: "Draft",
  PLANNING: "Perencanaan",
  ACTIVE: "Aktif",
  ON_HOLD: "Ditunda",
  COMPLETED: "Selesai",
  CANCELLED: "Dibatalkan",
};

export const PRIORITY_LABELS: Record<Priority, string> = {
  LOW: "Rendah",
  MEDIUM: "Sedang",
  HIGH: "Tinggi",
  CRITICAL: "Kritis",
};

export const NEXT_STATUS: Record<ProjectStatus, ProjectStatus[]> = {
  DRAFT: ["PLANNING", "CANCELLED"],
  PLANNING: ["ACTIVE", "CANCELLED"],
  ACTIVE: ["ON_HOLD", "COMPLETED", "CANCELLED"],
  ON_HOLD: ["ACTIVE", "CANCELLED"],
  COMPLETED: ["ACTIVE"],
  CANCELLED: [],
};

export const projectFormSchema = z.object({
  code: z.string().trim().min(1, "Kode wajib diisi").max(100),
  name: z.string().trim().min(1, "Nama proyek wajib diisi").max(500),
  description: z.string().trim().min(1, "Deskripsi wajib diisi"),
  objectives: z.string().trim().min(1, "Tujuan wajib diisi"),
  priority: z.enum(["LOW", "MEDIUM", "HIGH", "CRITICAL"]),
  category: z.string().trim().min(1, "Kategori wajib dipilih"),
  start_date: z.string().trim().min(1, "Tanggal mulai wajib diisi"),
  end_date: z.string().trim().min(1, "Tanggal selesai wajib diisi"),
  budget_total: z.coerce.number().min(1, "Anggaran wajib lebih dari 0"),
  currency: z.string().trim().min(1).max(10),
  progress_pct: z.coerce.number().min(0).max(100),
  org_unit_id: z.string().trim().min(1, "Balai / Unit Pemilik wajib dipilih"),
  program_id: z.string().trim().min(1, "Program wajib dipilih"),
  sector_id: z.string().trim().min(1, "Sektor SDA wajib dipilih"),
  region_id: z.string().trim().min(1, "Wilayah wajib dipilih"),
  river_basin_id: z.string().trim().min(1, "DAS wajib dipilih"),
});

export type ProjectFormValues = z.infer<typeof projectFormSchema>;

export type FormMode = "create" | "edit";

export const inputClassName = cn(
  "h-9 w-full rounded-md border border-input bg-background px-3 text-sm",
  "placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring",
  "disabled:cursor-not-allowed disabled:bg-muted disabled:text-muted-foreground"
);
