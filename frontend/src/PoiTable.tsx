import type {CSSProperties} from 'react';
import {Tooltip, tokens, Text, Table, TableBody, TableCell, TableHeader, TableHeaderCell, TableRow} from '@fluentui/react-components';

export interface TablePoi {
    name: string;
    lng: number;
    lat: number;
    address: string;
    telephone: string;
    province: string;
    city: string;
    area: string;
    uid: string;
    query: string;
    type: string;
    taskName: string;
    target: string;
}

const COL_WIDTH: Record<string, number> = {
    lng: 108, lat: 108, name: 150, address: 240, telephone: 120,
    province: 72, city: 72, area: 72, uid: 128, query: 100, type: 88,
    taskName: 110, target: 110,
};
const COL_MONO = new Set(['lng', 'lat', 'uid']);
const COL_CLAMP2 = new Set(['name', 'address']);

function colLabel(f: string, labels: Record<string, string>) {
    if (f === 'lng') return '经度';
    if (f === 'lat') return '纬度';
    return labels[f] || f;
}

function cellValue(r: TablePoi, f: string) {
    if (f === 'lng' || f === 'lat') {
        const n = r[f];
        return Number.isFinite(n) ? n.toFixed(6) : '';
    }
    const v: Record<string, string> = {
        name: r.name, address: r.address, telephone: r.telephone,
        province: r.province, city: r.city, area: r.area, uid: r.uid,
        query: r.query, type: r.type, taskName: r.taskName, target: r.target,
    };
    return v[f] ?? '';
}

function cellStyle(field: string, width: number): CSSProperties {
    const base: CSSProperties = {
        maxWidth: width,
        width,
        minWidth: width,
        overflow: 'hidden',
        fontSize: COL_MONO.has(field) ? 12 : 13,
        lineHeight: 1.45,
        color: tokens.colorNeutralForeground1,
    };
    if (COL_CLAMP2.has(field)) {
        return {
            ...base,
            display: '-webkit-box',
            WebkitLineClamp: 2,
            WebkitBoxOrient: 'vertical',
            whiteSpace: 'normal',
            wordBreak: 'break-all',
        };
    }
    return {
        ...base,
        textOverflow: 'ellipsis',
        whiteSpace: 'nowrap',
    };
}

function TableCellText({text, field}: {text: string; field: string}) {
    const width = COL_WIDTH[field] ?? 100;
    const display = text || '—';
    const empty = !text;
    const inner = (
        <div style={{
            ...cellStyle(field, width),
            color: empty ? tokens.colorNeutralForeground3 : tokens.colorNeutralForeground1,
            fontFamily: COL_MONO.has(field) ? 'ui-monospace, SFMono-Regular, Menlo, monospace' : undefined,
        }}>
            {display}
        </div>
    );
    if (empty || display.length < 18) return inner;
    return (
        <Tooltip content={display} relationship="description" withArrow positioning="above">
            <div style={{minWidth: 0, maxWidth: width, cursor: 'default'}}>{inner}</div>
        </Tooltip>
    );
}

interface PoiTableProps {
    records: TablePoi[];
    columns: string[];
    fieldLabels: Record<string, string>;
    limit?: number;
}

export function PoiTable({records, columns, fieldLabels, limit = 1000}: PoiTableProps) {
    const shown = records.slice(0, limit);
    const tableWidth = columns.reduce((s, f) => s + (COL_WIDTH[f] ?? 100), 0);

    return (
        <div style={{display: 'flex', flexDirection: 'column', flex: 1, minHeight: 0}}>
            <div style={{
                flexShrink: 0, padding: '12px 0 8px',
                display: 'flex', alignItems: 'center', gap: 12,
            }}>
                <Text size={200} weight="semibold">共 {records.length} 条</Text>
                {records.length > limit && (
                    <Text size={200} style={{color: tokens.colorNeutralForeground3}}>
                        表格显示前 {limit} 条
                    </Text>
                )}
            </div>
            <div style={{
                flex: 1, minHeight: 0, overflow: 'auto',
                border: `1px solid ${tokens.colorNeutralStroke2}`,
                borderRadius: '8px',
                background: tokens.colorNeutralBackground1,
            }}>
                <Table
                    size="small"
                    style={{tableLayout: 'fixed', width: tableWidth, minWidth: '100%'}}
                    aria-label="POI 数据表"
                >
                    <TableHeader
                        style={{
                            position: 'sticky', top: 0, zIndex: 2,
                            background: tokens.colorNeutralBackground2,
                            boxShadow: `0 1px 0 ${tokens.colorNeutralStroke2}`,
                        }}
                    >
                        <TableRow>
                            <TableHeaderCell style={{
                                width: 44, minWidth: 44, maxWidth: 44,
                                color: tokens.colorNeutralForeground3,
                                fontWeight: 600, fontSize: 12,
                            }}>#</TableHeaderCell>
                            {columns.map(f => {
                                const w = COL_WIDTH[f] ?? 100;
                                return (
                                    <TableHeaderCell
                                        key={f}
                                        style={{
                                            width: w, minWidth: w, maxWidth: w,
                                            fontWeight: 600, fontSize: 12,
                                            color: tokens.colorNeutralForeground2,
                                        }}
                                    >
                                        {colLabel(f, fieldLabels)}
                                    </TableHeaderCell>
                                );
                            })}
                        </TableRow>
                    </TableHeader>
                    <TableBody>
                        {shown.map((r, i) => (
                            <TableRow
                                key={r.uid || `${i}`}
                                style={{
                                    background: i % 2 === 1 ? tokens.colorNeutralBackground2 : tokens.colorNeutralBackground1,
                                }}
                            >
                                <TableCell style={{
                                    width: 44, minWidth: 44, maxWidth: 44,
                                    color: tokens.colorNeutralForeground3,
                                    fontSize: 12, verticalAlign: 'middle',
                                    padding: '10px 8px',
                                }}>
                                    {i + 1}
                                </TableCell>
                                {columns.map(f => (
                                    <TableCell
                                        key={f}
                                        style={{
                                            width: COL_WIDTH[f] ?? 100,
                                            verticalAlign: 'middle',
                                            padding: '10px 12px',
                                        }}
                                    >
                                        <TableCellText text={cellValue(r, f)} field={f}/>
                                    </TableCell>
                                ))}
                            </TableRow>
                        ))}
                    </TableBody>
                </Table>
            </div>
        </div>
    );
}
