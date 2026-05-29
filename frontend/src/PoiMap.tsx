import {useEffect, useRef} from 'react';
import L from 'leaflet';
import 'leaflet/dist/leaflet.css';
import 'leaflet.markercluster';
import 'leaflet.markercluster/dist/MarkerCluster.css';
import 'leaflet.markercluster/dist/MarkerCluster.Default.css';
import {tokens} from '@fluentui/react-components';

export interface MapPoi {
    name: string;
    lng: number;
    lat: number;
    address: string;
    province: string;
    city: string;
    area: string;
}

interface PoiMapProps {
    records: MapPoi[];
    active: boolean;
}

function validCoord(lat: number, lng: number) {
    return Number.isFinite(lat) && Number.isFinite(lng) && lat >= -90 && lat <= 90 && lng >= -180 && lng <= 180 && !(lat === 0 && lng === 0);
}

function escapeHtml(s: string) {
    return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
}

export function PoiMap({records, active}: PoiMapProps) {
    const containerRef = useRef<HTMLDivElement>(null);
    const mapRef = useRef<L.Map | null>(null);
    const clusterRef = useRef<L.MarkerClusterGroup | null>(null);

    useEffect(() => {
        if (!containerRef.current || mapRef.current) return;
        const map = L.map(containerRef.current, {zoomControl: true}).setView([35.86, 104.19], 4);
        L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
            attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>',
            maxZoom: 19,
        }).addTo(map);
        const cluster = L.markerClusterGroup({
            maxClusterRadius: 50,
            disableClusteringAtZoom: 16,
        });
        map.addLayer(cluster);
        mapRef.current = map;
        clusterRef.current = cluster;
        return () => {
            map.remove();
            mapRef.current = null;
            clusterRef.current = null;
        };
    }, []);

    useEffect(() => {
        if (!active || !mapRef.current) return;
        const t = window.setTimeout(() => mapRef.current?.invalidateSize(), 0);
        return () => window.clearTimeout(t);
    }, [active]);

    useEffect(() => {
        const map = mapRef.current;
        const cluster = clusterRef.current;
        if (!map || !cluster) return;

        cluster.clearLayers();
        const points = records.filter(r => validCoord(r.lat, r.lng));
        for (const r of points) {
            const marker = L.circleMarker([r.lat, r.lng], {
                radius: 6,
                color: tokens.colorBrandStroke1 as string,
                fillColor: tokens.colorBrandBackground as string,
                fillOpacity: 0.85,
                weight: 1,
            });
            const popup = [
                r.name ? `<b>${escapeHtml(r.name)}</b>` : '',
                r.address ? escapeHtml(r.address) : '',
                `${r.lng.toFixed(6)}, ${r.lat.toFixed(6)}`,
                [r.province, r.city, r.area].filter(Boolean).join(' '),
            ].filter(Boolean).join('<br/>');
            marker.bindPopup(popup);
            cluster.addLayer(marker);
        }
        if (points.length > 0) {
            map.fitBounds(cluster.getBounds().pad(0.08));
        }
    }, [records]);

    return (
        <div style={{position: 'relative', width: '100%', height: '100%', minHeight: 0}}>
            <div ref={containerRef} style={{width: '100%', height: '100%'}}/>
            {records.length > 0 && (
                <div style={{
                    position: 'absolute', top: 10, right: 10, zIndex: 1000,
                    padding: '6px 10px', borderRadius: 6, fontSize: 12,
                    background: 'rgba(255,255,255,0.92)',
                    border: `1px solid ${tokens.colorNeutralStroke2}`,
                    boxShadow: '0 1px 4px rgba(0,0,0,0.12)',
                }}>
                    共 {records.length} 条
                </div>
            )}
        </div>
    );
}
