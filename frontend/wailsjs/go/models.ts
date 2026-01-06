export namespace models {
	
	export class DiffItem {
	    relPath: string;
	    type: string;
	    selected: boolean;
	    sourcePath: string;
	
	    static createFrom(source: any = {}) {
	        return new DiffItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.relPath = source["relPath"];
	        this.type = source["type"];
	        this.selected = source["selected"];
	        this.sourcePath = source["sourcePath"];
	    }
	}
	export class CompareResult {
	    items: DiffItem[];
	    totalFiles: number;
	    added: number;
	    modified: number;
	    deleted: number;
	
	    static createFrom(source: any = {}) {
	        return new CompareResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.items = this.convertValues(source["items"], DiffItem);
	        this.totalFiles = source["totalFiles"];
	        this.added = source["added"];
	        this.modified = source["modified"];
	        this.deleted = source["deleted"];
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
	export class ExcludeRule {
	    pattern: string;
	    type: string;
	    isDir: boolean;
	    enabled: boolean;
	    comment: string;
	
	    static createFrom(source: any = {}) {
	        return new ExcludeRule(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.pattern = source["pattern"];
	        this.type = source["type"];
	        this.isDir = source["isDir"];
	        this.enabled = source["enabled"];
	        this.comment = source["comment"];
	    }
	}
	export class Config {
	    lastZipPath: string;
	    lastWorkDir: string;
	    lastOutputDir: string;
	    excludeRules: ExcludeRule[];
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.lastZipPath = source["lastZipPath"];
	        this.lastWorkDir = source["lastWorkDir"];
	        this.lastOutputDir = source["lastOutputDir"];
	        this.excludeRules = this.convertValues(source["excludeRules"], ExcludeRule);
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
	
	export class DiffLine {
	    type: string;
	    content: string;
	
	    static createFrom(source: any = {}) {
	        return new DiffLine(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.content = source["content"];
	    }
	}
	
	export class TextDiff {
	    oldContent: string;
	    newContent: string;
	    lines: DiffLine[];
	
	    static createFrom(source: any = {}) {
	        return new TextDiff(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.oldContent = source["oldContent"];
	        this.newContent = source["newContent"];
	        this.lines = this.convertValues(source["lines"], DiffLine);
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

