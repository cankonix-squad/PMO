'use client'

import { useState, useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import dynamic from 'next/dynamic'
import { MapPin, AlertTriangle, CheckCircle, Clock, BarChart2, RefreshCw } from 'lucide-react'
import { DashboardLayout } from '@/components/layout/DashboardLayout'
import { gisService } from '@/services/gis.service'
import { GISProjectMarker, GISFilter } from '@/types/gis'

// Dynamic import wajib untuk Leaflet — library ini browser-only, tidak kompatibel SSR
const GISMap = dynamic(() => import('@/components/gis/GISMap'), {
  ssr: false,
  loading: () => (
    <div className="w-full h-full flex items-center justify-center bg-gray-50 rounded-lg" style={{ minHeight: '480px' }}>
      <div className="text-center">
        <MapPin className="w-8 h-8 text-gray-400 mx-auto mb-2 animate-pulse" />
        <p className="text-sm text-gray-500">Memuat peta...</p>
      </div>
    </div>
  ),
})

const HEALTH_BADGE: Record<string, { label: string; className: string }> = {
  GREEN:    { label: 'Hijau',    className: 'bg-green-100 text-green-800' },
  YELLOW:   { label: 'Kuning',   className: 'bg-yellow-100 text-yellow-800' },
  RED:      { label: 'Merah',    className: 'bg-red-100 text-red-800' },
  CRITICAL: { label: 'Kritis',   className: 'bg-purple-100 text-purple-800' },
  UNSCORED: { label: 'Unscored', className: 'bg-gray-100 text-gray-600' },
}

const STATUS_OPTIONS = [
  { value: '', label: 'Semua Status' },
  { value: 'ACTIVE', label: 'Aktif' },
  { value: 'DRAFT', label: 'Draft' },
  { value: 'COMPLETED', label: 'Selesai' },
  { value: 'ON_HOLD', label: 'Ditunda' },
  { value: 'CANCELLED', label: 'Dibatalkan' },
]

const HEALTH_OPTIONS = [
  { value: '', label: 'Semua Health' },
  { value: 'GREEN', label: 'Hijau' },
  { value: 'YELLOW', label: 'Kuning' },
  { value: 'RED', label: 'Merah' },
  { value: 'CRITICAL', label: 'Kritis' },
  { value: 'UNSCORED', label: 'Belum Dinilai' },
]

function formatCurrency(val: number): string {
  if (val >= 1_000_000_000_000) return `Rp ${(val / 1_000_000_000_000).toFixed(1)}T`
  if (val >= 1_000_000_000) return `Rp ${(val / 1_000_000_000).toFixed(1)}M`
  if (val >= 1_000_000) return `Rp ${(val / 1_000_000).toFixed(1)}jt`
  return `Rp ${val.toLocaleString('id-ID')}`
}

export default function GISPage() {
  const [filter, setFilter] = useState<GISFilter>({})
  const [selectedProject, setSelectedProject] = useState<GISProjectMarker | null>(null)

  const { data: allMarkers = [], isLoading: loadingMarkers, refetch, isFetching } = useQuery({
    queryKey: ['gis-projects', filter],
    queryFn: () => gisService.getProjects(filter),
  })

  const { data: summary, isLoading: loadingSummary } = useQuery({
    queryKey: ['gis-summary'],
    queryFn: () => gisService.getSummary(),
  })

  // Hitung mapped (punya koordinat) untuk tampilan
  const mappedMarkers = useMemo(
    () => allMarkers.filter((m) => m.latitude != null && m.longitude != null),
    [allMarkers],
  )

  return (
    <DashboardLayout title="GIS Map">
      <div className="flex flex-col h-full gap-4 overflow-x-hidden">
      {/* Header */}
      <div className="flex items-center justify-between flex-shrink-0">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">GIS Map</h1>
          <p className="text-sm text-gray-500 mt-0.5">Sebaran geografis proyek infrastruktur</p>
        </div>
        <button
          onClick={() => refetch()}
          disabled={isFetching}
          className="flex items-center gap-2 px-3 py-2 text-sm text-gray-600 bg-white border border-gray-200 rounded-lg hover:bg-gray-50 disabled:opacity-50 transition-colors"
          aria-label="Refresh data peta"
        >
          <RefreshCw className={`w-4 h-4 ${isFetching ? 'animate-spin' : ''}`} />
          Muat Ulang
        </button>
      </div>

      {/* Summary Cards */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-3 flex-shrink-0">
        <SummaryCard
          icon={<MapPin className="w-5 h-5 text-blue-600" />}
          label="Total Proyek"
          value={loadingSummary ? '—' : String(summary?.total_projects ?? 0)}
          sub={loadingSummary ? '' : `${summary?.mapped_projects ?? 0} terpetakan`}
          iconBg="bg-blue-50"
        />
        <SummaryCard
          icon={<CheckCircle className="w-5 h-5 text-green-600" />}
          label="Health Hijau"
          value={loadingSummary ? '—' : String(summary?.health_green ?? 0)}
          sub="Proyek sehat"
          iconBg="bg-green-50"
        />
        <SummaryCard
          icon={<AlertTriangle className="w-5 h-5 text-red-600" />}
          label="Merah / Kritis"
          value={loadingSummary ? '—' : String((summary?.health_red ?? 0) + (summary?.health_critical ?? 0))}
          sub="Butuh perhatian"
          iconBg="bg-red-50"
        />
        <SummaryCard
          icon={<BarChart2 className="w-5 h-5 text-purple-600" />}
          label="Rata-rata Progres"
          value={loadingSummary ? '—' : `${(summary?.avg_progress_pct ?? 0).toFixed(1)}%`}
          sub="Fisik rata-rata"
          iconBg="bg-purple-50"
        />
      </div>

      {/* Filter Row */}
      <div className="flex flex-wrap gap-2 flex-shrink-0">
        <select
          value={filter.status ?? ''}
          onChange={(e) => setFilter((f) => ({ ...f, status: e.target.value || undefined }))}
          className="text-sm border border-gray-200 rounded-lg px-3 py-1.5 bg-white text-gray-700 focus:outline-none focus:ring-2 focus:ring-blue-500"
          aria-label="Filter status proyek"
        >
          {STATUS_OPTIONS.map((o) => (
            <option key={o.value} value={o.value}>{o.label}</option>
          ))}
        </select>
        <select
          value={filter.health_class ?? ''}
          onChange={(e) => setFilter((f) => ({ ...f, health_class: e.target.value || undefined }))}
          className="text-sm border border-gray-200 rounded-lg px-3 py-1.5 bg-white text-gray-700 focus:outline-none focus:ring-2 focus:ring-blue-500"
          aria-label="Filter health class proyek"
        >
          {HEALTH_OPTIONS.map((o) => (
            <option key={o.value} value={o.value}>{o.label}</option>
          ))}
        </select>
        <input
          type="text"
          placeholder="Filter provinsi..."
          value={filter.province ?? ''}
          onChange={(e) => setFilter((f) => ({ ...f, province: e.target.value || undefined }))}
          className="w-full sm:w-auto text-sm border border-gray-200 rounded-lg px-3 py-1.5 bg-white text-gray-700 focus:outline-none focus:ring-2 focus:ring-blue-500"
          aria-label="Filter provinsi"
        />
        {(filter.status || filter.health_class || filter.province) && (
          <button
            onClick={() => setFilter({})}
            className="text-sm text-blue-600 hover:underline px-2"
          >
            Reset filter
          </button>
        )}
        <span className="ml-auto text-xs text-gray-400 self-center">
          {loadingMarkers ? 'Memuat...' : `${allMarkers.length} proyek (${mappedMarkers.length} di peta)`}
        </span>
      </div>

      {/* Main Area: Map + Sidebar */}
      <div className="flex flex-col lg:flex-row gap-4 flex-1 min-h-0">
        {/* Map — container terkontrol: tinggi tetap agar Leaflet tidak melar */}
        <div className="flex-1 min-w-0 h-[420px] lg:h-[calc(100vh-280px)] min-h-[360px] rounded-xl overflow-hidden border border-gray-200 shadow-sm bg-white">
          {loadingMarkers ? (
            <div className="w-full h-full flex items-center justify-center">
              <div className="text-center">
                <MapPin className="w-8 h-8 text-gray-300 mx-auto mb-2 animate-pulse" />
                <p className="text-sm text-gray-400">Memuat data proyek...</p>
              </div>
            </div>
          ) : (
            <GISMap markers={allMarkers} onMarkerClick={setSelectedProject} />
          )}
        </div>

        {/* Sidebar: daftar proyek / detail proyek */}
        <div className="w-full lg:w-72 flex-shrink-0 flex flex-col gap-3 overflow-y-auto max-h-96 lg:max-h-none">
          {selectedProject ? (
            <ProjectDetailPanel
              project={selectedProject}
              onClose={() => setSelectedProject(null)}
            />
          ) : (
            <ProjectListPanel
              markers={allMarkers}
              onSelect={setSelectedProject}
              loading={loadingMarkers}
            />
          )}
        </div>
      </div>

      {/* Legend — flex-wrap agar badge tidak melebar di viewport kecil */}
      <div className="flex flex-wrap items-center gap-x-4 gap-y-1.5 flex-shrink-0 bg-white border border-gray-100 rounded-lg px-4 py-2">
        <span className="text-xs font-medium text-gray-500">Legenda:</span>
        {Object.entries(HEALTH_BADGE).map(([key, { label, className }]) => (
          <span key={key} className={`text-xs font-medium px-2 py-0.5 rounded-full ${className}`}>
            {label}
          </span>
        ))}
        <span className="ml-auto text-xs text-gray-400">Tile: © OpenStreetMap</span>
      </div>
      </div>
    </DashboardLayout>
  )
}

// ─── Sub-components ───────────────────────────────────────────────────────────

function SummaryCard({
  icon, label, value, sub, iconBg,
}: {
  icon: React.ReactNode
  label: string
  value: string
  sub: string
  iconBg: string
}) {
  return (
    <div className="bg-white border border-gray-100 rounded-xl p-4 flex items-center gap-3 shadow-sm min-w-0">
      <div className={`p-2 rounded-lg shrink-0 ${iconBg}`}>{icon}</div>
      <div className="min-w-0">
        <p className="text-xs text-gray-500">{label}</p>
        <p className="text-xl font-bold text-gray-900 truncate">{value}</p>
        <p className="text-xs text-gray-400 truncate">{sub}</p>
      </div>
    </div>
  )
}

function ProjectListPanel({
  markers, onSelect, loading,
}: {
  markers: GISProjectMarker[]
  onSelect: (p: GISProjectMarker) => void
  loading: boolean
}) {
  const sorted = useMemo(
    () => [...markers].sort((a, b) => {
      // Prioritaskan CRITICAL dan RED
      const order: Record<string, number> = { CRITICAL: 0, RED: 1, YELLOW: 2, GREEN: 3, UNSCORED: 4 }
      return (order[a.health_class] ?? 9) - (order[b.health_class] ?? 9)
    }),
    [markers],
  )

  if (loading) return (
    <div className="bg-white rounded-xl border border-gray-100 p-4 text-sm text-gray-400 text-center">
      Memuat daftar proyek...
    </div>
  )

  return (
    <div className="bg-white rounded-xl border border-gray-100 overflow-hidden shadow-sm">
      <div className="px-4 py-3 border-b border-gray-100">
        <h2 className="text-sm font-semibold text-gray-700">Daftar Proyek</h2>
        <p className="text-xs text-gray-400">{sorted.length} proyek · klik untuk detail</p>
      </div>
      <ul className="divide-y divide-gray-50 max-h-[520px] overflow-y-auto">
        {sorted.map((p) => {
          const badge = HEALTH_BADGE[p.health_class] ?? HEALTH_BADGE.UNSCORED
          return (
            <li key={p.project_id}>
              <button
                onClick={() => onSelect(p)}
                className="w-full text-left px-4 py-3 hover:bg-gray-50 transition-colors"
              >
                <div className="flex items-start justify-between gap-2">
                  <div className="min-w-0">
                    <p className="text-xs font-semibold text-gray-800 truncate">{p.project_name}</p>
                    <p className="text-xs text-gray-400">{p.project_code}</p>
                  </div>
                  <span className={`text-xs font-medium px-2 py-0.5 rounded-full flex-shrink-0 ${badge.className}`}>
                    {badge.label}
                  </span>
                </div>
                <div className="flex items-center gap-3 mt-1">
                  <span className="text-xs text-gray-500">{p.progress_pct.toFixed(1)}%</span>
                  {p.latitude != null ? (
                    <MapPin className="w-3 h-3 text-blue-400" />
                  ) : (
                    <span className="text-xs text-gray-300">No GPS</span>
                  )}
                  {p.open_risks > 0 && (
                    <span className="text-xs text-red-500">{p.open_risks} risiko</span>
                  )}
                </div>
              </button>
            </li>
          )
        })}
        {sorted.length === 0 && (
          <li className="px-4 py-6 text-sm text-gray-400 text-center">
            Tidak ada proyek yang sesuai filter
          </li>
        )}
      </ul>
    </div>
  )
}

function ProjectDetailPanel({
  project, onClose,
}: {
  project: GISProjectMarker
  onClose: () => void
}) {
  const badge = HEALTH_BADGE[project.health_class] ?? HEALTH_BADGE.UNSCORED
  return (
    <div className="bg-white rounded-xl border border-gray-100 overflow-hidden shadow-sm">
      <div className="px-4 py-3 border-b border-gray-100 flex items-center justify-between">
        <h2 className="text-sm font-semibold text-gray-700">Detail Proyek</h2>
        <button
          onClick={onClose}
          className="text-xs text-blue-600 hover:underline"
          aria-label="Tutup detail proyek"
        >
          ← Kembali
        </button>
      </div>
      <div className="p-4 space-y-3">
        <div>
          <p className="text-xs text-gray-400">{project.project_code}</p>
          <p className="text-sm font-bold text-gray-900">{project.project_name}</p>
        </div>

        <span className={`inline-block text-xs font-medium px-2 py-0.5 rounded-full ${badge.className}`}>
          {badge.label}
        </span>

        <div className="space-y-2 text-sm">
          <DetailRow label="Status" value={project.status} />
          <DetailRow label="Progres Fisik" value={`${project.progress_pct.toFixed(1)}%`} />
          <DetailRow label="Anggaran" value={formatCurrency(project.budget_total)} />
          <DetailRow label="Priority Score" value={project.priority_score.toFixed(1)} />
          <DetailRow label="Risiko Terbuka" value={String(project.open_risks)} />
          <DetailRow label="Isu Terbuka" value={String(project.open_issues)} />
        </div>

        <div className="border-t border-gray-100 pt-3 space-y-1">
          <p className="text-xs font-medium text-gray-500">Lokasi</p>
          {project.location_name && (
            <p className="text-xs text-gray-700 font-medium">{project.location_name}</p>
          )}
          {(project.city || project.province) && (
            <p className="text-xs text-gray-500">
              {[project.city, project.province].filter(Boolean).join(', ')}
            </p>
          )}
          {project.region_name && (
            <p className="text-xs text-gray-400">Wilayah: {project.region_name}</p>
          )}
          {project.latitude != null ? (
            <p className="text-xs text-gray-400 font-mono">
              {project.latitude.toFixed(4)}, {project.longitude!.toFixed(4)}
            </p>
          ) : (
            <p className="text-xs text-gray-300 italic">Koordinat belum diisi</p>
          )}
        </div>
      </div>
    </div>
  )
}

function DetailRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex justify-between">
      <span className="text-gray-500">{label}</span>
      <span className="font-medium text-gray-800">{value}</span>
    </div>
  )
}
