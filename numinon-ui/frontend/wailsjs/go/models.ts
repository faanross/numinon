export namespace main {
	
	export class SystemInfo {
	    os: string;
	    arch: string;
	    hostname: string;
	    current_time: string;
	
	    static createFrom(source: any = {}) {
	        return new SystemInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.os = source["os"];
	        this.arch = source["arch"];
	        this.hostname = source["hostname"];
	        this.current_time = source["current_time"];
	    }
	}

}

