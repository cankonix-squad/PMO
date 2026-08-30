"use client";

/**
 * Settings — Org Units page — P1-008
 *
 * Allows admins to manage Satker/Balai/BBWS/BWS and other org units.
 * Supports create, edit, toggle active, and soft-delete.
 */

import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  Building2,
  ChevronDown,
  ChevronRight,
  Pencil,
  Plus,
  Trash2,
  ToggleLeft,
  ToggleRight,
} from "lucide-react";
import { DashboardLayout } from "@/components/layout/DashboardLayout";
import { orgUnitService } from "@/services/org-unit.service";
import { cn } from "@/lib/utils";
import type {
  OrgUnit,
  OrgUnitLevel,
  CreateOrgUnitRequest,
  UpdateOrgUnitRequest,
} from "@/types/org-unit";
import { OrgUnitLevelLabel } from "@/types/org-unit";

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

/** Build a tree from a flat list of org units. */
function buildTree(units: OrgUnit[]): OrgUnit[] {
  const map = new Map<string, OrgUnit & { children: OrgUnit[] }>();
  units.forEach((u) => map.set(u.id, { ...u, children: [] }));

  const roots: OrgUnit[] = [];
  map.forEach((u) => {
    if (u.parent_id && map.has(u.parent_id)) {
      map.get(u.parent_id)!.children!.push(u);
    } else {
      roots.push(u);
    }
  });
  return roots;
}

const LEVEL_OPTIONS: { value: OrgUnitLevel; label: string }[] = (
  Object.entries(OrgUnitLevelLabel) as [string, string][]
).map(([v, label]) => ({ value: Number(v) as OrgUnitLevel, label }));

// ---------------------------------------------------------------------------
// Sub-components
// ---------------------------------------------------------------------------

interface OrgUnitRowProps {
  unit: OrgUnit;
  depth: number;
  allUnits: OrgUnit[];
  onEdit: (unit: OrgUnit) => void;
  onDelete: (id: string) => void;
  onToggleActive: (unit: OrgUnit) => void;
}

function OrgUnitRow({
  unit,
  depth,
  allUnits,
  onEdit,
  onDelete,
  onToggleActive,
}: OrgUnitRowProps) {
  const [expanded, setExpanded] = useState(true);
  const children = unit.children ?? [];
  const hasChildren = children.length > 0;

  return (
    <>
      <tr className={cn("border-b", !unit.is_active && "opacity-50")}>
        <td className="py-2 px-4">
          <div
            className="flex items-center gap-1"
            style={{ paddingLeft: `${depth * 20}px` }}
          >
            {hasChildren ? (
              <button
                onClick={() => setExpanded((p) => !p)}
                className="text-muted-foreground hover:text-foreground"
              >
                {expanded ? (
                  <ChevronDown className="h-4 w-4" />
                ) : (
                  <ChevronRight className="h-4 w-4" />
                )}
              </button>
            ) : (
              <span className="w-5" />
            )}
            <Building2 className="h-4 w-4 text-muted-foreground shrink-0" />
            <span className="font-medium text-sm">{unit.name}</span>
          </div>
        </td>
        <td className="py-2 px-4 text-sm text-muted-foreground">{unit.code}</td>
        <td className="py-2 px-4 text-sm text-muted-foreground">
          {OrgUnitLevelLabel[unit.level]}
        </td>
        <td className="py-2 px-4">
          <span
            className={cn(
              "inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium",
              unit.is_active
                ? "bg-green-100 text-green-700"
                : "bg-gray-100 text-gray-500"
            )}
          >
            {unit.is_active ? "Aktif" : "Nonaktif"}
          </span>
        </td>
        <td className="py-2 px-4">
          <div className="flex items-center gap-2 justify-end">
            <button
              onClick={() => onToggleActive(unit)}
              title={unit.is_active ? "Nonaktifkan" : "Aktifkan"}
              className="text-muted-foreground hover:text-foreground"
            >
              {unit.is_active ? (
                <ToggleRight className="h-4 w-4 text-green-600" />
              ) : (
                <ToggleLeft className="h-4 w-4" />
              )}
            </button>
            <button
              onClick={() => onEdit(unit)}
              title="Edit"
              className="text-muted-foreground hover:text-foreground"
            >
              <Pencil className="h-4 w-4" />
            </button>
            <button
              onClick={() => onDelete(unit.id)}
              title="Hapus"
              className="text-muted-foreground hover:text-red-600"
            >
              <Trash2 className="h-4 w-4" />
            </button>
          </div>
        </td>
      </tr>
      {expanded &&
        hasChildren &&
        children.map((child) => (
          <OrgUnitRow
            key={child.id}
            unit={child}
            depth={depth + 1}
            allUnits={allUnits}
            onEdit={onEdit}
            onDelete={onDelete}
            onToggleActive={onToggleActive}
          />
        ))}
    </>
  );
}

// ---------------------------------------------------------------------------
// Modal
// ---------------------------------------------------------------------------

interface ModalProps {
  mode: "create" | "edit";
  initial?: OrgUnit | null;
  allUnits: OrgUnit[];
  onClose: () => void;
  onSubmit: (data: CreateOrgUnitRequest | UpdateOrgUnitRequest) => void;
  isLoading: boolean;
}

function OrgUnitModal({
  mode,
  initial,
  allUnits,
  onClose,
  onSubmit,
  isLoading,
}: ModalProps) {
  const [code, setCode] = useState(initial?.code ?? "");
  const [name, setName] = useState(initial?.name ?? "");
  const [level, setLevel] = useState<OrgUnitLevel>(initial?.level ?? 5);
  const [parentId, setParentId] = useState<string>(
    initial?.parent_id ?? ""
  );
  const [sortOrder, setSortOrder] = useState(initial?.sort_order ?? 0);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const payload: CreateOrgUnitRequest = {
      code: code.trim(),
      name: name.trim(),
      level,
      parent_id: parentId || null,
      sort_order: sortOrder,
    };
    onSubmit(payload);
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
      <div className="bg-background rounded-lg shadow-xl w-full max-w-md p-6">
        <h2 className="text-lg font-semibold mb-4">
          {mode === "create" ? "Tambah Org Unit" : "Edit Org Unit"}
        </h2>
        <form onSubmit={handleSubmit} className="space-y-4">
          {/* Code */}
          <div>
            <label className="block text-sm font-medium mb-1">Kode</label>
            <input
              className="w-full rounded-md border px-3 py-2 text-sm bg-background focus:outline-none focus:ring-2 focus:ring-ring"
              value={code}
              onChange={(e) => setCode(e.target.value)}
              required
              maxLength={100}
              placeholder="cth. BBWS-CIMANUK"
            />
          </div>

          {/* Name */}
          <div>
            <label className="block text-sm font-medium mb-1">Nama</label>
            <input
              className="w-full rounded-md border px-3 py-2 text-sm bg-background focus:outline-none focus:ring-2 focus:ring-ring"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
              maxLength={300}
              placeholder="cth. BBWS Cimanuk Cisanggarung"
            />
          </div>

          {/* Level */}
          <div>
            <label className="block text-sm font-medium mb-1">Level</label>
            <select
              className="w-full rounded-md border px-3 py-2 text-sm bg-background focus:outline-none focus:ring-2 focus:ring-ring"
              value={level}
              onChange={(e) => setLevel(Number(e.target.value) as OrgUnitLevel)}
              required
            >
              {LEVEL_OPTIONS.map((o) => (
                <option key={o.value} value={o.value}>
                  {o.label}
                </option>
              ))}
            </select>
          </div>

          {/* Parent */}
          <div>
            <label className="block text-sm font-medium mb-1">
              Unit Induk (opsional)
            </label>
            <select
              className="w-full rounded-md border px-3 py-2 text-sm bg-background focus:outline-none focus:ring-2 focus:ring-ring"
              value={parentId}
              onChange={(e) => setParentId(e.target.value)}
            >
              <option value="">— Tidak ada (root) —</option>
              {allUnits
                .filter((u) => u.id !== initial?.id)
                .map((u) => (
                  <option key={u.id} value={u.id}>
                    {u.name} ({OrgUnitLevelLabel[u.level]})
                  </option>
                ))}
            </select>
          </div>

          {/* Sort order */}
          <div>
            <label className="block text-sm font-medium mb-1">Sort Order</label>
            <input
              type="number"
              className="w-full rounded-md border px-3 py-2 text-sm bg-background focus:outline-none focus:ring-2 focus:ring-ring"
              value={sortOrder}
              onChange={(e) => setSortOrder(Number(e.target.value))}
              min={0}
            />
          </div>

          {/* Actions */}
          <div className="flex justify-end gap-2 pt-2">
            <button
              type="button"
              onClick={onClose}
              className="rounded-md border px-4 py-2 text-sm hover:bg-muted"
            >
              Batal
            </button>
            <button
              type="submit"
              disabled={isLoading}
              className="rounded-md bg-primary text-primary-foreground px-4 py-2 text-sm font-medium hover:bg-primary/90 disabled:opacity-50"
            >
              {isLoading ? "Menyimpan..." : "Simpan"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

// ---------------------------------------------------------------------------
// Page
// ---------------------------------------------------------------------------

export default function OrgUnitsPage() {
  const qc = useQueryClient();
  const [showInactive, setShowInactive] = useState(false);
  const [modalMode, setModalMode] = useState<"create" | "edit" | null>(null);
  const [editTarget, setEditTarget] = useState<OrgUnit | null>(null);

  const { data: units = [], isLoading, isError } = useQuery({
    queryKey: ["org-units", showInactive],
    queryFn: () => orgUnitService.listOrgUnits(showInactive),
  });

  const tree = buildTree(units);

  const invalidate = () => {
    qc.invalidateQueries({ queryKey: ["org-units"] });
    // org_unit_name di project response di-resolve server-side —
    // invalidate semua cache yang meng-embed project list agar BALAI terupdate.
    qc.invalidateQueries({ queryKey: ["dashboard", "projects"] });
    qc.invalidateQueries({ queryKey: ["projects"] });
  };

  const createMutation = useMutation({
    mutationFn: (req: CreateOrgUnitRequest) => orgUnitService.createOrgUnit(req),
    onSuccess: () => { invalidate(); setModalMode(null); },
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, req }: { id: string; req: UpdateOrgUnitRequest }) =>
      orgUnitService.updateOrgUnit(id, req),
    onSuccess: () => { invalidate(); setModalMode(null); setEditTarget(null); },
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => orgUnitService.deleteOrgUnit(id),
    onSuccess: invalidate,
  });

  const handleModalSubmit = (data: CreateOrgUnitRequest | UpdateOrgUnitRequest) => {
    if (modalMode === "create") {
      createMutation.mutate(data as CreateOrgUnitRequest);
    } else if (modalMode === "edit" && editTarget) {
      updateMutation.mutate({ id: editTarget.id, req: data as UpdateOrgUnitRequest });
    }
  };

  const handleToggleActive = (unit: OrgUnit) => {
    updateMutation.mutate({
      id: unit.id,
      req: { is_active: !unit.is_active },
    });
  };

  const handleDelete = (id: string) => {
    if (!confirm("Hapus org unit ini? Data tidak dapat dipulihkan.")) return;
    deleteMutation.mutate(id);
  };

  const openCreate = () => {
    setEditTarget(null);
    setModalMode("create");
  };

  const openEdit = (unit: OrgUnit) => {
    setEditTarget(unit);
    setModalMode("edit");
  };

  const isMutating =
    createMutation.isPending ||
    updateMutation.isPending ||
    deleteMutation.isPending;

  return (
    <DashboardLayout title="Org Unit">
      <div className="space-y-4">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-xl font-semibold">Org Unit</h1>
            <p className="text-sm text-muted-foreground mt-0.5">
              Kelola Satker, Balai, BBWS, BWS, dan unit lainnya dalam organisasi.
            </p>
          </div>
          <div className="flex items-center gap-3">
            <label className="flex items-center gap-2 text-sm cursor-pointer select-none">
              <input
                type="checkbox"
                checked={showInactive}
                onChange={(e) => setShowInactive(e.target.checked)}
                className="rounded"
              />
              Tampilkan nonaktif
            </label>
            <button
              onClick={openCreate}
              className="inline-flex items-center gap-1.5 rounded-md bg-primary text-primary-foreground px-3 py-2 text-sm font-medium hover:bg-primary/90"
            >
              <Plus className="h-4 w-4" />
              Tambah Unit
            </button>
          </div>
        </div>

        {/* Table */}
        <div className="rounded-lg border overflow-hidden">
          {isLoading ? (
            <div className="py-16 text-center text-sm text-muted-foreground">
              Memuat data...
            </div>
          ) : isError ? (
            <div className="py-16 text-center text-sm text-destructive">
              Gagal memuat org unit. Coba refresh halaman.
            </div>
          ) : units.length === 0 ? (
            <div className="py-16 text-center text-sm text-muted-foreground">
              <Building2 className="h-8 w-8 mx-auto mb-2 opacity-30" />
              <p>Belum ada org unit.</p>
              <button
                onClick={openCreate}
                className="mt-3 text-primary hover:underline text-sm"
              >
                Tambah unit pertama
              </button>
            </div>
          ) : (
            <table className="w-full text-left">
              <thead className="bg-muted/50 text-xs uppercase tracking-wide text-muted-foreground">
                <tr>
                  <th className="py-2 px-4 w-1/2">Nama</th>
                  <th className="py-2 px-4">Kode</th>
                  <th className="py-2 px-4">Level</th>
                  <th className="py-2 px-4">Status</th>
                  <th className="py-2 px-4 text-right">Aksi</th>
                </tr>
              </thead>
              <tbody>
                {tree.map((unit) => (
                  <OrgUnitRow
                    key={unit.id}
                    unit={unit}
                    depth={0}
                    allUnits={units}
                    onEdit={openEdit}
                    onDelete={handleDelete}
                    onToggleActive={handleToggleActive}
                  />
                ))}
              </tbody>
            </table>
          )}
        </div>
      </div>

      {/* Modal */}
      {modalMode && (
        <OrgUnitModal
          mode={modalMode}
          initial={editTarget}
          allUnits={units}
          onClose={() => { setModalMode(null); setEditTarget(null); }}
          onSubmit={handleModalSubmit}
          isLoading={isMutating}
        />
      )}
    </DashboardLayout>
  );
}
