export namespace picker {
	
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

}

