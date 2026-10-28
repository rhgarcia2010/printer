export namespace book {
	
	export class Size {
	    width: number;
	    height: number;
	
	    static createFrom(source: any = {}) {
	        return new Size(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.width = source["width"];
	        this.height = source["height"];
	    }
	}
	export class Paper {
	    id: string;
	    name: string;
	    size: Size;
	
	    static createFrom(source: any = {}) {
	        return new Paper(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.size = this.convertValues(source["size"], Size);
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
	export class Side {
	    left: number;
	    right: number;
	
	    static createFrom(source: any = {}) {
	        return new Side(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.left = source["left"];
	        this.right = source["right"];
	    }
	}
	export class Sheet {
	    signature: number;
	    nest: number;
	    nestCount: number;
	    front: Side;
	    back: Side;
	
	    static createFrom(source: any = {}) {
	        return new Sheet(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.signature = source["signature"];
	        this.nest = source["nest"];
	        this.nestCount = source["nestCount"];
	        this.front = this.convertValues(source["front"], Side);
	        this.back = this.convertValues(source["back"], Side);
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
	export class Plan {
	    sourcePages: number;
	    bookPages: number;
	    blankAt: number[];
	    sheets: Sheet[];
	    signatures: number;
	
	    static createFrom(source: any = {}) {
	        return new Plan(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sourcePages = source["sourcePages"];
	        this.bookPages = source["bookPages"];
	        this.blankAt = source["blankAt"];
	        this.sheets = this.convertValues(source["sheets"], Sheet);
	        this.signatures = source["signatures"];
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
	export class Settings {
	    paperId: string;
	    sheetsPerSignature: number;
	    blanks: string;
	    scaling: string;
	    creepMm: number;
	    printerMarginMm: number;
	    safeMarginMm: number;
	    foldMarks: boolean;
	    duplex: string;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.paperId = source["paperId"];
	        this.sheetsPerSignature = source["sheetsPerSignature"];
	        this.blanks = source["blanks"];
	        this.scaling = source["scaling"];
	        this.creepMm = source["creepMm"];
	        this.printerMarginMm = source["printerMarginMm"];
	        this.safeMarginMm = source["safeMarginMm"];
	        this.foldMarks = source["foldMarks"];
	        this.duplex = source["duplex"];
	    }
	}
	
	

}

export namespace job {
	
	export class Info {
	    path: string;
	    name: string;
	    pageCount: number;
	    pageSize: book.Size;
	
	    static createFrom(source: any = {}) {
	        return new Info(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.name = source["name"];
	        this.pageCount = source["pageCount"];
	        this.pageSize = this.convertValues(source["pageSize"], book.Size);
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
	export class SidePreview {
	    sheet: number;
	    signature: number;
	    back: boolean;
	    left: string;
	    right: string;
	    image: string;
	
	    static createFrom(source: any = {}) {
	        return new SidePreview(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sheet = source["sheet"];
	        this.signature = source["signature"];
	        this.back = source["back"];
	        this.left = source["left"];
	        this.right = source["right"];
	        this.image = source["image"];
	    }
	}
	export class Result {
	    plan: book.Plan;
	    issues: preflight.Issue[];
	    sides: SidePreview[];
	    sheet: book.Size;
	
	    static createFrom(source: any = {}) {
	        return new Result(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.plan = this.convertValues(source["plan"], book.Plan);
	        this.issues = this.convertValues(source["issues"], preflight.Issue);
	        this.sides = this.convertValues(source["sides"], SidePreview);
	        this.sheet = this.convertValues(source["sheet"], book.Size);
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

export namespace main {
	
	export class Config {
	    settings: book.Settings;
	    lastFile: string;
	    printer: string;
	    reverseBacks: boolean;
	    rotateBacks: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.settings = this.convertValues(source["settings"], book.Settings);
	        this.lastFile = source["lastFile"];
	        this.printer = source["printer"];
	        this.reverseBacks = source["reverseBacks"];
	        this.rotateBacks = source["rotateBacks"];
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
	export class Options {
	    papers: book.Paper[];
	    config: Config;
	    printers: printing.Printer[];
	    directPrint: boolean;
	    defaults: book.Settings;
	
	    static createFrom(source: any = {}) {
	        return new Options(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.papers = this.convertValues(source["papers"], book.Paper);
	        this.config = this.convertValues(source["config"], Config);
	        this.printers = this.convertValues(source["printers"], printing.Printer);
	        this.directPrint = source["directPrint"];
	        this.defaults = this.convertValues(source["defaults"], book.Settings);
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
	export class PrintRequest {
	    settings: book.Settings;
	    printer: string;
	    pass: string;
	    reverseBacks: boolean;
	    rotateBacks: boolean;
	
	    static createFrom(source: any = {}) {
	        return new PrintRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.settings = this.convertValues(source["settings"], book.Settings);
	        this.printer = source["printer"];
	        this.pass = source["pass"];
	        this.reverseBacks = source["reverseBacks"];
	        this.rotateBacks = source["rotateBacks"];
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

export namespace preflight {
	
	export class Issue {
	    severity: string;
	    title: string;
	    detail: string;
	    pages?: number[];
	
	    static createFrom(source: any = {}) {
	        return new Issue(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.severity = source["severity"];
	        this.title = source["title"];
	        this.detail = source["detail"];
	        this.pages = source["pages"];
	    }
	}

}

export namespace printing {
	
	export class Printer {
	    name: string;
	    default: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Printer(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.default = source["default"];
	    }
	}

}

