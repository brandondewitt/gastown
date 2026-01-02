// WebSocket client with auto-reconnect for Gas Town real-time updates

export interface WSMessage {
  type: string;
  timestamp: string;
  payload: unknown;
}

export interface AgentUpdate {
  address: string;
  running: boolean;
  has_work: boolean;
  hook_bead?: string;
  work_title?: string;
  state?: string;
  change_type: string;
}

// Topics for subscription
export const Topics = {
  ALL: 'all',
  EVENTS: 'events',
  STATUS: 'status',
  AGENTS: 'agents',
  CONVOYS: 'convoys',
  MQ: 'mq',
} as const;

// Message types from server
export const MessageTypes = {
  EVENT: 'event',
  STATUS_UPDATE: 'status_update',
  AGENT_UPDATE: 'agent_update',
  CONVOY_UPDATE: 'convoy_update',
  MQ_UPDATE: 'mq_update',
  PING: 'ping',
  PONG: 'pong',
} as const;

export type MessageHandler = (message: WSMessage) => void;

interface WebSocketClientOptions {
  url?: string;
  reconnectInterval?: number;
  maxReconnectInterval?: number;
  reconnectDecay?: number;
  maxReconnectAttempts?: number;
  topics?: string[];
}

export class WebSocketClient {
  private ws: WebSocket | null = null;
  private url: string;
  private reconnectInterval: number;
  private maxReconnectInterval: number;
  private reconnectDecay: number;
  private maxReconnectAttempts: number;
  private reconnectAttempts: number = 0;
  private shouldReconnect: boolean = true;
  private handlers: Map<string, Set<MessageHandler>> = new Map();
  private topics: string[];
  private reconnectTimeoutId: ReturnType<typeof setTimeout> | null = null;

  constructor(options: WebSocketClientOptions = {}) {
    // Default to current host for WebSocket URL
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const defaultUrl = `${protocol}//${window.location.host}/api/v1/ws`;

    this.url = options.url || defaultUrl;
    this.reconnectInterval = options.reconnectInterval || 1000;
    this.maxReconnectInterval = options.maxReconnectInterval || 30000;
    this.reconnectDecay = options.reconnectDecay || 1.5;
    this.maxReconnectAttempts = options.maxReconnectAttempts || 0; // 0 = unlimited
    this.topics = options.topics || [Topics.ALL];
  }

  connect(): void {
    if (this.ws && (this.ws.readyState === WebSocket.CONNECTING || this.ws.readyState === WebSocket.OPEN)) {
      return;
    }

    try {
      this.ws = new WebSocket(this.url);
      this.setupEventHandlers();
    } catch (error) {
      console.error('[WS] Connection error:', error);
      this.scheduleReconnect();
    }
  }

  private setupEventHandlers(): void {
    if (!this.ws) return;

    this.ws.onopen = () => {
      console.log('[WS] Connected');
      this.reconnectAttempts = 0;
      this.reconnectInterval = 1000;

      // Subscribe to topics
      this.subscribe(this.topics);

      // Notify connection handlers
      this.emit('_connected', { type: '_connected', timestamp: new Date().toISOString(), payload: null });
    };

    this.ws.onclose = (event) => {
      console.log(`[WS] Disconnected (code: ${event.code}, reason: ${event.reason})`);
      this.ws = null;

      // Notify disconnection handlers
      this.emit('_disconnected', { type: '_disconnected', timestamp: new Date().toISOString(), payload: { code: event.code, reason: event.reason } });

      if (this.shouldReconnect) {
        this.scheduleReconnect();
      }
    };

    this.ws.onerror = (error) => {
      console.error('[WS] Error:', error);
      this.emit('_error', { type: '_error', timestamp: new Date().toISOString(), payload: error });
    };

    this.ws.onmessage = (event) => {
      try {
        // Handle multiple messages in one frame (newline-separated)
        const messages = event.data.split('\n').filter((s: string) => s.trim());
        for (const msgStr of messages) {
          const message: WSMessage = JSON.parse(msgStr);
          this.handleMessage(message);
        }
      } catch (error) {
        console.error('[WS] Failed to parse message:', error, event.data);
      }
    };
  }

  private handleMessage(message: WSMessage): void {
    // Emit to type-specific handlers
    this.emit(message.type, message);

    // Emit to wildcard handlers
    this.emit('*', message);
  }

  private emit(type: string, message: WSMessage): void {
    const handlers = this.handlers.get(type);
    if (handlers) {
      handlers.forEach(handler => {
        try {
          handler(message);
        } catch (error) {
          console.error(`[WS] Handler error for ${type}:`, error);
        }
      });
    }
  }

  private scheduleReconnect(): void {
    if (!this.shouldReconnect) return;

    if (this.maxReconnectAttempts > 0 && this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.log('[WS] Max reconnect attempts reached');
      this.emit('_max_reconnect', { type: '_max_reconnect', timestamp: new Date().toISOString(), payload: null });
      return;
    }

    const delay = Math.min(
      this.reconnectInterval * Math.pow(this.reconnectDecay, this.reconnectAttempts),
      this.maxReconnectInterval
    );

    console.log(`[WS] Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts + 1})`);

    this.reconnectTimeoutId = setTimeout(() => {
      this.reconnectAttempts++;
      this.connect();
    }, delay);
  }

  subscribe(topics: string[]): void {
    this.topics = topics;
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({ topics }));
    }
  }

  on(type: string, handler: MessageHandler): () => void {
    if (!this.handlers.has(type)) {
      this.handlers.set(type, new Set());
    }
    this.handlers.get(type)!.add(handler);

    // Return unsubscribe function
    return () => {
      this.handlers.get(type)?.delete(handler);
    };
  }

  off(type: string, handler: MessageHandler): void {
    this.handlers.get(type)?.delete(handler);
  }

  disconnect(): void {
    this.shouldReconnect = false;
    if (this.reconnectTimeoutId) {
      clearTimeout(this.reconnectTimeoutId);
      this.reconnectTimeoutId = null;
    }
    if (this.ws) {
      this.ws.close(1000, 'Client disconnect');
      this.ws = null;
    }
  }

  get isConnected(): boolean {
    return this.ws !== null && this.ws.readyState === WebSocket.OPEN;
  }

  get readyState(): number {
    return this.ws?.readyState ?? WebSocket.CLOSED;
  }
}

// Singleton instance for the app
let clientInstance: WebSocketClient | null = null;

export function getWebSocketClient(options?: WebSocketClientOptions): WebSocketClient {
  if (!clientInstance) {
    clientInstance = new WebSocketClient(options);
  }
  return clientInstance;
}

export function destroyWebSocketClient(): void {
  if (clientInstance) {
    clientInstance.disconnect();
    clientInstance = null;
  }
}
