<script lang="ts">
  interface Agent {
    name: string;
    role: string;
    running: boolean;
    unread_mail: number;
  }

  interface ChatMessage {
    id: string;
    from: string;
    role: string;
    timestamp: Date;
    content: string;
    type: 'message' | 'system' | 'error';
  }

  export let agent: Agent | null;

  let messages: ChatMessage[] = [];
  let messageInput = '';
  let terminalRef: HTMLDivElement;

  function sendMessage() {
    if (!messageInput.trim() || !agent) return;

    const newMessage: ChatMessage = {
      id: Date.now().toString(),
      from: 'You',
      role: 'user',
      timestamp: new Date(),
      content: messageInput,
      type: 'message',
    };

    messages = [...messages, newMessage];
    messageInput = '';

    // Scroll to bottom
    setTimeout(() => {
      if (terminalRef) {
        terminalRef.scrollTop = terminalRef.scrollHeight;
      }
    }, 0);

    // Simulate agent response
    setTimeout(() => {
      const response: ChatMessage = {
        id: (Date.now() + 1).toString(),
        from: agent.name,
        role: agent.role,
        timestamp: new Date(),
        content: `Received: "${newMessage.content}"`,
        type: 'message',
      };
      messages = [...messages, response];

      if (terminalRef) {
        terminalRef.scrollTop = terminalRef.scrollHeight;
      }
    }, 500);
  }

  function getRoleIcon(role: string): string {
    const icons: Record<string, string> = {
      mayor: '👨‍⚖️',
      deacon: '⛪',
      witness: '👀',
      refinery: '⚙️',
      polecat: '🐱',
      crew: '👷',
      user: '👤',
    };
    return icons[role] || '🤖';
  }

  function getRoleColor(role: string): string {
    const colors: Record<string, string> = {
      mayor: 'var(--color-mayor)',
      deacon: 'var(--color-deacon)',
      witness: 'var(--color-witness)',
      refinery: 'var(--color-refinery)',
      polecat: 'var(--color-polecat)',
      crew: 'var(--color-crew)',
      user: 'var(--color-primary)',
    };
    return colors[role] || 'var(--color-gray)';
  }

  function formatTime(date: Date): string {
    return date.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit', second: '2-digit' });
  }
</script>

{#if agent}
  <div class="chat-container">
    <!-- Header -->
    <div class="chat-header">
      <div class="header-agent">
        <span class="agent-icon" style="color: {getRoleColor(agent.role)}">
          {getRoleIcon(agent.role)}
        </span>
        <div class="agent-info">
          <div class="agent-name">{agent.name}</div>
          <div class="agent-status">
            <span class="status-dot" class:running={agent.running}></span>
            {agent.running ? 'Connected' : 'Offline'}
          </div>
        </div>
      </div>
      <div class="header-stats">
        <div class="stat">
          <span class="stat-label">Status</span>
          <span class="stat-value" class:online={agent.running}>
            {agent.running ? 'ACTIVE' : 'INACTIVE'}
          </span>
        </div>
      </div>
    </div>

    <!-- Terminal Output -->
    <div class="terminal-output" bind:this={terminalRef}>
      {#if messages.length === 0}
        <div class="welcome-message">
          <div class="welcome-title">⛽ GAS TOWN COMMAND INTERFACE</div>
          <div class="welcome-text">Connected to {agent.name}</div>
          <div class="welcome-command">> Ready for input</div>
        </div>
      {:else}
        {#each messages as message}
          <div class="message" class:message-from-user={message.role === 'user'}>
            <div class="message-header">
              <span class="message-icon" style="color: {getRoleColor(message.role)}">
                {getRoleIcon(message.role)}
              </span>
              <span class="message-from">{message.from}</span>
              <span class="message-time">{formatTime(message.timestamp)}</span>
            </div>
            <div class="message-content">
              <code>{message.content}</code>
            </div>
          </div>
        {/each}
      {/if}
    </div>

    <!-- Input -->
    <div class="message-input-container">
      <div class="input-prompt">> </div>
      <input
        type="text"
        class="message-input"
        placeholder="Enter command..."
        bind:value={messageInput}
        on:keydown={(e) => e.key === 'Enter' && sendMessage()}
      />
      <button class="send-button" on:click={sendMessage} disabled={!messageInput.trim()}>
        SEND
      </button>
    </div>
  </div>
{:else}
  <div class="no-agent-selected">
    <div class="no-agent-icon">👤</div>
    <p>No agent selected</p>
    <p class="text-muted">Select an agent from the sidebar to begin communication</p>
  </div>
{/if}

<style>
  .chat-container {
    display: flex;
    flex-direction: column;
    height: 100%;
    background-color: var(--color-bg);
    font-family: var(--font-mono);
  }

  .chat-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: var(--space-4);
    background: linear-gradient(90deg, var(--color-surface), rgba(168, 85, 247, 0.05));
    border-bottom: 2px solid var(--color-border);
  }

  .header-agent {
    display: flex;
    align-items: center;
    gap: var(--space-3);
  }

  .agent-icon {
    font-size: 1.75rem;
  }

  .agent-info {
    display: flex;
    flex-direction: column;
  }

  .agent-name {
    font-size: 1rem;
    font-weight: 700;
    letter-spacing: 0.05em;
  }

  .agent-status {
    font-size: 0.75rem;
    color: var(--color-text-muted);
    display: flex;
    align-items: center;
    gap: var(--space-1);
    margin-top: 2px;
  }

  .status-dot {
    display: inline-block;
    width: 0.35rem;
    height: 0.35rem;
    border-radius: 50%;
    background-color: var(--color-gray);
  }

  .status-dot.running {
    background-color: var(--color-success);
    box-shadow: 0 0 6px var(--color-success);
  }

  .header-stats {
    display: flex;
    gap: var(--space-4);
  }

  .stat {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 4px;
  }

  .stat-label {
    font-size: 0.65rem;
    color: var(--color-text-muted);
    text-transform: uppercase;
    letter-spacing: 0.1em;
  }

  .stat-value {
    font-size: 0.85rem;
    font-weight: 700;
    color: var(--color-primary);
    letter-spacing: 0.05em;
  }

  .stat-value.online {
    color: var(--color-success);
    text-shadow: 0 0 8px rgba(34, 197, 94, 0.3);
  }

  .terminal-output {
    flex: 1;
    overflow-y: auto;
    padding: var(--space-4);
    background-color: var(--color-bg);
    border-bottom: 1px solid var(--color-border);
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
  }

  .welcome-message {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: var(--space-2);
    color: var(--color-text-muted);
    padding: var(--space-8);
    text-align: center;
    height: 100%;
  }

  .welcome-title {
    font-size: 1.125rem;
    font-weight: 700;
    color: var(--color-primary);
    letter-spacing: 0.05em;
    text-shadow: 0 0 10px rgba(59, 130, 246, 0.3);
  }

  .welcome-text {
    font-size: 0.9rem;
    color: var(--color-text-muted);
  }

  .welcome-command {
    font-size: 0.85rem;
    color: var(--color-success);
    margin-top: var(--space-2);
  }

  .message {
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
    padding: var(--space-2) var(--space-3);
    background-color: rgba(255, 255, 255, 0.02);
    border-left: 2px solid var(--color-border);
    border-radius: 2px;
    animation: slideIn 0.3s ease-out;
  }

  .message-from-user {
    border-left-color: var(--color-primary);
    background-color: rgba(59, 130, 246, 0.05);
  }

  .message-header {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    font-size: 0.8rem;
  }

  .message-icon {
    font-size: 1rem;
  }

  .message-from {
    font-weight: 600;
    min-width: 80px;
  }

  .message-time {
    color: var(--color-text-muted);
    font-size: 0.7rem;
    margin-left: auto;
  }

  .message-content {
    margin-left: var(--space-4);
    font-size: 0.85rem;
    color: var(--color-text);
    word-break: break-word;
  }

  .message-content code {
    background-color: rgba(255, 255, 255, 0.05);
    padding: var(--space-1) var(--space-2);
    border-radius: 2px;
    display: block;
  }

  .message-input-container {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    padding: var(--space-4);
    background-color: var(--color-surface);
    border-top: 2px solid var(--color-border);
  }

  .input-prompt {
    font-weight: 700;
    color: var(--color-success);
    font-size: 0.9rem;
    flex-shrink: 0;
  }

  .message-input {
    flex: 1;
    padding: var(--space-2) var(--space-3);
    background-color: rgba(255, 255, 255, 0.05);
    border: 1px solid var(--color-border);
    border-radius: 2px;
    color: var(--color-text);
    font-family: var(--font-mono);
    font-size: 0.9rem;
  }

  .message-input:focus {
    outline: none;
    border-color: var(--color-primary);
    box-shadow: 0 0 8px rgba(59, 130, 246, 0.3);
  }

  .message-input::placeholder {
    color: var(--color-text-muted);
  }

  .send-button {
    padding: var(--space-2) var(--space-3);
    background-color: var(--color-primary);
    color: white;
    border: none;
    border-radius: 2px;
    font-family: var(--font-mono);
    font-size: 0.8rem;
    font-weight: 700;
    cursor: pointer;
    transition: all var(--transition-normal);
    letter-spacing: 0.05em;
    flex-shrink: 0;
  }

  .send-button:hover:not(:disabled) {
    background-color: color-mix(in srgb, var(--color-primary) 80%, white);
    box-shadow: 0 0 10px rgba(59, 130, 246, 0.4);
  }

  .send-button:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .no-agent-selected {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100%;
    gap: var(--space-4);
    color: var(--color-text-muted);
  }

  .no-agent-icon {
    font-size: 4rem;
    opacity: 0.3;
  }

  /* Scrollbar styling */
  .terminal-output::-webkit-scrollbar {
    width: 8px;
  }

  .terminal-output::-webkit-scrollbar-track {
    background: var(--color-bg);
  }

  .terminal-output::-webkit-scrollbar-thumb {
    background: var(--color-border);
    border-radius: 4px;
  }

  .terminal-output::-webkit-scrollbar-thumb:hover {
    background: var(--color-gray);
  }

  @keyframes slideIn {
    from {
      opacity: 0;
      transform: translateY(10px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }
</style>
