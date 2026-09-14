"use client";

import { Filter, Plus } from "lucide-react";

export function EmptyState({ search, onCreate }: { search: string; onCreate: () => void }) {
  return (
    <div className="flex flex-col items-center justify-center py-24 text-center">
      <div className="mb-4 rounded-full bg-muted p-4">
        <Filter className="h-6 w-6 text-muted-foreground" aria-hidden="true" />
      </div>
      <p className="text-sm font-medium text-foreground">Proyek tidak ditemukan</p>
      <p className="mt-1 text-sm text-muted-foreground">
        {search ? "Coba kata kunci lain." : "Tambahkan proyek pertama untuk memulai."}
      </p>
      {!search && (
        <button
          type="button"
          onClick={onCreate}
          className="mt-4 inline-flex h-9 items-center justify-center gap-2 rounded-md bg-primary px-3 text-sm font-medium text-primary-foreground hover:bg-primary/90"
        >
          <Plus className="h-4 w-4" aria-hidden="true" />
          Proyek Baru
        </button>
      )}
    </div>
  );
}
