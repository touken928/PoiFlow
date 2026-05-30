import {useMemo, useState, type CSSProperties} from 'react';
import {Tooltip, tokens, Text, Input, Table, TableBody, TableCell, TableHeader, TableHeaderCell, TableRow} from '@fluentui/react-components';

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
const NUMERIC_COLS = new Set(['lng', 'lat']);

export function poiRowId(r: TablePoi) {
    return r.uid || `${r.lng},${r.lat}`;
}

function colLabel(f: string, labels: Record<string, string>) {
    if (f === 'lng') return '经度';
    if (f === 'lat') return '纬度';
    return labels[f] || f;
}

export function cellValue(r: TablePoi, f: string) {
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

function compareRows(a: TablePoi, b: TablePoi, field: string, dir: 'asc' | 'desc') {
    const va = cellValue(a, field);
    const vb = cellValue(b, field);
    let cmp = 0;
    if (NUMERIC_COLS.has(field)) {
        const na = parseFloat(va), nb = parseFloat(vb);
        cmp = (Number.isFinite(na) ? na : 0) - (Number.isFinite(nb) ? nb : 0);
    } else {
        cmp = va.localeCompare(vb, 'zh-CN');
    }
    return dir === 'asc' ? cmp : -cmp;
}

function validCoord(r: TablePoi) {
    return Number.isFinite(r.lat) && Number.isFinite(r.lng) && !(r.lat === 0 && r.lng === 0);
}

interface PoiTableProps {
    records: TablePoi[];
    columns: string[];
    fieldLabels: Record<string, string>;
    limit?: number;
    highlightUid?: string | null;
    onLocate?: (record: TablePoi) => void;
}

export function PoiTable({records, columns, fieldLabels, limit = 1000, highlightUid, onLocate}: PoiTableProps) {
    const [query, setQuery] = useState('');
    const [sortField, setSortField] = useState<string | null>(null);
    const [sortDir, setSortDir] = useState<'asc' | 'desc'>('asc');

    const filtered = useMemo(() => {
        const q = query.trim().toLowerCase();
        let rows = q
            ? records.filter(r => columns.some(f => cellValue(r, f).toLowerCase().includes(q)))
            : records;
        if (sortField) {
            rows = [...rows].sort((a, b) => compareRows(a, b, sortField, sortDir));
        }
        return rows;
    }, [records, columns, query, sortField, sortDir]);

    const shown = filtered.slice(0, limit);
    const tableWidth = columns.reduce((s, f) => s + (COL_WIDTH[f] ?? 100), 0);

    const toggleSort = (field: string) => {
        if (sortField === field) {
            setSortDir(d => d === 'asc' ? 'desc' : 'asc');
        } else {
            setSortField(field);
            setSortDir('asc');
        }
    };

    const sortMark = (field: string) => {
        if (sortField !== field) return '';
        return sortDir === 'asc' ? ' ↑' : ' ↓';
    };

    return (
        <div style={{display: 'flex', flexDirection: 'column', flex: 1, minHeight: 0}}>
            <div style={{
                flexShrink: 0, padding: '12px 0 8px',
                display: 'flex', alignItems: 'center', gap: 12, flexWrap: 'wrap',
            }}>
                <Input
                    placeholder="搜索表格内容…"
                    value={query}
                    onChange={(_e, d) => setQuery(d.value)}
                    style={{flex: 1, minWidth: 180, maxWidth: 320}}
                    size="small"
                />
                <Text size={200} weight="semibold">
                    {query.trim() ? `匹配 ${filtered.length} / ${records.length} 条` : `共 ${records.length} 条`}
                </Text>
                {records.length > limit && (
                    <Text size={200} style={{color: tokens.colorNeutralForeground3}}>
                        显示前 {limit} 条
                    </Text>
                )}
                {onLocate && (
                    <Text size={100} style={{color: tokens.colorNeutralForeground3}}>
                        点击行在地图中定位
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
                                            cursor: 'pointer',
                                            userSelect: 'none',
                                        }}
                                        onClick={() => toggleSort(f)}
                                        title="点击排序"
                                    >
                                        {colLabel(f, fieldLabels)}{sortMark(f)}
                                    </TableHeaderCell>
                                );
                            })}
                        </TableRow>
                    </TableHeader>
                    <TableBody>
                        {shown.length === 0 && (
                            <TableRow>
                                <TableCell colSpan={columns.length + 1} style={{padding: '24px', textAlign: 'center'}}>
                                    <Text style={{color: tokens.colorNeutralForeground3}}>无匹配结果</Text>
                                </TableCell>
                            </TableRow>
                        )}
                        {shown.map((r, i) => {
                            const id = poiRowId(r);
                            const selected = highlightUid === id;
                            const canLocate = onLocate && validCoord(r);
                            return (
                                <TableRow
                                    key={id}
                                    onClick={canLocate ? () => onLocate(r) : undefined}
                                    style={{
                                        cursor: canLocate ? 'pointer' : 'default',
                                        background: selected
                                            ? tokens.colorBrandBackground2
                                            : i % 2 === 1 ? tokens.colorNeutralBackground2 : tokens.colorNeutralBackground1,
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
                            );
                        })}
                    </TableBody>
                </Table>
            </div>
        </div>
    );
}
