<script lang="ts">
  interface Agent {
    name: string;
    address: string;
    session: string;
    role: string;
    running: boolean;
    has_work: boolean;
    work_title?: string;
    hook_bead?: string;
    state?: string;
    unread_mail: number;
  }

  interface MailMessage {
    id: string;
    from: string;
    to: string;
    subject: string;
    body: string;
    timestamp: string;
    read: boolean;
    priority: string;
  }

  let {
    agent,
    onDispatchWork = () => {},
    onSendMail = () => {},
    onClose = () => {}
  } = $props<{
    agent: Agent;
    onDispatchWork?: () => void;
    onSendMail?: () => void;
    onClose?: () => void;
  }>();

  let mailMessages: MailMessage[] = $state([]);
  let mailLoading = $state(false);
  let mailExpanded = $state(false);
  let workExpanded = $state(true);

  async function fetchAgentMail() {
    if (!agent.address) return;

    mailLoading = true;
    try {
      // Fetch mail for this specific agent
      const res = await fetch(`/api/v1/mail?to=${encodeURIComponent(agent.address)}`);
      const data = await res.json();
      if (data.success) {
        mailMessages = (data.data || []).slice(0, 5); // Show latest 5
      }
    } catch (e) {
      console.error('Failed to fetch agent mail:', e);
    } finally {
      mailLoading = false;
    }
  }

  function formatTimestamp(timestamp: string): string {
    const date = new Date(timestamp);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMs / 3600000);

    if (diffMins < 1) return 'just now';
    if (diffMins < 60) return `${diffMins}m ago`;
    if (diffHours < 24) return `${diffHours}h ago`;
    return date.toLocaleDateString();
  }

  // Fetch mail when agent changes
  $effect(() => {
    if (agent.address && mailExpanded) {
      fetchAgentMail();
    }
  });
</script>

<aside class="agent-context">
  <div class="context-header">
    <h3>Context</h3>
    <button class="close-btn" onclick={onClose} title="Close panel">×</button>
  </div>

  <!-- Current Work Section -->
  <section class="context-section">
    <button class="section-header" onclick={() => workExpanded = !workExpanded}>
      <span class="section-icon">📋</span>
      <span class="section-title">Current Work</span>
      <span class="expand-icon">{workExpanded ? '▼' : '▶'}</span>
    </button>

    {#if workExpanded}
      <div class="section-content">
        {#if agent.has_work && agent.hook_bead}
          <div class="work-card">
            <div class="work-id">{agent.hook_bead}</div>
            {#if agent.work_title}
              <div class="work-title">{agent.work_title}</div>
            {/if}
            <div class="work-status">
              <span class="status-indicator active"></span>
              <span>In Progress</span>
            </div>
          </div>
        {:else}
          <div class="empty-state">
            <span class="empty-icon">💤</span>
            <span class="empty-text">No active work</span>
          </div>
        {/if}
      </div>
    {/if}
  </section>

  <!-- Mail Section -->
  <section class="context-section">
    <button class="section-header" onclick={() => { mailExpanded = !mailExpanded; if (mailExpanded) fetchAgentMail(); }}>
      <span class="section-icon">📧</span>
      <span class="section-title">Mail</span>
      {#if agent.unread_mail > 0}
        <span class="badge">{agent.unread_mail}</span>
      {/if}
      <span class="expand-icon">{mailExpanded ? '▼' : '▶'}</span>
    </button>

    {#if mailExpanded}
      <div class="section-content">
        {#if mailLoading}
          <div class="loading-state">Loading...</div>
        {:else if mailMessages.length === 0}
          <div class="empty-state">
            <span class="empty-icon">📭</span>
            <span class="empty-text">No messages</span>
          </div>
        {:else}
          <div class="mail-list">
            {#each mailMessages as msg}
              <div class="mail-item" class:unread={!msg.read}>
                <div class="mail-from">{msg.from}</div>
                <div class="mail-subject">{msg.subject}</div>
                <div class="mail-time">{formatTimestamp(msg.timestamp)}</div>
              </div>
            {/each}
          </div>
        {/if}
      </div>
    {/if}
  </section>

  <!-- Agent Info Section -->
  <section class="context-section">
    <div class="section-header static">
      <span class="section-icon">ℹ️</span>
      <span class="section-title">Info</span>
    </div>
    <div class="section-content">
      <div class="info-grid">
        <div class="info-item">
          <span class="info-label">Status</span>
          <span class="info-value">
            <span class="status-dot" class:active={agent.running}></span>
            {agent.running ? 'Running' : 'Stopped'}
          </span>
        </div>
        <div class="info-item">
          <span class="info-label">Role</span>
          <span class="info-value">{agent.role}</span>
        </div>
        <div class="info-item">
          <span class="info-label">Session</span>
          <span class="info-value mono">{agent.session}</span>
        </div>
      </div>
    </div>
  </section>

  <!-- Quick Actions -->
  <section class="context-section actions-section">
    <div class="section-header static">
      <span class="section-icon">⚡</span>
      <span class="section-title">Actions</span>
    </div>
    <div class="section-content">
      <div class="action-buttons">
        <button class="action-btn" onclick={onDispatchWork}>
          <span class="action-icon">🚀</span>
          <span>Dispatch Work</span>
        </button>
        <button class="action-btn" onclick={onSendMail}>
          <span class="action-icon">✉️</span>
          <span>Send Mail</span>
        </button>
      </div>
    </div>
  </section>
</aside>

<style>
  .agent-context {
    width: 280px;
    min-width: 280px;
    background: var(--color-surface);
    border-left: 1px solid var(--color-border);
    height: 100%;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
  }

  .context-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: var(--space-3) var(--space-4);
    border-bottom: 1px solid var(--color-border);
  }

  .context-header h3 {
    margin: 0;
    font-size: 0.875rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: var(--color-text-muted);
  }

  .close-btn {
    display: none;
    align-items: center;
    justify-content: center;
    width: 1.5rem;
    height: 1.5rem;
    background: none;
    border: none;
    color: var(--color-text-muted);
    font-size: 1.25rem;
    cursor: pointer;
    border-radius: var(--radius-sm);
  }

  .close-btn:hover {
    background: var(--color-surface-raised);
    color: var(--color-text);
  }

  @media (max-width: 1024px) {
    .close-btn {
      display: flex;
    }
  }

  .context-section {
    border-bottom: 1px solid var(--color-border);
  }

  .section-header {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    width: 100%;
    padding: var(--space-3) var(--space-4);
    background: none;
    border: none;
    text-align: left;
    cursor: pointer;
    color: var(--color-text);
    transition: background-color var(--transition-fast);
  }

  .section-header:hover {
    background: var(--color-surface-raised);
  }

  .section-header.static {
    cursor: default;
  }

  .section-header.static:hover {
    background: none;
  }

  .section-icon {
    font-size: 0.875rem;
  }

  .section-title {
    flex: 1;
    font-size: 0.8125rem;
    font-weight: 600;
  }

  .expand-icon {
    font-size: 0.625rem;
    color: var(--color-text-muted);
  }

  .badge {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-width: 1.25rem;
    height: 1.25rem;
    padding: 0 var(--space-1);
    font-size: 0.6875rem;
    font-weight: 600;
    background: var(--color-error);
    color: white;
    border-radius: 9999px;
  }

  .section-content {
    padding: 0 var(--space-4) var(--space-3);
  }

  /* Work Card */
  .work-card {
    padding: var(--space-3);
    background: var(--color-surface-raised);
    border-radius: var(--radius-md);
    border-left: 3px solid var(--color-primary);
  }

  .work-id {
    font-family: var(--font-mono);
    font-size: 0.75rem;
    color: var(--color-primary);
    margin-bottom: var(--space-1);
  }

  .work-title {
    font-size: 0.8125rem;
    color: var(--color-text);
    margin-bottom: var(--space-2);
  }

  .work-status {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    font-size: 0.75rem;
    color: var(--color-text-muted);
  }

  .status-indicator {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: var(--color-text-muted);
  }

  .status-indicator.active {
    background: var(--color-success);
    animation: pulse 2s ease-in-out infinite;
  }

  @keyframes pulse {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.5; }
  }

  /* Mail List */
  .mail-list {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }

  .mail-item {
    padding: var(--space-2);
    background: var(--color-surface-raised);
    border-radius: var(--radius-sm);
    font-size: 0.75rem;
  }

  .mail-item.unread {
    border-left: 2px solid var(--color-primary);
  }

  .mail-from {
    font-weight: 500;
    color: var(--color-text);
    margin-bottom: 2px;
  }

  .mail-subject {
    color: var(--color-text-muted);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .mail-time {
    color: var(--color-text-muted);
    font-size: 0.6875rem;
    margin-top: 2px;
  }

  /* Info Grid */
  .info-grid {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }

  .info-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    font-size: 0.75rem;
  }

  .info-label {
    color: var(--color-text-muted);
  }

  .info-value {
    display: flex;
    align-items: center;
    gap: var(--space-1);
    color: var(--color-text);
  }

  .info-value.mono {
    font-family: var(--font-mono);
    font-size: 0.6875rem;
  }

  .status-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: var(--color-text-muted);
  }

  .status-dot.active {
    background: var(--color-success);
  }

  /* Action Buttons */
  .action-buttons {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }

  .action-btn {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    width: 100%;
    padding: var(--space-2) var(--space-3);
    background: var(--color-surface-raised);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-md);
    color: var(--color-text);
    font-size: 0.8125rem;
    cursor: pointer;
    transition: all var(--transition-fast);
  }

  .action-btn:hover {
    background: var(--color-primary);
    border-color: var(--color-primary);
    color: white;
  }

  .action-icon {
    font-size: 0.875rem;
  }

  /* Empty & Loading States */
  .empty-state, .loading-state {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: var(--space-2);
    padding: var(--space-4);
    color: var(--color-text-muted);
    font-size: 0.75rem;
  }

  .empty-icon {
    font-size: 1rem;
  }

  /* Responsive */
  @media (max-width: 1024px) {
    .agent-context {
      position: fixed;
      right: 0;
      top: 0;
      bottom: 0;
      z-index: 100;
      box-shadow: -4px 0 12px rgba(0, 0, 0, 0.15);
    }

    .section-header {
      min-height: 44px; /* Touch-friendly target */
    }

    .action-btn {
      min-height: 44px;
      padding: var(--space-3);
    }
  }

  @media (max-width: 768px) {
    .agent-context {
      width: 100%;
      max-width: 320px;
    }

    .close-btn {
      width: 2rem;
      height: 2rem;
      font-size: 1.5rem;
    }
  }
</style>
