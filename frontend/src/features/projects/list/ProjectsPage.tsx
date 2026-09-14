"use client";

import { DashboardLayout } from "@/components/layout/DashboardLayout";
import { cn, formatIDR } from "@/lib/utils";
import { Plus, Search } from "lucide-react";
import Link from "next/link";
import { EmptyState } from "./components/EmptyState";
import { ErrorState } from "./components/ErrorState";
import { ProgressCell } from "./components/ProgressCell";
import { ProjectDetail } from "./components/ProjectDetail";
import { ProjectForm } from "./components/ProjectForm";
import { StatusBadge } from "./components/StatusBadge";
import { useProjects } from "./hooks/useProjects";

export default function ProjectsPage() {
  const {
    projects,
    search,
    setSearch,
    startCreate,
    showForm,
    formMode,
    selectedProject,
    formError,
    isSaving,
    orgUnits,
    programs,
    sectors,
    regions,
    riverBasins,
    projectCategories,
    descAttachments,
    objAttachments,
    setDescAttachments,
    setObjAttachments,
    handleSubmit,
    setShowForm,
    setFormError,
    projectListQuery,
    setSelectedId,
    selectedId,
    detailQuery,
    historyQuery,
    transitionMutation,
    startEdit,
  } = useProjects();

  return (
    <DashboardLayout title="Proyek">
      <div className="mb-6 flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h2 className="text-lg font-semibold text-foreground">Ruang Kerja Proyek</h2>
          <p className="text-sm text-muted-foreground">
            {projects.length} proyek dalam cakupan
          </p>
        </div>
        <div className="flex flex-col gap-2 sm:flex-row sm:items-center">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <input
              type="search"
              placeholder="Cari proyek"
              value={search}
              onChange={(event) => setSearch(event.target.value)}
              className={cn(
                "h-9 w-full rounded-md border border-input bg-background pl-9 pr-3 text-sm sm:w-56",
                "placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
              )}
            />
          </div>
          <button
            type="button"
            onClick={startCreate}
            className={cn(
              "inline-flex h-9 items-center justify-center gap-2 rounded-md bg-primary px-3 text-sm font-medium text-primary-foreground",
              "transition-colors hover:bg-primary/90 focus:outline-none focus:ring-2 focus:ring-ring"
            )}
          >
            <Plus className="h-4 w-4" aria-hidden="true" />
            Proyek Baru
          </button>
        </div>
      </div>

      {showForm && (
        <ProjectForm
          key={formMode === "edit" ? (selectedProject?.id ?? "edit") : "create"}
          mode={formMode}
          project={formMode === "edit" ? selectedProject : null}
          error={formError}
          isSaving={isSaving}
          orgUnits={orgUnits}
          programs={programs}
          sectors={sectors}
          regions={regions}
          riverBasins={riverBasins}
          projectCategories={projectCategories}
          descAttachments={descAttachments}
          objAttachments={objAttachments}
          onDescAttach={(file) => setDescAttachments((prev) => [...prev, file])}
          onObjAttach={(file) => setObjAttachments((prev) => [...prev, file])}
          onDescRemove={(idx) => setDescAttachments((prev) => prev.filter((_, i) => i !== idx))}
          onObjRemove={(idx) => setObjAttachments((prev) => prev.filter((_, i) => i !== idx))}
          onSubmit={handleSubmit}
          onCancel={() => {
            setShowForm(false);
            setFormError(null);
            setDescAttachments([]);
            setObjAttachments([]);
          }}
        />
      )}

      <div className="grid grid-cols-1 gap-6 xl:grid-cols-[minmax(0,1.25fr)_minmax(360px,0.75fr)]">
        <section className="rounded-lg border border-border bg-card shadow-sm">
          <div className="border-b border-border px-4 py-3">
            <h3 className="text-sm font-semibold text-foreground">Daftar Proyek</h3>
          </div>
          {projectListQuery.isLoading ? (
            <div className="p-6 text-sm text-muted-foreground">Memuat daftar proyek...</div>
          ) : projectListQuery.isError ? (
            <ErrorState message="Daftar proyek belum dapat dimuat." />
          ) : projects.length === 0 ? (
            <EmptyState
              search={search}
              onCreate={startCreate}
            />
          ) : (
            <div className="overflow-hidden">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-border bg-muted/40">
                    <th className="px-4 py-3 text-left font-medium text-muted-foreground">
                      Proyek
                    </th>
                    <th className="hidden px-4 py-3 text-left font-medium text-muted-foreground sm:table-cell">
                      Status
                    </th>
                    <th className="hidden px-4 py-3 text-left font-medium text-muted-foreground md:table-cell">
                      Progres
                    </th>
                    <th className="hidden px-4 py-3 text-left font-medium text-muted-foreground lg:table-cell">
                      Anggaran
                    </th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border">
                  {projects.map((project) => (
                    <tr
                      key={project.id}
                      onClick={() => setSelectedId(project.id)}
                      className={cn(
                        "cursor-pointer transition-colors hover:bg-muted/30",
                        selectedId === project.id && "bg-muted/40"
                      )}
                    >
                      <td className="px-4 py-3">
                        <div>
                          <Link
                            href={`/projects/${project.id}`}
                            className="font-medium text-foreground hover:text-primary hover:underline"
                            onClick={(event) => event.stopPropagation()}
                          >
                            {project.name}
                          </Link>
                          <p className="text-xs text-muted-foreground">{project.code}</p>
                        </div>
                      </td>
                      <td className="hidden px-4 py-3 sm:table-cell">
                        <StatusBadge status={project.status} />
                      </td>
                      <td className="hidden px-4 py-3 md:table-cell">
                        <ProgressCell value={project.progress_pct} />
                      </td>
                      <td className="hidden px-4 py-3 text-muted-foreground lg:table-cell">
                        {formatIDR(project.budget_total)}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </section>

        <ProjectDetail
          project={selectedProject}
          isLoading={detailQuery.isLoading}
          history={historyQuery.data ?? []}
          historyLoading={historyQuery.isLoading}
          transitionPending={transitionMutation.isPending}
          onEdit={startEdit}
          onTransition={(project, status) =>
            transitionMutation.mutate({
              id: project.id,
              payload: { to_status: status, comment: `Status diubah ke ${status}` },
            })
          }
        />
      </div>
    </DashboardLayout>
  );
}
