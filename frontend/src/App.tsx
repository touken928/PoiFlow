import {useState, useEffect, useCallback, useRef} from 'react';
import {
    Button, Text, Input, Dialog, DialogTrigger, DialogSurface,
    DialogTitle, DialogBody, DialogActions, DialogContent, Dropdown, Option,
    Badge, tokens, FluentProvider, webLightTheme, ProgressBar, TabList, Tab,
} from '@fluentui/react-components';
import {
    Settings24Regular, Add24Regular, Play24Regular, Pause24Regular,
    Delete24Regular, ArrowDownload24Regular,
} from '@fluentui/react-icons';

import {
    GetProvinces, GetCities, GetCounties, CreateTask, GetTasks,
    CancelTask, PauseTask, ResumeTask, DeleteTask, RetryTask, GetAKItems,
    ResetAKPool, AddAK, RemoveAK, GetTaskLogs, ExportTaskDialog, ExportTaskGeoJSON, ExpandCount,
    GetExportConfig, SetExportConfig, GetVersion,
} from '../wailsjs/go/main/App';
import {EventsOn} from '../wailsjs/runtime/runtime';

const STATUS_LABELS = ['等待中','执行中','已暂停','已完成','失败','已取消'];
const GRAN = ['省级','市级','区县级'];

interface Task {
    id: string; name: string;
    queries: {query:string;type:string}[];
    exportPath: string;
    areaGranularity: number; queryGranularity: number;
    targets: {province:string;city:string;name:string}[];
    status: number; progress: string; records: number; error: string;
    createdAt: string; updatedAt: string;
}
interface LogEntry{time:string;message:string;level:string}
interface AKItem{name:string;ak:string;used:number;failed:boolean;failMsg:string}

const st = (styles: Record<string,any>) => styles;

const css = st({
    root:{height:'100vh',display:'flex',background:tokens.colorNeutralBackground1},
    sidebar:{display:'flex',flexDirection:'column',borderRight:`1px solid ${tokens.colorNeutralStroke1}`,background:tokens.colorNeutralBackground2},
    brand:{height:'56px',padding:'0 20px',fontSize:'20px',fontWeight:'700',borderBottom:`1px solid ${tokens.colorNeutralStroke2}`,display:'flex',alignItems:'center',justifyContent:'space-between',gap:'8px'},
    brandDot:{width:'10px',height:'10px',borderRadius:'50%',background:tokens.colorBrandForeground1,display:'inline-block'},
    listHeader:{height:'56px',display:'flex',alignItems:'center',justifyContent:'space-between',padding:'0 20px',borderBottom:`1px solid ${tokens.colorNeutralStroke1}`},
    listBody:{flex:'1',overflowY:'auto',padding:'8px'},
    taskCard:{padding:'12px',marginBottom:'8px',borderRadius:'8px',border:`1px solid ${tokens.colorNeutralStroke2}`,background:tokens.colorNeutralBackground1,cursor:'pointer'},
    taskCardSel:{border:`2px solid ${tokens.colorBrandStroke1}`},
    taskName:{fontWeight:'600',marginBottom:'2px'},
    taskMeta:{fontSize:'11px',color:tokens.colorNeutralForeground3,marginBottom:'6px'},
    statusBar:{display:'flex',alignItems:'center',gap:'8px',marginTop:'4px'},
    footer:{padding:'10px 16px',borderTop:`1px solid ${tokens.colorNeutralStroke1}`,display:'flex',alignItems:'center',cursor:'pointer',color:tokens.colorNeutralForeground2},
    main:{flex:'1',display:'flex',flexDirection:'column',overflow:'hidden'},
    empty:{flex:'1',display:'flex',alignItems:'center',justifyContent:'center',color:tokens.colorNeutralForeground3,flexDirection:'column',gap:'8px'},
    logPanel:{flex:'1',display:'flex',flexDirection:'column',overflow:'hidden'},
    logHeader:{height:'56px',padding:'0 20px',fontFamily:'sans-serif',borderBottom:`1px solid ${tokens.colorNeutralStroke2}`,background:tokens.colorNeutralBackground1,flexShrink:0,display:'flex',alignItems:'center',justifyContent:'space-between'},
    logBody:{flex:'1',overflowY:'auto',padding:'0 20px 20px',fontFamily:'monospace',fontSize:'13px',lineHeight:'1.7'},
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
    const [nQueries,setNQueries]=useState([{query:'',type:''}]);
    const [nQueryGran,setNQueryGran]=useState(0);
    const [provinces,setProvinces]=useState<string[]>([]);
    const [nTargets,setNTargets]=useState<{province:string;city:string;name:string}[]>([]);
    const [expandCount,setExpandCount]=useState(0);
    const [treeExp,setTreeExp]=useState<Record<string,boolean>>({});
    const [treeCities,setTreeCities]=useState<Record<string,string[]>>({});
    const [treeCounties,setTreeCounties]=useState<Record<string,string[]>>({});
    const [newAkName,setNewAkName]=useState('');
    const [newAkKey,setNewAkKey]=useState('');
    const [settingsTab,setSettingsTab]=useState('ak');
    const [appVersion,setAppVersion]=useState('');
    const [exportFields,setExportFields]=useState<string[]>([]);
    const [sidebarWidth,setSidebarWidth]=useState(300);
    const dragging=useRef(false);
    const allFields=['name','address','telephone','province','city','area','uid','query','type','taskName','target'];
    const fieldLabels:Record<string,string>={name:'名称',address:'地址',telephone:'电话',province:'省份',city:'城市',area:'区县',uid:'UID',query:'搜索词',type:'分类',taskName:'任务名',target:'搜索目标'};

    useEffect(()=>{
        const onMove=(e:MouseEvent)=>{if(dragging.current){setSidebarWidth(Math.max(240,Math.min(600,e.clientX)));}};
        const onUp=()=>{dragging.current=false;document.body.style.cursor='';document.body.style.userSelect='';};
        window.addEventListener('mousemove',onMove);
        window.addEventListener('mouseup',onUp);
        return ()=>{window.removeEventListener('mousemove',onMove);window.removeEventListener('mouseup',onUp);};
    },[]);

    useEffect(()=>{GetExportConfig().then(c=>setExportFields(c.Fields||allFields)).catch(()=>{});},[]);

    const loadAll=useCallback(()=>{GetTasks().then(setTasks).catch(()=>{});GetAKItems().then(setAkItems).catch(()=>{});},[]);

    useEffect(()=>{GetProvinces().then(setProvinces).catch(()=>{});GetVersion().then(setAppVersion).catch(()=>{});loadAll();const c=[EventsOn('task:added',loadAll),EventsOn('task:updated',loadAll),EventsOn('task:completed',loadAll),EventsOn('task:failed',loadAll),EventsOn('task:deleted',loadAll)];return()=>{c.forEach(f=>f());};},[loadAll]);

    useEffect(()=>{if(sel){const u=tasks.find(t=>t.id===sel.id);if(u)setSel(u);}},[tasks]);

    useEffect(()=>{
        if(!sel)return;
        const off=EventsOn('task:log',(data:any)=>{
            if(data&&data.taskID===sel.id&&data.entry){
                setLogs(prev=>[...prev,data.entry]);
            }
        });
        GetTaskLogs(sel.id).then(setLogs).catch(()=>{});
        return ()=>{off();};
    },[sel?.id]);

    useEffect(()=>{const iv=setInterval(()=>{GetTasks().then(setTasks).catch(()=>{});},3000);return()=>clearInterval(iv);},[]);

    useEffect(()=>{if(logEnd.current)logEnd.current.scrollIntoView({behavior:'smooth'});},[logs]);

    useEffect(()=>{if(nTargets.length>0)ExpandCount(0,nQueryGran,nTargets).then(c=>setExpandCount(c*nQueries.filter(q=>q.query.trim()||q.type.trim()).length||c)).catch(()=>setExpandCount(0));else setExpandCount(0);},[nTargets,nQueryGran,nQueries]);

    const toggleTarget=(province:string,city:string,name:string)=>setNTargets(p=>{const i=p.findIndex(t=>t.name===name);if(i>=0)return p.filter((_,j)=>j!==i);return[...p,{province,city,name}];});

    const toggleExp=async(level:string,parent:string,name:string)=>{
        const key=parent?parent+'/'+name:name;
        if(treeExp[key]){const n={...treeExp};delete n[key];setTreeExp(n);return;}
        setTreeExp({...treeExp,[key]:true});
        if(level==='prov'&&!treeCities[name]){const c=await GetCities(name);setTreeCities({...treeCities,[name]:c||[]});}
        if(level==='city'&&!treeCounties[key]){const c=await GetCounties(parent,name);setTreeCounties({...treeCounties,[key]:c||[]});}
    };

    function renderTree(){
        let items:any[]=[];
        for(let p of provinces){
            let ex=treeExp[p];
            items.push(<div key={p} style={{display:'flex',alignItems:'center',gap:4,padding:'2px 0',cursor:'pointer',userSelect:'none'}}>
                <span onClick={()=>toggleExp('prov','',p)} style={{width:16,textAlign:'center',fontSize:11,color:tokens.colorNeutralForeground3}}>{ex?'▼':'▶'}</span>
                <input type="checkbox" checked={nTargets.some(t=>t.name===p)} onChange={()=>toggleTarget(p,'',p)}/>
                <Text size={200}>{p}</Text>
            </div>);
            if(ex&&treeCities[p]){
                for(let c of treeCities[p]){
                    let ck=p+'/'+c,ex2=treeExp[ck];
                    items.push(<div key={ck} style={{display:'flex',alignItems:'center',gap:4,padding:'2px 0',paddingLeft:24,cursor:'pointer',userSelect:'none'}}>
                        <span onClick={()=>toggleExp('city',p,c)} style={{width:16,textAlign:'center',fontSize:11,color:tokens.colorNeutralForeground3}}>{ex2?'▼':'▶'}</span>
                        <input type="checkbox" checked={nTargets.some(t=>t.name===c)} onChange={()=>toggleTarget(p,'',c)}/>
                        <Text size={200}>{c}</Text>
                    </div>);
                    if(ex2&&treeCounties[ck]){
                        for(let co of treeCounties[ck]){
                            items.push(<div key={ck+'/'+co} style={{display:'flex',alignItems:'center',gap:4,padding:'2px 0',paddingLeft:48,cursor:'pointer',userSelect:'none'}}>
                                <span style={{width:16}}/>
                                <input type="checkbox" checked={nTargets.some(t=>t.name===co)} onChange={()=>toggleTarget(p,c,co)}/>
                                <Text size={200}>{co}</Text>
                            </div>);
                        }
                    }
                }
            }
        }
        return items;
    }

    const handleCreate = async () => {
        const valid = nQueries.filter(q => q.query.trim() || q.type.trim());
        if(!nName||valid.length===0||nTargets.length===0){setMsg('请填写任务名称、搜索词/分类并选择目标');return;}
        try{await CreateTask(nName,'',0,nQueryGran,nTargets,valid);setOpenNew(false);setMsg('');setNName('');setNQueries([{query:'',type:''}]);
        setNTargets([]);setExpandCount(0);setTreeExp({});}catch(e:any){setMsg('创建失败: '+e);}}
    const selectTask=async(t:Task)=>{setSel(t);try{setLogs(await GetTaskLogs(t.id)||[]);}catch(e){}};
    const handleExport=async(id:string)=>{console.log('export csv',id);try{const r=await ExportTaskDialog(id);console.log('export result',r);if(r)setMsg(r);}catch(e:any){console.error(e);setMsg('导出失败: '+e);}};
    const handleExportGeoJSON=async(id:string)=>{console.log('export geojson',id);try{const r=await ExportTaskGeoJSON(id);console.log('export result',r);if(r)setMsg(r);}catch(e:any){console.error(e);setMsg('导出GeoJSON失败: '+e);}};
    const handlePause=async(id:string)=>{await PauseTask(id);loadAll();};
    const handleResume=async(id:string)=>{await ResumeTask(id);loadAll();};
    const handleCancel=async(id:string)=>{await CancelTask(id);loadAll();};
    const handleDelete=async(id:string)=>{await DeleteTask(id);loadAll();if(sel?.id===id){setSel(null);setLogs([]);}};
    const handleRetry=async(id:string)=>{await RetryTask(id);loadAll();};
    const handleAddAk=async()=>{if(!newAkKey){setMsg('请输入AK');return;}const r=await AddAK(newAkName,newAkKey);if(r)setMsg(r);else{setNewAkName('');setNewAkKey('');setMsg('AK已添加');}GetAKItems().then(setAkItems).catch(()=>{});};
    const handleRemoveAk=async(ak:string)=>{const r=await RemoveAK(ak);if(r)setMsg(r);GetAKItems().then(setAkItems).catch(()=>{});};

    const toggleField=async(f:string)=>{const next=exportFields.includes(f)?exportFields.filter(x=>x!==f):[...exportFields,f];setExportFields(next);await SetExportConfig(next);};


    const badge=(st:number)=>{const m:Record<number,'success'|'warning'|'danger'|'important'>={0:'important',1:'warning',2:'warning',3:'success',4:'danger',5:'important'};return<Badge appearance="filled" color={m[st]||'important'} size="small">{STATUS_LABELS[st]||'?'}</Badge>;};
    const nQueriesText=(t:Task)=>{if(!t.queries||t.queries.length===0)return'';return t.queries.map(q=>q.query+(q.type?'('+q.type+')':'')).join(', ');};

    const logLevelStyle=(lvl:string)=>({color:lvl==='error'?tokens.colorStatusDangerForeground1:lvl==='warn'?tokens.colorStatusWarningForeground1:tokens.colorNeutralForeground1});

    const mainView=sel?(
        <div style={css.logPanel}>
            <div style={css.logHeader}>
                <Text weight="semibold">{sel.name} - 日志</Text>
                <div style={{display:'flex',gap:'4px',flexShrink:0}}>
                        {sel.status===1&&<Button icon={<Pause24Regular/>} onClick={()=>handlePause(sel.id)}>暂停</Button>}
                        {sel.status===2&&<Button icon={<Play24Regular/>} onClick={()=>handleResume(sel.id)}>继续</Button>}
                        {(sel.status===1||sel.status===2)&&<Button icon={<Delete24Regular/>} onClick={()=>handleDelete(sel.id)}>删除</Button>}
                        {sel.status===0&&<Button icon={<Delete24Regular/>} onClick={()=>handleCancel(sel.id)}>取消</Button>}
                        {(sel.status===3||sel.status===4)&&<><Button icon={<ArrowDownload24Regular/>} onClick={()=>handleExport(sel.id)}>CSV</Button><Button icon={<ArrowDownload24Regular/>} onClick={()=>handleExportGeoJSON(sel.id)}>GeoJSON</Button></>}
                        {sel.status===4&&<Button icon={<Play24Regular/>} onClick={()=>handleRetry(sel.id)}>重试</Button>}
                        {(sel.status===3||sel.status===4||sel.status===5)&&<Button icon={<Delete24Regular/>} onClick={()=>handleDelete(sel.id)}>删除</Button>}
                        {sel.status===5&&<Button icon={<Delete24Regular/>} onClick={()=>handleDelete(sel.id)}>删除</Button>}
                </div>
            </div>
            <div style={css.logBody}>
                {logs.length===0&&<Text style={{color:tokens.colorNeutralForeground3}}>暂无日志</Text>}
                {logs.map((l,i)=>(
                    <div key={i} style={logLevelStyle(l.level)}>
                        <span style={{color:tokens.colorNeutralForeground3}}>[{l.time}]</span> {l.message}
                    </div>
                ))}
                <div ref={logEnd}/>
            </div>
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
                <div style={{...css.sidebar, width: sidebarWidth, flexShrink: 0}}>
                    <div style={css.brand}>
                        <span>
                            <span style={{...css.brandDot, marginRight: '8px'}}/>
                            PoiFlow
                        </span>
                        <Button appearance="subtle" icon={<Add24Regular/>} size="small" onClick={()=>setOpenNew(true)}>新建</Button>
                    </div>
                    <div style={css.listBody}>
                        {tasks.length===0&&<Text style={{color:tokens.colorNeutralForeground3,textAlign:'center',display:'block',marginTop:'32px'}}>暂无任务</Text>}
                        {tasks.map(t=>(
                            <div key={t.id} style={{...css.taskCard,...(sel?.id===t.id?css.taskCardSel:{})}} onClick={()=>selectTask(t)}>
                                <div style={css.taskName}>{t.name}</div>
                                <div style={css.taskMeta}>{nQueriesText(t)} · 区域{GRAN[t.areaGranularity]}→{GRAN[t.queryGranularity]}</div>
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
                <div
                    style={{width:'2px',cursor:'col-resize',background:tokens.colorNeutralStroke2,flexShrink:0}}
                    onMouseDown={()=>{dragging.current=true;document.body.style.cursor='col-resize';document.body.style.userSelect='none';}}
                    onMouseEnter={e=>{e.currentTarget.style.background=tokens.colorBrandStroke1 as string}}
                    onMouseLeave={e=>{e.currentTarget.style.background=tokens.colorNeutralStroke2 as string}}
                />
                <div style={css.main}>{mainView}</div>
            </div>

            <Dialog open={openNew} onOpenChange={(_e,d)=>setOpenNew(d.open)}>
                <DialogSurface>
                    <DialogBody>
                        <DialogTitle>新建任务</DialogTitle>
                        <DialogContent>
                            <div style={css.form}>
                                <Input placeholder="任务名称" value={nName} onChange={(_e,d)=>setNName(d.value)}/>

                                <Text weight="semibold" size={200} style={{color:tokens.colorNeutralForeground2,letterSpacing:'0.5px',marginTop:'4px'}}>── 搜索词和分类 ──</Text>
                                {nQueries.map((q,i)=>(
                                    <div key={i} style={{display:'flex',gap:'8px',alignItems:'center'}}>
                                        <Input placeholder="搜索词" value={q.query} onChange={(_e,d)=>{const copy=[...nQueries];copy[i]={...copy[i],query:d.value};setNQueries(copy);}} style={{flex:1}}/>
                                        <Input placeholder="分类" value={q.type} onChange={(_e,d)=>{const copy=[...nQueries];copy[i]={...copy[i],type:d.value};setNQueries(copy);}} style={{flex:1}}/>
                                        <Button disabled={nQueries.length<=1} onClick={()=>setNQueries(nQueries.filter((_,j)=>j!==i))}>×</Button>
                                    </div>
                                ))}
                                <Button appearance="subtle" size="small" onClick={()=>setNQueries([...nQueries,{query:'',type:''}])}>+ 添加搜索词</Button>

                                <Text weight="semibold" size={200} style={{color:tokens.colorNeutralForeground2,letterSpacing:'0.5px',marginTop:'4px'}}>── 目标范围 ──</Text>
                                <div style={css.targets}>
                                    {renderTree()}
                                </div>

                                <Text weight="semibold" size={200} style={{color:tokens.colorNeutralForeground2,letterSpacing:'0.5px',marginTop:'4px'}}>── 搜索精度 ──</Text>
                                <Dropdown placeholder="每个目标细分到" value={GRAN[nQueryGran]} onOptionSelect={(_e,d)=>{const v=Number(d.optionValue)||0;setNQueryGran(v);}}>
                                    {GRAN.map((l,i)=><Option key={i} value={String(i)} text={l} disabled={i<0}>{l}</Option>)}
                                </Dropdown>
                                {nTargets.length>0&&(
                                    <div style={{padding:'10px 12px',background:tokens.colorNeutralBackground3,borderRadius:'6px',fontSize:'13px',color:tokens.colorNeutralForeground2}}>
                                        <Text size={200}>已选 <b>{nTargets.length}</b> 个目标，将分别搜索每个目标下所有<b>{GRAN[nQueryGran]}</b>级区域</Text>
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
                <DialogSurface style={{minWidth:'600px'}}>
                    <DialogBody>
                        <DialogTitle>Settings</DialogTitle>
                        <DialogContent>
                            <TabList selectedValue={settingsTab} onTabSelect={(_e,d)=>setSettingsTab(d.value as string)} style={{marginBottom:'16px'}}>
                                <Tab value="ak">API Keys</Tab>
                                <Tab value="export">导出</Tab>
                                <Tab value="about">关于</Tab>
                            </TabList>
                            {settingsTab==='ak'&&<>
                                <div style={{display:'flex',gap:'8px',marginBottom:'12px'}}>
                                    <Input placeholder="名称" value={newAkName} onChange={(_e,d)=>setNewAkName(d.value)} style={{width:'140px'}}/>
                                    <Input placeholder="API Key" value={newAkKey} onChange={(_e,d)=>setNewAkKey(d.value)} style={{flex:1}}/>
                                    <Button appearance="primary" onClick={handleAddAk}>添加</Button>
                                </div>
                                <div style={{marginBottom:'12px',display:'flex',gap:'8px'}}>
                                    <Button onClick={()=>{ResetAKPool();GetAKItems().then(setAkItems).catch(()=>{});}}>重置AK状态</Button>
                                </div>
                                {akItems.map((it,i)=>(
                                    <div key={i} style={{display:'flex',alignItems:'center',gap:'12px',padding:'8px 0',borderBottom:`1px solid ${tokens.colorNeutralStroke2}`}}>
                                        <Text size={200} style={{width:'100px',fontWeight:'600'}}>{it.name||'未命名'}</Text>
                                        <Text size={200} style={{fontFamily:'monospace',flex:1,color:tokens.colorNeutralForeground3}}>{it.ak.slice(0,16)}...</Text>
                                        <Text size={200}>本次: {it.used}</Text>
                                        <Badge appearance="filled" color={it.failed?'danger':'success'} size="small">{it.failed?'失效: '+it.failMsg:'正常'}</Badge>
                                        <Button size="small" onClick={()=>handleRemoveAk(it.ak)}>删除</Button>
                                    </div>
                                ))}
                            </>}
                            {settingsTab==='export'&&<div style={{display:'flex',flexDirection:'column',gap:'8px'}}>
                                <Text weight="semibold">导出字段设置</Text>
                                <Text size={200} style={{color:tokens.colorNeutralForeground3}}>经纬度始终包含，选择其他要导出的字段：</Text>
                                {allFields.map(f=>(
                                    <div key={f} style={{display:'flex',alignItems:'center',gap:'8px',cursor:'pointer'}} onClick={()=>toggleField(f)}>
                                        <input type="checkbox" checked={exportFields.includes(f)} readOnly/>
                                        <Text>{fieldLabels[f]||f}</Text>
                                    </div>
                                ))}
                            </div>}
                            {settingsTab==='about'&&<div style={{display:'flex',flexDirection:'column',gap:'12px'}}>
                                <Text weight="semibold" size={400}>PoiFlow</Text>
                                <Text size={200}>版本: {appVersion || 'dev'}</Text>
                                <Text size={200}>百度POI数据采集桌面工具</Text>
                                <div style={{height:'1px',background:tokens.colorNeutralStroke2,margin:'4px 0'}}/>
                                <Text weight="semibold">作者</Text>
                                <Text size={200}>touken928</Text>
                                <Text weight="semibold">仓库地址</Text>
                                <Text size={200}><a href="https://github.com/touken928/PoiFlow" target="_blank" style={{color:tokens.colorBrandForeground1}}>github.com/touken928/PoiFlow</a></Text>
                                <Text weight="semibold">开源依赖</Text>
                                <Text size={200}>Go: Wails v2, yaml.v3 | 前端: React, Fluent UI, Vite</Text>
                                <div style={{height:'1px',background:tokens.colorNeutralStroke2,margin:'4px 0'}}/>
                                <Text weight="semibold">免责声明</Text>
                                <Text size={200} style={{color:tokens.colorNeutralForeground3}}>本软件仅供学习研究使用。用户必须遵守百度地图API服务协议及相关法律法规，不得用于任何非法用途。使用者需自行承担全部法律责任。</Text>
                            </div>}
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
