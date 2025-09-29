export namespace frontend {
	
	export class AgentDTO {
	    id: string;
	    hostname: string;
	    os: string;
	    status: string;
	    lastSeen: string;
	    ipAddress: string;
	
	    static createFrom(source: any = {}) {
	        return new AgentDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.hostname = source["hostname"];
	        this.os = source["os"];
	        this.status = source["status"];
	        this.lastSeen = source["lastSeen"];
	        this.ipAddress = source["ipAddress"];
	    }
	}
	export class CommandRequestDTO {
	    agentId: string;
	    command: string;
	    arguments: string;
	
	    static createFrom(source: any = {}) {
	        return new CommandRequestDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.agentId = source["agentId"];
	        this.command = source["command"];
	        this.arguments = source["arguments"];
	    }
	}
	export class CommandResponseDTO {
	    success: boolean;
	    output: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new CommandResponseDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.output = source["output"];
	        this.error = source["error"];
	    }
	}
	export class ConnectionStatusDTO {
	    connected: boolean;
	    serverUrl: string;
	    lastPing: string;
	    latency: number;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new ConnectionStatusDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.connected = source["connected"];
	        this.serverUrl = source["serverUrl"];
	        this.lastPing = source["lastPing"];
	        this.latency = source["latency"];
	        this.error = source["error"];
	    }
	}
	export class ServerMessageDTO {
	    type: string;
	    timestamp: string;
	    payload: any;
	
	    static createFrom(source: any = {}) {
	        return new ServerMessageDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.timestamp = source["timestamp"];
	        this.payload = source["payload"];
	    }
	}

}

