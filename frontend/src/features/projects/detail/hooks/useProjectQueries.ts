"use client";
import { baselineService, snapshotService } from "@/services/monitoring.service";
import { projectService } from "@/services/project.service";
import { useQuery } from "@tanstack/react-query";
import { useMemo } from "react";
import { TASK_STATUSES } from "../schemas";
import { toDateInput } from "../utils";

interface Options {
  projectId: string;
}

export function useProjectQueries({ projectId }: Options) {
  const projectQuery = useQuery({
    queryKey: ["projects", projectId],
    queryFn: () => projectService.get(projectId),
    select: (res) => res.data.data,
  });

  const tasksQuery = useQuery({
    queryKey: ["projects", projectId, "tasks"],
    queryFn: () => projectService.listTasks(projectId, { page: 1, page_size: 100 }),
    select: (res) => res.data.data ?? [],
  });

  const milestonesQuery = useQuery({
    queryKey: ["projects", projectId, "milestones"],
    queryFn: () => projectService.listMilestones(projectId),
    select: (res) => res.data.data ?? [],
  });

  const issuesQuery = useQuery({
    queryKey: ["projects", projectId, "issues"],
    queryFn: () => projectService.listIssues(projectId, { page: 1, page_size: 100 }),
    select: (res) => res.data.data ?? [],
  });

  const risksQuery = useQuery({
    queryKey: ["projects", projectId, "risks"],
    queryFn: () => projectService.listRisks(projectId, { page: 1, page_size: 100 }),
    select: (res) => res.data.data ?? [],
  });

  const budgetsQuery = useQuery({
    queryKey: ["projects", projectId, "budgets"],
    queryFn: () => projectService.listBudgets(projectId, { page: 1, page_size: 100 }),
    select: (res) => res.data.data ?? [],
  });

  const contractsQuery = useQuery({
    queryKey: ["projects", projectId, "contracts"],
    queryFn: () => projectService.listContracts(projectId, { page: 1, page_size: 100 }),
    select: (res) => res.data.data ?? [],
  });

  const documentsQuery = useQuery({
    queryKey: ["projects", projectId, "documents"],
    queryFn: () => projectService.listDocuments(projectId, { page: 1, page_size: 100 }),
    select: (res) => res.data.data ?? [],
  });

  const vendorsQuery = useQuery({
    queryKey: ["vendors", "all"],
    queryFn: () => projectService.listVendors({ page: 1, page_size: 100, is_active: true }),
    select: (res) => res.data.data ?? [],
  });

  const correctiveActionsQuery = useQuery({
    queryKey: ["projects", projectId, "corrective-actions"],
    queryFn: () => projectService.listCorrectiveActions(projectId, { page: 1, page_size: 100 }),
    select: (res) => res.data.data ?? [],
  });

  const tasks = useMemo(() => tasksQuery.data ?? [], [tasksQuery.data]);

  const milestones = useMemo(
    () => milestonesQuery.data ?? [],
    [milestonesQuery.data]
  );

  const issues = useMemo(() => issuesQuery.data ?? [], [issuesQuery.data]);

  const risks = useMemo(() => risksQuery.data ?? [], [risksQuery.data]);

  const budgets = useMemo(() => budgetsQuery.data ?? [], [budgetsQuery.data]);

  const contracts = useMemo(() => contractsQuery.data ?? [], [contractsQuery.data]);

  const documents = useMemo(() => documentsQuery.data ?? [], [documentsQuery.data]);

  const vendors = useMemo(() => vendorsQuery.data ?? [], [vendorsQuery.data]);

  const correctiveActions = useMemo(() => correctiveActionsQuery.data ?? [], [correctiveActionsQuery.data]);

  const vendorOptions = useMemo(
    () => vendors.filter((v) => v.type === "VENDOR" && v.is_active),
    [vendors]
  );

  const consultantOptions = useMemo(
    () => vendors.filter((v) => v.type === "CONSULTANT" && v.is_active),
    [vendors]
  );

  const contractTotals = useMemo(() => {
    let totalValue = 0;
    let activeCount = 0;
    let mainVendorId: string | null = null;
    const vendorValue: Record<string, number> = {};
    const sorted = [...contracts].sort((a, b) => b.contract_value - a.contract_value);
    for (const contract of contracts) {
      totalValue += contract.contract_value;
      if (contract.status === "ACTIVE") activeCount += 1;
      vendorValue[contract.vendor_id] =
        (vendorValue[contract.vendor_id] ?? 0) + contract.contract_value;
    }
    const main = sorted[0] ?? null;
    if (main) mainVendorId = main.vendor_id;
    let period = "–";
    if (contracts.length > 0) {
      const start = contracts.reduce<string | null>(
        (acc, c) => (c.start_date && (!acc || c.start_date < acc) ? c.start_date : acc),
        null
      );
      const end = contracts.reduce<string | null>(
        (acc, c) => (c.end_date && (!acc || c.end_date > acc) ? c.end_date : acc),
        null
      );
      if (start || end) period = `${start ? toDateInput(start) : "?"} — ${end ? toDateInput(end) : "?"}`;
    }
    return { totalValue, activeCount, mainVendorId, mainValue: main?.contract_value ?? 0, period };
  }, [contracts]);

  const budgetTotals = useMemo(() => {
    let planned = 0;
    let actual = 0;
    for (const line of budgets) {
      planned += line.planned;
      actual += line.actual;
    }
    const usage = planned > 0 ? Math.round((actual / planned) * 10000) / 100 : 0;
    return { planned, actual, variance: planned - actual, usage };
  }, [budgets]);

  const documentTotals = useMemo(() => {
    let totalSize = 0;
    const byCategory: Record<string, number> = {};
    for (const doc of documents) {
      totalSize += doc.file_size ?? 0;
      const cat = doc.category || "UNKNOWN";
      byCategory[cat] = (byCategory[cat] ?? 0) + 1;
    }
    return { count: documents.length, totalSize, byCategory };
  }, [documents]);

  const taskGroups = useMemo(
    () =>
      TASK_STATUSES.map((status) => ({
        status,
        tasks: tasks.filter((task) => task.status === status),
      })),
    [tasks]
  );

  const baselinesQuery = useQuery({
    queryKey: ["projects", projectId, "baselines"],
    queryFn: () => baselineService.list(projectId),
    select: (data) => data,
  });

  const baselines = baselinesQuery.data ?? [];

  const snapshotsQuery = useQuery({
    queryKey: ["projects", projectId, "snapshots"],
    queryFn: () => snapshotService.list(projectId),
    select: (data) => data,
  });

  const snapshots = snapshotsQuery.data ?? [];
  return {
    projectQuery,
    tasksQuery,
    milestonesQuery,
    issuesQuery,
    risksQuery,
    budgetsQuery,
    contractsQuery,
    documentsQuery,
    vendorsQuery,
    correctiveActionsQuery,
    tasks,
    milestones,
    issues,
    risks,
    budgets,
    contracts,
    documents,
    vendors,
    correctiveActions,
    vendorOptions,
    consultantOptions,
    contractTotals,
    budgetTotals,
    documentTotals,
    taskGroups,
    baselinesQuery,
    baselines,
    snapshotsQuery,
    snapshots,
  };
}
