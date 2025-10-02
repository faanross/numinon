/**
 * WebSocket Message Types
 * These define the protocol between our client and the C2 server
 */

export enum WSMessageType {
    // Connection Management
    CONNECT = 'connect',
    DISCONNECT = 'disconnect',
    PING = 'ping',
    PONG = 'pong',

    // Authentication
    AUTH_REQUEST = 'auth:request',
    AUTH_RESPONSE = 'auth:response',
    AUTH_ERROR = 'auth:error',

    // Agent Operations
    AGENT_LIST = 'agent:list',
    AGENT_CONNECTED = 'agent:connected',
    AGENT_DISCONNECTED = 'agent:disconnected',
    AGENT_UPDATE = 'agent:update',

    // Task Operations
    TASK_EXECUTE = 'task:execute',
    TASK_RESULT = 'task:result',
    TASK_STATUS = 'task:status',
    TASK_CANCEL = 'task:cancel',

    // Listener Operations
    LISTENER_CREATE = 'listener:create',
    LISTENER_DELETE = 'listener:delete',
    LISTENER_UPDATE = 'listener:update',
    LISTENER_LIST = 'listener:list',

    // System
    ERROR = 'error',
    NOTIFICATION = 'notification'
}

/**
 * Base WebSocket Message Structure
 */
export interface WSMessage<T = any> {
    id: string;           // Unique message ID for request/response matching
    type: WSMessageType;  // Message type
    timestamp: string;    // ISO 8601 timestamp
    payload?: T;          // Type-specific payload
}

/**
 * WebSocket Connection State
 */
export enum WSState {
    CONNECTING = 'connecting',
    CONNECTED = 'connected',
    AUTHENTICATED = 'authenticated',
    DISCONNECTING = 'disconnecting',
    DISCONNECTED = 'disconnected',
    RECONNECTING = 'reconnecting',
    ERROR = 'error'
}

/**
 * Connection Configuration
 */
export interface WSConfig {
    url: string;
    reconnect: boolean;
    reconnectDelay: number;      // milliseconds
    maxReconnectAttempts: number;
    pingInterval: number;         // milliseconds
    pongTimeout: number;          // milliseconds
    authToken?: string;
}

/**
 * Authentication Messages
 */
export interface AuthRequest {
    token?: string;
    username?: string;
    password?: string;
}

export interface AuthResponse {
    success: boolean;
    token?: string;
    user?: {
        id: string;
        username: string;
        role: string;
        permissions: string[];
    };
    error?: string;
}

/**
 * Agent-related Message Payloads
 */
export interface AgentListPayload {
    agents: Array<{
        id: string;
        hostname: string;
        username: string;
        os: string;
        arch: string;
        ip: string;
        lastSeen: string;
        status: 'online' | 'offline' | 'stale';
        listener: string;
    }>;
    total: number;
}

export interface AgentUpdatePayload {
    agentId: string;
    changes: Partial<{
        status: string;
        lastSeen: string;
        ip: string;
        hostname: string;
    }>;
}

/**
 * Task Execution Payloads
 */
export interface TaskExecutePayload {
    agentId: string;
    taskType: string;  // 'cmd', 'upload', 'download', 'shellcode', etc.
    parameters: Record<string, any>;
    timeout?: number;  // milliseconds
}

export interface TaskResultPayload {
    taskId: string;
    agentId: string;
    success: boolean;
    output?: string;
    error?: string;
    executionTime: number; // milliseconds
}

/**
 * WebSocket Event Handlers
 */
export interface WSEventHandlers {
    onOpen?: () => void;
    onClose?: (event: CloseEvent) => void;
    onError?: (error: Error) => void;
    onMessage?: (message: WSMessage) => void;
    onStateChange?: (state: WSState) => void;
    onReconnect?: (attempt: number) => void;
}

/**
 * Message Queue for offline/reconnection scenarios
 */
export interface QueuedMessage {
    message: WSMessage;
    timestamp: number;
    retries: number;
    maxRetries: number;
    onSuccess?: (response: WSMessage) => void;
    onError?: (error: Error) => void;
}