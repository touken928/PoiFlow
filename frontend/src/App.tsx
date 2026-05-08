import {useState, useEffect, useCallback, useRef} from 'react';
import {
    Button, Text, Input, Dialog, DialogTrigger, DialogSurface,
    DialogTitle, DialogBody, DialogActions, DialogContent, Dropdown, Option,
    Badge, tokens, FluentProvider, webLightTheme, ProgressBar, ToolbarButton,
} from '@fluentui/react-components';
import {
    Settings24Regular, Add24Regular, Play24Regular, Pause24Regular,
    Delete24Regular, ArrowDownload24Regular,
} from '@fluentui/react-icons';

import {
    GetProvinces, GetCities, GetCounties, CreateTask, GetTasks,
    CancelTask, PauseTask, ResumeTask, DeleteTask, GetAKItems,
    ResetAKPool, AddAK, RemoveAK, VerifyAK, GetTaskLogs, ExportTaskDialog, ExpandCount,
} from '../wailsjs/go/main/App';
import {EventsOn} from '../wailsjs/runtime/runtime';

const STATUS_LABELS = ['等待中','执行中','已暂停','已完成','失败','已取消'];
const GRAN = ['省级','市级','区县级'];

interface Task {
    id:string; name:string; query:string; type:string; exportPath:string;
    areaGranularity:number; queryGranularity:number;
    targets:{province:string;city:string;name:string}[];
    status:number; progress:string; records:number; error:string;
    createdAt:string; updatedAt:string;
}
interface LogEntry{time:string;message:string;level:string}
interface AKItem{ak:string;used:number;failed:boolean;failMsg:string}

const st = (styles: Record<string,any>) => styles;

const css = st({
    root:{height:'100vh',display:'flex',background:tokens.colorNeutralBackground1},
    sidebar:{width:'300px',display:'flex',flexDirection:'column',borderRight:`1px solid ${tokens.colorNeutralStroke1}`,background:tokens.colorNeutralBackground2},
    brand:{padding:'20px 20px 12px',fontSize:'22px',fontWeight:'700',borderBottom:`1px solid ${tokens.colorNeutralStroke2}`,display:'flex',alignItems:'center',gap:'8px'},
    brandDot:{width:'10px',height:'10px',borderRadius:'50%',background:tokens.colorBrandForeground1,display:'inlineBlock'},
    listHeader:{display:'flex',alignItems:'center',justifyContent:'space-between',padding:'12px 16px',borderBottom:`1px solid ${tokens.colorNeutralStroke1}`},
    listBody:{flex:'1',overflowY:'auto',padding:'8px'},
    taskCard:{padding:'12px',marginBottom:'8px',borderRadius:'8px',border:`1px solid ${tokens.colorNeutralStroke2}`,background:tokens.colorNeutralBackground1,cursor:'pointer'},
    taskCardSel:{border:`2px solid ${tokens.colorBrandStroke1}`},
    taskName:{fontWeight:'600',marginBottom:'2px'},
    taskMeta:{fontSize:'11px',color:tokens.colorNeutralForeground3,marginBottom:'6px'},
    statusBar:{display:'flex',alignItems:'center',gap:'8px',marginTop:'4px'},
    footer:{padding:'10px 16px',borderTop:`1px solid ${tokens.colorNeutralStroke1}`,display:'flex',alignItems:'center',cursor:'pointer',color:tokens.colorNeutralForeground2},
    main:{flex:'1',display:'flex',flexDirection:'column',overflow:'hidden'},
    empty:{flex:'1',display:'flex',alignItems:'center',justifyContent:'center',color:tokens.colorNeutralForeground3,flexDirection:'column',gap:'8px'},
    logPanel:{flex:'1',overflowY:'auto',padding:'20px',fontFamily:'monospace',fontSize:'13px',lineHeight:'1.7'},
    form:{display:'flex',flexDirection:'column',gap:'14px',minWidth:'460px'},
    targets:{border:`1px solid ${tokens.colorNeutralStroke2}`,borderRadius:'8px',padding:'12px',maxHeight:'200px',overflowY:'auto'},
});

function App(){
    const logEnd=useRef<HTMLDivElement>(null);
    const [tasks,setTasks]=useState<Task[]>([]);
    const [akItems,setAkItems]=useState<AKItem[]>([]);
    const [sel,setSel]=useState<Task|null>(null);
    const [logs,setLogs]=useState<LogEntry[]>([]);
    const [msg,setMsg]=useState('');
    const [openNew,setOpenNew]=useState(false);
    const [openAk,setOpenAk]=useState(false);
    const [nName,setNName]=useState('');
    const [nQuery,setNQuery]=useState('');
    const [nType,setNType]=useState('');
    const [nExport,setNExport]=useState('');
    const [nAreaGran,setNAreaGran]=useState(0);
    const [nQueryGran,setNQueryGran]=useState(0);
    const [provinces,setProvinces]=useState<string[]>([]);
    const [cities,setCities]=useState<string[]>([]);
    const [counties,setCounties]=useState<string[]>([]);
    const [nProv,setNProv]=useState('');
    const [nCity,setNCity]=useState('');
    const [nTargets,setNTargets]=useState<{province:string;city:string;name:string}[]>([]);
    const [expandCount,setExpandCount]=useState(0);
    const [newAk,setNewAk]=useState('');
    const [verifying,setVerifying]=useState('');

    const loadAll=useCallback(()=>{GetTasks().then(setTasks).catch(()=>{});GetAKItems().then(setAkItems).catch(()=>{});},[]);

    useEffect(()=>{GetProvinces().then(setProvinces).catch(()=>{});loadAll();const c=[EventsOn('task:added',loadAll),EventsOn('task:updated',loadAll),EventsOn('task:completed',loadAll),EventsOn('task:failed',loadAll),EventsOn('task:deleted',loadAll)];return()=>{c.forEach(f=>f());};},[loadAll]);

    useEffect(()=>{if(sel){const u=tasks.find(t=>t.id===sel.id);if(u)setSel(u);}},[tasks]);

    useEffect(()=>{const iv=setInterval(()=>{if(sel&&(sel.status===1||sel.status===2)){GetTaskLogs(sel.id).then(setLogs).catch(()=>{});GetTasks().then(setTasks).catch(()=>{});}},2000);return()=>clearInterval(iv);},[sel]);

    useEffect(()=>{if(logEnd.current)logEnd.current.scrollIntoView({behavior:'smooth'});},[logs]);

    useEffect(()=>{if(nTargets.length>0)ExpandCount(nAreaGran,nQueryGran,nTargets).then(setExpandCount).catch(()=>setExpandCount(0));else setExpandCount(0);},[nTargets,nAreaGran,nQueryGran]);

    const onProvChange=(v:string)=>{setNProv(v);setNCity('');setCities([]);setCounties([]);if(v)GetCities(v).then(setCities).catch(()=>{});};
    const onCityChange=(v:string)=>{setNCity(v);setCounties([]);if(v&&nProv)GetCounties(nProv,v).then(setCounties).catch(()=>{});};
    const toggleTarget=(name:string)=>setNTargets(p=>{if(p.some(t=>t.name===name))return p.filter(t=>t.name!==name);if(nAreaGran===0)return[...p,{province:name,city:'',name}];if(nAreaGran===1)return[...p,{province:nProv,city:'',name}];return[...p,{province:nProv,city:nCity,name}];});

    const handleCreate=async()=>{if(!nName||!nQuery||nTargets.length===0){setMsg('请填写完整信息');return;}try{await CreateTask(nName,nQuery,nType,nExport,nAreaGran,nQueryGran,nTargets);setOpenNew(false);setMsg('');setNName('');setNQuery('');setNType('');setNExport('');setNTargets([]);setNProv('');setNCity('');setExpandCount(0);}catch(e:any){setMsg('创建失败: '+e);}};
    const selectTask=async(t:Task)=>{setSel(t);try{setLogs(await GetTaskLogs(t.id)||[]);}catch(e){}};
    const handleExport=async(id:string)=>{try{const r=await ExportTaskDialog(id);if(r)setMsg(r);}catch(e:any){setMsg('导出失败: '+e);}};
    const handlePause=async(id:string)=>{await PauseTask(id);loadAll();};
    const handleResume=async(id:string)=>{await ResumeTask(id);loadAll();};
    const handleCancel=async(id:string)=>{await CancelTask(id);loadAll();};
    const handleDelete=async(id:string)=>{await DeleteTask(id);loadAll();if(sel?.id===id){setSel(null);setLogs([]);}};
    const handleAddAk=async()=>{if(!newAk){setMsg('请输入AK');return;}const r=await AddAK(newAk);if(r)setMsg(r);else{setNewAk('');setMsg('AK已添加');}GetAKItems().then(setAkItems).catch(()=>{});};
    const handleRemoveAk=async(ak:string)=>{const r=await RemoveAK(ak);if(r)setMsg(r);GetAKItems().then(setAkItems).catch(()=>{});};
    const handleVerifyAk=async(ak:string)=>{setVerifying(ak);const r=await VerifyAK(ak);if(r)setMsg(ak.slice(0,8)+'... '+r);else setMsg(ak.slice(0,8)+'... 正常');setVerifying('');GetAKItems().then(setAkItems).catch(()=>{});};
    const handleVerifyAll=async()=>{for(const it of akItems){setVerifying(it.ak);const r=await VerifyAK(it.ak);if(r)setMsg(it.ak.slice(0,8)+'... '+r);};setVerifying('');GetAKItems().then(setAkItems).catch(()=>{});};

    const available=nAreaGran===0?provinces:nAreaGran===1?cities:counties;

    const badge=(st:number)=>{const m:Record<number,'success'|'warning'|'danger'|'important'>={0:'important',1:'warning',2:'warning',3:'success',4:'danger',5:'important'};return<Badge appearance="filled" color={m[st]||'important'} size="small">{STATUS_LABELS[st]||'?'}</Badge>;};

    const logLevelStyle=(lvl:string)=>({color:lvl==='error'?tokens.colorStatusDangerForeground1:lvl==='warn'?tokens.colorStatusWarningForeground1:tokens.colorNeutralForeground1});

    const mainView=sel?(
        <div style={css.logPanel}>
            <div style={{display:'flex',justifyContent:'space-between',marginBottom:'12px',fontFamily:'sans-serif'}}>
                <Text weight="semibold">{sel.name} - 日志</Text>
                <div style={{display:'flex',gap:'4px'}}>
                    {sel.status===1&&<ToolbarButton icon={<Pause24Regular/>} onClick={()=>handlePause(sel.id)}>暂停</ToolbarButton>}
                    {sel.status===2&&<ToolbarButton icon={<Play24Regular/>} onClick={()=>handleResume(sel.id)}>继续</ToolbarButton>}
                    {sel.status===0&&<ToolbarButton icon={<Delete24Regular/>} onClick={()=>handleCancel(sel.id)}>取消</ToolbarButton>}
                    {(sel.status===3||sel.status===4)&&<ToolbarButton icon={<ArrowDownload24Regular/>} onClick={()=>handleExport(sel.id)}>导出</ToolbarButton>}
                    {(sel.status===3||sel.status===4||sel.status===5)&&<ToolbarButton icon={<Delete24Regular/>} onClick={()=>handleDelete(sel.id)}>删除</ToolbarButton>}
                </div>
            </div>
            {logs.length===0&&<Text style={{color:tokens.colorNeutralForeground3}}>暂无日志</Text>}
            {logs.map((l,i)=>(
                <div key={i} style={logLevelStyle(l.level)}>
                    <span style={{color:tokens.colorNeutralForeground3}}>[{l.time}]</span> {l.message}
                </div>
            ))}
            <div ref={logEnd}/>
        </div>
    ):(
        <div style={css.empty}>
            <Text size={500} style={{fontWeight:'200',color:tokens.colorNeutralForeground3}}>PoiFlow</Text>
            <Text>选择一个任务查看日志，或创建新任务</Text>
        </div>
    );

    return(
        <FluentProvider theme={webLightTheme}>
            <div style={css.root}>
                <div style={css.sidebar}>
                    <div style={css.brand}><span style={css.brandDot}/> PoiFlow</div>
                    <div style={css.listHeader}>
                        <Text weight="semibold">任务列表</Text>
                        <Button appearance="subtle" icon={<Add24Regular/>} size="small" onClick={()=>setOpenNew(true)}>新建</Button>
                    </div>
                    <div style={css.listBody}>
                        {tasks.length===0&&<Text style={{color:tokens.colorNeutralForeground3,textAlign:'center',display:'block',marginTop:'32px'}}>暂无任务</Text>}
                        {tasks.map(t=>(
                            <div key={t.id} style={{...css.taskCard,...(sel?.id===t.id?css.taskCardSel:{})}} onClick={()=>selectTask(t)}>
                                <div style={css.taskName}>{t.name}</div>
                                <div style={css.taskMeta}>{t.query} · 区域{GRAN[t.areaGranularity]}→{GRAN[t.queryGranularity]}</div>
                                <div style={css.statusBar}>
                                    {badge(t.status)}
                                    <Text size={100}>{t.records}条</Text>
                                    {t.status===1&&<ProgressBar thickness="medium" style={{flex:1}}/>}
                                </div>
                            </div>
                        ))}
                    </div>
                    <div style={css.footer} onClick={()=>setOpenAk(true)}>
                        <Settings24Regular style={{marginRight:'8px'}}/>
                        <Text>Settings</Text>
                    </div>
                </div>
                <div style={css.main}>{mainView}</div>
            </div>

            <Dialog open={openNew} onOpenChange={(_e,d)=>setOpenNew(d.open)}>
                <DialogSurface>
                    <DialogBody>
                        <DialogTitle>新建任务</DialogTitle>
                        <DialogContent>
                            <div style={css.form}>
                                <Input placeholder="任务名称" value={nName} onChange={(_e,d)=>setNName(d.value)}/>
                                <Input placeholder="搜索词（如：ATM机）" value={nQuery} onChange={(_e,d)=>setNQuery(d.value)}/>
                                <Input placeholder="类型（可选，如：银行）" value={nType} onChange={(_e,d)=>setNType(d.value)}/>
                                <Input placeholder="CSV导出路径（可选，支持续采）" value={nExport} onChange={(_e,d)=>setNExport(d.value)}/>

                                <Text weight="semibold" size={200} style={{color:tokens.colorNeutralForeground2,letterSpacing:'0.5px',marginTop:'4px'}}>── 目标范围 ──</Text>
                                <Dropdown placeholder="选择目标级别" value={GRAN[nAreaGran]} onOptionSelect={(_e,d)=>{const v=Number(d.optionValue)||0;setNAreaGran(v);if(v>nQueryGran)setNQueryGran(v);setNTargets([]);setNProv('');setNCity('');}}>
                                    {GRAN.map((l,i)=><Option key={i} value={String(i)} text={l}>{l}</Option>)}
                                </Dropdown>
                                {nAreaGran>=1&&<Dropdown placeholder="省份" value={nProv} onOptionSelect={(_e,d)=>onProvChange(d.optionValue||'')}>{provinces.map(p=><Option key={p} value={p} text={p}>{p}</Option>)}</Dropdown>}
                                {nAreaGran>=2&&nProv&&<Dropdown placeholder="城市" value={nCity} onOptionSelect={(_e,d)=>onCityChange(d.optionValue||'')}>{cities.map(c=><Option key={c} value={c} text={c}>{c}</Option>)}</Dropdown>}
                                <Text weight="semibold" size={200}>选择{nAreaGran===0?'省/直辖市':nAreaGran===1?'城市':'区县'} <Badge>{nTargets.length}</Badge></Text>
                                <div style={css.targets}>
                                    {available.length===0&&<Text size={200} style={{color:'#888'}}>请先选择上级区域</Text>}
                                    {available.map(name=>(
                                        <div key={name} style={{display:'flex',alignItems:'center',gap:'8px',padding:'4px 8px',borderRadius:'4px',cursor:'pointer',background:nTargets.some(t=>t.name===name)?tokens.colorBrandBackground2:'transparent'}} onClick={()=>toggleTarget(name)}>
                                            <input type="checkbox" checked={nTargets.some(t=>t.name===name)} readOnly/>
                                            <Text size={200}>{name}</Text>
                                        </div>
                                    ))}
                                </div>

                                <Text weight="semibold" size={200} style={{color:tokens.colorNeutralForeground2,letterSpacing:'0.5px',marginTop:'4px'}}>── 搜索精度 ──</Text>
                                <Dropdown placeholder="每个目标细分到" value={GRAN[nQueryGran]} onOptionSelect={(_e,d)=>{const v=Number(d.optionValue)||nAreaGran;if(v>=nAreaGran)setNQueryGran(v);}}>
                                    {GRAN.map((l,i)=><Option key={i} value={String(i)} text={l} disabled={i<nAreaGran}>{l}</Option>)}
                                </Dropdown>
                                {nTargets.length>0&&(
                                    <div style={{padding:'10px 12px',background:tokens.colorNeutralBackground3,borderRadius:'6px',fontSize:'13px',color:tokens.colorNeutralForeground2}}>
                                        <Text size={200}>已选 <b>{nTargets.length}</b> 个{GRAN[nAreaGran]}目标，将分别搜索每个目标下所有<b>{GRAN[nQueryGran]}</b>级区域</Text>
                                        {expandCount>0&&<div style={{marginTop:'4px'}}><Text size={200}>预计发送 <b>{expandCount}</b> 次API查询</Text></div>}
                                    </div>
                                )}
                            </div>
                        </DialogContent>
                        <DialogActions>
                            <DialogTrigger disableButtonEnhancement><Button>取消</Button></DialogTrigger>
                            <Button appearance="primary" onClick={handleCreate}>创建</Button>
                        </DialogActions>
                    </DialogBody>
                </DialogSurface>
            </Dialog>

            <Dialog open={openAk} onOpenChange={(_e,d)=>setOpenAk(d.open)}>
                <DialogSurface>
                    <DialogBody>
                        <DialogTitle>AK 管理</DialogTitle>
                        <DialogContent>
                            <div style={{display:'flex',gap:'8px',marginBottom:'12px'}}>
                                <Input placeholder="输入新AK" value={newAk} onChange={(_e,d)=>setNewAk(d.value)} style={{flex:1}}/>
                                <Button appearance="primary" onClick={handleAddAk}>添加</Button>
                            </div>
                            <div style={{marginBottom:'12px',display:'flex',gap:'8px'}}>
                                <Button onClick={handleVerifyAll} disabled={!!verifying}>{verifying?'验证中...':'一键刷新所有AK状态'}</Button>
                                <Button onClick={()=>{ResetAKPool();GetAKItems().then(setAkItems).catch(()=>{});}}>重置计数</Button>
                            </div>
                            {akItems.map((it,i)=>(
                                <div key={i} style={{display:'flex',alignItems:'center',gap:'12px',padding:'8px 0',borderBottom:`1px solid ${tokens.colorNeutralStroke2}`}}>
                                    <Text size={200} style={{fontFamily:'monospace',flex:1}}>{it.ak.slice(0,16)}...</Text>
                                    <Text size={200}>本次: {it.used}</Text>
                                    <Badge appearance="filled" color={it.failed?'danger':'success'} size="small">{it.failed?'失效':'正常'}</Badge>
                                    <Button size="small" onClick={()=>handleVerifyAk(it.ak)} disabled={!!verifying}>验证</Button>
                                    <Button size="small" onClick={()=>handleRemoveAk(it.ak)}>删除</Button>
                                </div>
                            ))}
                        </DialogContent>
                        <DialogActions>
                            <DialogTrigger disableButtonEnhancement><Button>关闭</Button></DialogTrigger>
                        </DialogActions>
                    </DialogBody>
                </DialogSurface>
            </Dialog>
        </FluentProvider>
    );
}
export default App;
