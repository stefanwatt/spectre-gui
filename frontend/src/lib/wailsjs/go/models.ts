export namespace match {
	
	export class Match {
	    Id: string;
	    FileName: string;
	    AbsolutePath: string;
	    MatchedLine: string;
	    TextBeforeMatch: string;
	    TextAfterMatch: string;
	    MatchedText: string;
	    ReplacementText: string;
	    Row: number;
	    Col: number;
	    Html: string;
	
	    static createFrom(source: any = {}) {
	        return new Match(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Id = source["Id"];
	        this.FileName = source["FileName"];
	        this.AbsolutePath = source["AbsolutePath"];
	        this.MatchedLine = source["MatchedLine"];
	        this.TextBeforeMatch = source["TextBeforeMatch"];
	        this.TextAfterMatch = source["TextAfterMatch"];
	        this.MatchedText = source["MatchedText"];
	        this.ReplacementText = source["ReplacementText"];
	        this.Row = source["Row"];
	        this.Col = source["Col"];
	        this.Html = source["Html"];
	    }
	}
	export class MatchesOfFile {
	    Path: string;
	    Matches: Match[];
	
	    static createFrom(source: any = {}) {
	        return new MatchesOfFile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Path = source["Path"];
	        this.Matches = this.convertValues(source["Matches"], Match);
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

export namespace neovim {
	
	export class QuickfixEntry {
	    filepath: string;
	    row: number;
	    col: number;
	    text: string;
	
	    static createFrom(source: any = {}) {
	        return new QuickfixEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.filepath = source["filepath"];
	        this.row = source["row"];
	        this.col = source["col"];
	        this.text = source["text"];
	    }
	}

}

export namespace picker {
	
	export class LiveGrepPage {
	    results: match.MatchesOfFile[];
	    pageIndex: number;
	    totalPages: number;
	    totalResults: number;
	    totalFiles: number;
	
	    static createFrom(source: any = {}) {
	        return new LiveGrepPage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.results = this.convertValues(source["results"], match.MatchesOfFile);
	        this.pageIndex = source["pageIndex"];
	        this.totalPages = source["totalPages"];
	        this.totalResults = source["totalResults"];
	        this.totalFiles = source["totalFiles"];
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
	export class LiveGrepPickerState {
	    searchTerm: string;
	    dir: string;
	    include: string;
	    exclude: string;
	    caseSensitive: boolean;
	    regex: boolean;
	    matchWholeWord: boolean;
	    totalResults: number;
	    totalFiles: number;
	
	    static createFrom(source: any = {}) {
	        return new LiveGrepPickerState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.searchTerm = source["searchTerm"];
	        this.dir = source["dir"];
	        this.include = source["include"];
	        this.exclude = source["exclude"];
	        this.caseSensitive = source["caseSensitive"];
	        this.regex = source["regex"];
	        this.matchWholeWord = source["matchWholeWord"];
	        this.totalResults = source["totalResults"];
	        this.totalFiles = source["totalFiles"];
	    }
	}
	export class Pagination {
	    PageIndex: number;
	
	    static createFrom(source: any = {}) {
	        return new Pagination(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.PageIndex = source["PageIndex"];
	    }
	}
	export class PickerResult {
	    filename: string;
	    relativePath: string;
	    absolutePath: string;
	    icon: string;
	    iconColor: string;
	    text: string;
	    row: number;
	    col: number;
	
	    static createFrom(source: any = {}) {
	        return new PickerResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.filename = source["filename"];
	        this.relativePath = source["relativePath"];
	        this.absolutePath = source["absolutePath"];
	        this.icon = source["icon"];
	        this.iconColor = source["iconColor"];
	        this.text = source["text"];
	        this.row = source["row"];
	        this.col = source["col"];
	    }
	}

}

