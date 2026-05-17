export namespace domain {
	
	export class CSVDelimiter {
	    name: string;
	    value: string;
	
	    static createFrom(source: any = {}) {
	        return new CSVDelimiter(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.value = source["value"];
	    }
	}
	export class Item {
	    itemNumber: string;
	    prices: string[];
	
	    static createFrom(source: any = {}) {
	        return new Item(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.itemNumber = source["itemNumber"];
	        this.prices = source["prices"];
	    }
	}
	export class Dataset {
	    itemHeader: string;
	    priceHeaders: string[];
	    items: Item[];
	    exportPrecision: number[];
	
	    static createFrom(source: any = {}) {
	        return new Dataset(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.itemHeader = source["itemHeader"];
	        this.priceHeaders = source["priceHeaders"];
	        this.items = this.convertValues(source["items"], Item);
	        this.exportPrecision = source["exportPrecision"];
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
	export class DecimalSettings {
	    mode: string;
	    places: number;
	
	    static createFrom(source: any = {}) {
	        return new DecimalSettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mode = source["mode"];
	        this.places = source["places"];
	    }
	}
	export class ExportReport {
	    path: string;
	    itemsWritten: number;
	    bytesWritten: number;
	    durationMillis: number;
	
	    static createFrom(source: any = {}) {
	        return new ExportReport(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.itemsWritten = source["itemsWritten"];
	        this.bytesWritten = source["bytesWritten"];
	        this.durationMillis = source["durationMillis"];
	    }
	}
	export class GreaterThanFilter {
	    column: number;
	    value: string;
	
	    static createFrom(source: any = {}) {
	        return new GreaterThanFilter(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.column = source["column"];
	        this.value = source["value"];
	    }
	}
	
	export class LessThanFilter {
	    column: number;
	    value: string;
	
	    static createFrom(source: any = {}) {
	        return new LessThanFilter(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.column = source["column"];
	        this.value = source["value"];
	    }
	}
	export class PreviewCell {
	    column: number;
	    original: string;
	    newValue: string;
	    diff: string;
	    changed: boolean;
	
	    static createFrom(source: any = {}) {
	        return new PreviewCell(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.column = source["column"];
	        this.original = source["original"];
	        this.newValue = source["newValue"];
	        this.diff = source["diff"];
	        this.changed = source["changed"];
	    }
	}
	export class RuleStat {
	    ruleId: string;
	    itemsMatched: number;
	    cellsChanged: number;
	    warning?: string;
	
	    static createFrom(source: any = {}) {
	        return new RuleStat(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ruleId = source["ruleId"];
	        this.itemsMatched = source["itemsMatched"];
	        this.cellsChanged = source["cellsChanged"];
	        this.warning = source["warning"];
	    }
	}
	export class PreviewRow {
	    itemNumber: string;
	    cells: PreviewCell[];
	
	    static createFrom(source: any = {}) {
	        return new PreviewRow(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.itemNumber = source["itemNumber"];
	        this.cells = this.convertValues(source["cells"], PreviewCell);
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
	export class PreviewResult {
	    totalAffectedItems: number;
	    totalAffectedCells: number;
	    totalItems: number;
	    sampleSize: number;
	    rows: PreviewRow[];
	    perRuleStats: RuleStat[];
	    generatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new PreviewResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.totalAffectedItems = source["totalAffectedItems"];
	        this.totalAffectedCells = source["totalAffectedCells"];
	        this.totalItems = source["totalItems"];
	        this.sampleSize = source["sampleSize"];
	        this.rows = this.convertValues(source["rows"], PreviewRow);
	        this.perRuleStats = this.convertValues(source["perRuleStats"], RuleStat);
	        this.generatedAt = source["generatedAt"];
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
	
	export class RuleAction {
	    type: string;
	    value: string;
	
	    static createFrom(source: any = {}) {
	        return new RuleAction(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.value = source["value"];
	    }
	}
	export class RuleFilter {
	    applyToAll: boolean;
	    wildcard?: string;
	    lessThan?: LessThanFilter;
	    greaterThan?: GreaterThanFilter;
	
	    static createFrom(source: any = {}) {
	        return new RuleFilter(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.applyToAll = source["applyToAll"];
	        this.wildcard = source["wildcard"];
	        this.lessThan = this.convertValues(source["lessThan"], LessThanFilter);
	        this.greaterThan = this.convertValues(source["greaterThan"], GreaterThanFilter);
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
	export class Rule {
	    id: string;
	    filter: RuleFilter;
	    columns: number[];
	    action: RuleAction;
	    decimalSettings: DecimalSettings;
	
	    static createFrom(source: any = {}) {
	        return new Rule(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.filter = this.convertValues(source["filter"], RuleFilter);
	        this.columns = source["columns"];
	        this.action = this.convertValues(source["action"], RuleAction);
	        this.decimalSettings = this.convertValues(source["decimalSettings"], DecimalSettings);
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
	    csvDelimiter: CSVDelimiter;
	    defaultDecimalPlaces: number;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.csvDelimiter = this.convertValues(source["csvDelimiter"], CSVDelimiter);
	        this.defaultDecimalPlaces = source["defaultDecimalPlaces"];
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

