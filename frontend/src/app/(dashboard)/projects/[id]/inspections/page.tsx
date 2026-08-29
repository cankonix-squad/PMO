"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { ArrowLeft } from "lucide-react";
import { DashboardLayout } from "@/components/layout/DashboardLayout";
import { FieldInspectionPanel } from "@/components/project/FieldInspectionPanel";
import { projectService } from "@/services/project.service";

export default function ProjectInspectionsPage() {
  const params = useParams<{ id: string }>();
  const projectId = params.id;

  const { data: project, isLoading } = useQuery({
    queryKey: ["projects", projectId],
    queryFn: async () => {
      const res = await projectService.get(projectId);
      return res.data.data;
    },
    enabled: !!projectId,
  });

  return (
    <DashboardLayout title={project?.name ?? "Inspeksi Lapangan"}>
      <div className="mx-auto max-w-3xl px-4 py-4 sm:px-6">
        {/* Back nav */}
        <Link
          href={`/projects/${projectId}`}
          className="mb-4 inline-flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="h-4 w-4" />
          Kembali ke Detail Proyek
        </Link>

        {/* Project title */}
        <div className="mb-2">
          {isLoading ? (
            <div className="h-6 w-48 animate-pulse rounded bg-muted" />
          ) : (
            <h1 className="text-base font-semibold sm:text-lg">
              {project?.name ?? "Proyek"}
            </h1>
          )}
          {project?.code && (
            <p className="text-xs text-muted-foreground">{project.code}</p>
          )}
        </div>

        {/* Field inspection panel — the main content */}
        <FieldInspectionPanel projectId={projectId} />
      </div>
    </DashboardLayout>
  );
}

