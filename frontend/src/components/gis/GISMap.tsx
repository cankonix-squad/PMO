'use client'

// Leaflet CSS wajib dimuat agar .leaflet-container memiliki position: relative
// dan ukuran terkontrol. Tanpa ini, map container melar setinggi isi tile
// (mis. 3704px) dan merusak layout halaman.
import 'leaflet/dist/leaflet.css'

import { useEffect, useRef } from 'react'
import type L from 'leaflet'
import type { Map as LeafletMap } from 'leaflet'
import { GISProjectMarker } from '@/types/gis'

type LeafletType = typeof L

// Warna marker berdasarkan health class
const HEALTH_COLORS: Record<string, string> = {
  GREEN: '#16a34a',
  YELLOW: '#ca8a04',
  RED: '#dc2626',
  CRITICAL: '#7c3aed',
  UNSCORED: '#6b7280',
}

const STATUS_LABEL: Record<string, string> = {
  ACTIVE: 'Aktif',
  DRAFT: 'Draft',
  COMPLETED: 'Selesai',
  ON_HOLD: 'Ditunda',
  CANCELLED: 'Dibatalkan',
}

interface GISMapProps {
  markers: GISProjectMarker[]
  onMarkerClick?: (project: GISProjectMarker) => void
}

export default function GISMap({ markers, onMarkerClick }: GISMapProps) {
  const mapRef = useRef<LeafletMap | null>(null)
  const containerRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (typeof window === 'undefined') return
    if (!containerRef.current) return
    if (mapRef.current) return // already initialized

    // Lazy-load leaflet only on client
    import('leaflet').then((L: LeafletType) => {
      // Fix default icon paths yang rusak di Next.js bundler
      delete (L.Icon.Default.prototype as unknown as Record<string, unknown>)._getIconUrl
      L.Icon.Default.mergeOptions({
        iconRetinaUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-icon-2x.png',
        iconUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-icon.png',
        shadowUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-shadow.png',
      })

      const map = L.map(containerRef.current!, {
        center: [-2.5, 118.0], // tengah Indonesia
        zoom: 5,
        zoomControl: true,
      })
      mapRef.current = map

      // Tile layer OpenStreetMap — tanpa API key
      L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
        attribution: '© <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>',
        maxZoom: 18,
      }).addTo(map)

      // Render marker
      renderMarkers(L, map, markers, onMarkerClick)

      // Pastikan ukuran map sinkron dengan container setelah inisialisasi
      // dan saat container berubah ukuran (mis. mobile ↔ desktop, drawer open).
      // Tanpa ini, Leaflet SVG pane bisa lebih lebar dari container → horizontal
      // overflow pada viewport kecil.
      requestAnimationFrame(() => map.invalidateSize())
    })

    return () => {
      if (mapRef.current) {
        mapRef.current.remove()
        mapRef.current = null
      }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  // Observasi perubahan ukuran container → invalidateSize agar Leaflet tidak melar
  useEffect(() => {
    const container = containerRef.current
    if (!container || typeof ResizeObserver === 'undefined') return
    const ro = new ResizeObserver(() => {
      if (mapRef.current) mapRef.current.invalidateSize()
    })
    ro.observe(container)
    return () => ro.disconnect()
  }, [])

  // Update marker saat data berubah
  useEffect(() => {
    if (!mapRef.current) return
    import('leaflet').then((L: LeafletType) => {
      if (!mapRef.current) return
      // Hapus semua layer marker lama (bukan tile layer)
      mapRef.current.eachLayer((layer) => {
        if ((layer as { _latlng?: unknown })._latlng !== undefined) {
          mapRef.current!.removeLayer(layer)
        }
      })
      renderMarkers(L, mapRef.current, markers, onMarkerClick)
    })
  }, [markers, onMarkerClick])

  return (
    <div
      ref={containerRef}
      className="w-full h-full rounded-lg"
      style={{ minHeight: '480px' }}
      aria-label="Peta sebaran proyek"
      role="img"
    />
  )
}

function renderMarkers(
  L: LeafletType,
  map: LeafletMap,
  markers: GISProjectMarker[],
  onMarkerClick?: (project: GISProjectMarker) => void,
) {
  markers.forEach((project) => {
    if (project.latitude == null || project.longitude == null) return

    const color = HEALTH_COLORS[project.health_class] ?? HEALTH_COLORS.UNSCORED

    // Custom circle marker dengan warna health
    const circleMarker = L.circleMarker([project.latitude, project.longitude], {
      radius: 10,
      fillColor: color,
      color: '#fff',
      weight: 2,
      opacity: 1,
      fillOpacity: 0.9,
    })

    const popupContent = `
      <div style="min-width:220px;font-family:sans-serif;font-size:13px;">
        <div style="font-weight:700;font-size:14px;margin-bottom:4px;">${project.project_name}</div>
        <div style="color:#6b7280;margin-bottom:8px;">${project.project_code}</div>
        <table style="width:100%;border-collapse:collapse;">
          <tr><td style="color:#6b7280;padding:2px 0;">Status</td><td style="text-align:right;font-weight:600;">${STATUS_LABEL[project.status] ?? project.status}</td></tr>
          <tr><td style="color:#6b7280;padding:2px 0;">Health</td><td style="text-align:right;"><span style="background:${color};color:#fff;padding:1px 8px;border-radius:9999px;font-size:11px;font-weight:600;">${project.health_class}</span></td></tr>
          <tr><td style="color:#6b7280;padding:2px 0;">Progres</td><td style="text-align:right;font-weight:600;">${project.progress_pct.toFixed(1)}%</td></tr>
          <tr><td style="color:#6b7280;padding:2px 0;">Risiko Terbuka</td><td style="text-align:right;font-weight:600;">${project.open_risks}</td></tr>
          <tr><td style="color:#6b7280;padding:2px 0;">Isu Terbuka</td><td style="text-align:right;font-weight:600;">${project.open_issues}</td></tr>
          <tr><td style="color:#6b7280;padding:2px 0;">Lokasi</td><td style="text-align:right;">${[project.city, project.province].filter(Boolean).join(', ') || '-'}</td></tr>
        </table>
      </div>
    `

    circleMarker.bindPopup(popupContent, { maxWidth: 280 })

    if (onMarkerClick) {
      circleMarker.on('click', () => onMarkerClick(project))
    }

    circleMarker.addTo(map)
  })
}
