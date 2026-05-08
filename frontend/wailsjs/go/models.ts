export namespace main {
	
	export class AKInfo {
	    name: string;
	    ak: string;
	    used: number;
	    failed: boolean;
	    failMsg: string;
	
	    static createFrom(source: any = {}) {
	        return new AKInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.ak = source["ak"];
	        this.used = source["used"];
	        this.failed = source["failed"];
	        this.failMsg = source["failMsg"];
	    }
	}
	export class SearchTermInput {
	    query: string;
	    type: string;
	
	    static createFrom(source: any = {}) {
	        return new SearchTermInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.query = source["query"];
	        this.type = source["type"];
	    }
	}
	export class TaskTargetInput {
	    province: string;
	    city: string;
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new TaskTargetInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.province = source["province"];
	        this.city = source["city"];
	        this.name = source["name"];
	    }
	}

}

export namespace store {
	
	export class ExportConfig {
	    Fields: string[];
	
	    static createFrom(source: any = {}) {
	        return new ExportConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Fields = source["Fields"];
	    }
	}

}

export namespace task {
	
	export class LogEntry {
	    time: string;
	    message: string;
	    level: string;
	
	    static createFrom(source: any = {}) {
	        return new LogEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.time = source["time"];
	        this.message = source["message"];
	        this.level = source["level"];
	    }
	}
	export class SearchTerm {
	    query: string;
	    type: string;
	
	    static createFrom(source: any = {}) {
	        return new SearchTerm(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.query = source["query"];
	        this.type = source["type"];
	    }
	}
	export class Target {
	    province: string;
	    city: string;
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new Target(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.province = source["province"];
	        this.city = source["city"];
	        this.name = source["name"];
	    }
	}
	export class Task {
	    id: string;
	    name: string;
	    queries: SearchTerm[];
	    exportPath: string;
	    areaGranularity: number;
	    queryGranularity: number;
	    targets: Target[];
	    status: number;
	    progress: string;
	    records: number;
	    completedTargets: number;
	    error: string;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new Task(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.queries = this.convertValues(source["queries"], SearchTerm);
	        this.exportPath = source["exportPath"];
	        this.areaGranularity = source["areaGranularity"];
	        this.queryGranularity = source["queryGranularity"];
	        this.targets = this.convertValues(source["targets"], Target);
	        this.status = source["status"];
	        this.progress = source["progress"];
	        this.records = source["records"];
	        this.completedTargets = source["completedTargets"];
	        this.error = source["error"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

