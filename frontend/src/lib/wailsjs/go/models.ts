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
	
	export class FindFilesResult {
	    filename: string;
	    relativePath: string;
	    absolutePath: string;
	    icon: string;
	    iconColor: string;
	
	    static createFrom(source: any = {}) {
	        return new FindFilesResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.filename = source["filename"];
	        this.relativePath = source["relativePath"];
	        this.absolutePath = source["absolutePath"];
	        this.icon = source["icon"];
	        this.iconColor = source["iconColor"];
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

}

