"use client";

import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { CheckCircle2, Clock3, RefreshCw, ShieldAlert, XCircle } from 'lucide-react';
import { DashboardLayout } from '@/components/layout/DashboardLayout';
import { dataQualityService } from '@/services/dataquality.service';
import type { ValidationStatus, ValidationSubmission } from '@/types/dataquality';
import { cn, formatDate } from '@/lib/utils';

const statuses: Array<ValidationStatus | ''> = ['', 'SUBMITTED', 'VALID', 'REJECTED', 'STALE'];

const statusStyles: Record<ValidationStatus, string> = {
  DRAFT: 'bg-slate-100 text-slate-700',
  SUBMITTED: 'bg-amber-100 text-amber-800',
  VALID: 'bg-emerald-100 text-emerald-800',
  REJECTED: 'bg-rose-100 text-rose-800',
  STALE: 'bg-orange-100 text-orange-800',
};

function isOverdue(item: ValidationSubmission): boolean {
  return item.status === 'SUBMITTED' && Boolean(item.sla_due_at) && new Date(item.sla_due_at as string) < new Date();
}

export default function ValidationPage() {
  const queryClient = useQueryClient();
  const [status, setStatus] = useState<ValidationStatus | ''>('SUBMITTED');
  const [rejectionId, setRejectionId] = useState<string | null>(null);
  const [rejectionReason, setRejectionReason] = useState('');

  const queueQuery = useQuery({
    queryKey: ['validation-queue', status],
    queryFn: () => dataQualityService.list(status || undefined),
  });

  const transitionMutation = useMutation({
    mutationFn: ({ id, nextStatus, reason }: { id: string; nextStatus: 'VALID' | 'REJECTED' | 'STALE'; reason?: string }) =>
      dataQualityService.transition(id, { status: nextStatus, rejection_reason: reason }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['validation-queue'] });
      setRejectionId(null);
      setRejectionReason('');
    },
  });

  const items = queueQuery.data ?? [];
  const pendingCount = items.filter((item) => item.status === 'SUBMITTED').length;
  const overdueCount = items.filter(isOverdue).length;

  return (
    <DashboardLayout title="Validation Queue">
      <div className="space-y-6">
        <div className="flex flex-col justify-between gap-3 sm:flex-row sm:items-end">
          <div>
            <p className="text-sm text-muted-foreground">Review submission snapshot sebelum masuk ke data resmi.</p>
          </div>
          <button type="button" onClick={() => void queueQuery.refetch()} className="inline-flex h-10 items-center justify-center gap-2 rounded-md border px-3 text-sm font-medium hover:bg-muted">
            <RefreshCw className={cn('h-4 w-4', queueQuery.isFetching && 'animate-spin')} aria-hidden="true" />
            Muat Ulang
          </button>
        </div>

        <div className="grid gap-3 sm:grid-cols-3">
          <div className="rounded-lg border bg-card p-4"><p className="text-xs text-muted-foreground">Pending validation</p><p className="mt-1 text-2xl font-bold">{pendingCount}</p></div>
          <div className="rounded-lg border bg-card p-4"><p className="text-xs text-muted-foreground">SLA overdue</p><p className="mt-1 text-2xl font-bold text-rose-600">{overdueCount}</p></div>
          <div className="rounded-lg border bg-card p-4"><p className="text-xs text-muted-foreground">Displayed records</p><p className="mt-1 text-2xl font-bold">{items.length}</p></div>
        </div>

        <div className="flex flex-wrap gap-2" role="tablist" aria-label="Validation status filter">
          {statuses.map((value) => (
            <button key={value || 'all'} type="button" onClick={() => setStatus(value)} className={cn('rounded-md border px-3 py-2 text-sm', status === value ? 'border-primary bg-primary text-primary-foreground' : 'hover:bg-muted')}>
              {value || 'All'}
            </button>
          ))}
        </div>

        {queueQuery.isLoading && <div className="rounded-lg border p-8 text-center text-sm text-muted-foreground">Memuat antrean validasi...</div>}
        {queueQuery.isError && <div className="rounded-lg border border-rose-200 bg-rose-50 p-4 text-sm text-rose-700">Antrean validasi belum dapat dimuat.</div>}
        {!queueQuery.isLoading && !queueQuery.isError && items.length === 0 && <div className="rounded-lg border border-dashed p-10 text-center text-sm text-muted-foreground">Tidak ada submission validasi untuk filter ini.</div>}

        <div className="space-y-3">
          {items.map((item) => {
            const overdue = isOverdue(item);
            return (
              <article key={item.id} className="rounded-lg border bg-card p-4 shadow-sm">
                <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
                  <div className="min-w-0 space-y-2">
                    <div className="flex flex-wrap items-center gap-2">
                      <span className={cn('rounded-full px-2 py-1 text-xs font-semibold', statusStyles[item.status])}>{item.status}</span>
                      {overdue && <span className="inline-flex items-center gap-1 text-xs font-semibold text-rose-600"><Clock3 className="h-3.5 w-3.5" /> SLA overdue</span>}
                    </div>
                    <p className="font-semibold">Snapshot {item.period_year}-{String(item.period_month).padStart(2, '0')}</p>
                    <p className="break-all text-xs text-muted-foreground">Project {item.project_id} · Snapshot {item.snapshot_id}</p>
                    <div className="flex flex-wrap gap-x-4 gap-y-1 text-xs text-muted-foreground">
                      <span>Completeness {item.completeness_pct}%</span>
                      <span>Submitted {item.submitted_at ? formatDate(item.submitted_at) : '-'}</span>
                      <span>SLA {item.sla_due_at ? formatDate(item.sla_due_at) : '-'}</span>
                    </div>
                    {item.rejection_reason && <p className="text-sm text-rose-700">Reason: {item.rejection_reason}</p>}
                  </div>
                  {item.status === 'SUBMITTED' && (
                    <div className="flex shrink-0 flex-wrap gap-2">
                      <button type="button" onClick={() => transitionMutation.mutate({ id: item.id, nextStatus: 'VALID' })} disabled={transitionMutation.isPending} className="inline-flex h-9 items-center gap-1 rounded-md bg-emerald-600 px-3 text-sm font-medium text-white hover:bg-emerald-700 disabled:opacity-50"><CheckCircle2 className="h-4 w-4" /> Validate</button>
                      <button type="button" onClick={() => setRejectionId(item.id)} disabled={transitionMutation.isPending} className="inline-flex h-9 items-center gap-1 rounded-md border border-rose-200 px-3 text-sm font-medium text-rose-700 hover:bg-rose-50 disabled:opacity-50"><XCircle className="h-4 w-4" /> Reject</button>
                      <button type="button" onClick={() => transitionMutation.mutate({ id: item.id, nextStatus: 'STALE' })} disabled={transitionMutation.isPending} className="inline-flex h-9 items-center gap-1 rounded-md border px-3 text-sm font-medium hover:bg-muted disabled:opacity-50"><ShieldAlert className="h-4 w-4" /> Mark stale</button>
                    </div>
                  )}
                </div>
                {rejectionId === item.id && (
                  <div className="mt-4 flex flex-col gap-2 border-t pt-4 sm:flex-row">
                    <input value={rejectionReason} onChange={(event) => setRejectionReason(event.target.value)} placeholder="Rejection reason" className="h-9 min-w-0 flex-1 rounded-md border bg-background px-3 text-sm" />
                    <button type="button" disabled={!rejectionReason.trim() || transitionMutation.isPending} onClick={() => transitionMutation.mutate({ id: item.id, nextStatus: 'REJECTED', reason: rejectionReason.trim() })} className="h-9 rounded-md bg-rose-600 px-3 text-sm font-medium text-white disabled:opacity-50">Confirm rejection</button>
                  </div>
                )}
              </article>
            );
          })}
        </div>
      </div>
    </DashboardLayout>
  );
}
