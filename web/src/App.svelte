<script lang="ts">
  import { onMount } from 'svelte';
  import './app.css';
  import MergeQueue from './lib/MergeQueue.svelte';
  import ActivityFeed from './lib/ActivityFeed.svelte';
  import SlingDialog from './lib/SlingDialog.svelte';
  import ConvoyCreator from './lib/ConvoyCreator.svelte';
  import IssueCreator from './lib/IssueCreator.svelte';
  import IssuesList from './lib/IssuesList.svelte';
  import MailComposer from './lib/MailComposer.svelte';
  import AgentSidebar from './lib/AgentSidebar.svelte';
  import TerminalOutput from './lib/TerminalOutput.svelte';
  import MessageInput from './lib/MessageInput.svelte';

  // Types
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

  interface Rig {
    name: string;
    path: string;
    polecats: string[];
    polecat_count: number;
    crews: string[];
    crew_count: number;
    has_witness: boolean;
    has_refinery: boolean;
    agents?: Agent[];
  }

  interface Summary {
    rig_count: number;
    polecat_count: number;
    crew_count: number;
    witness_count: number;
    refinery_count: number;
    active_hooks: number;
  }

  interface Overseer {
    name: string;
    email?: string;
    unread_mail: number;
  }

  interface TownStatus {
    name: string;
    location: string;
    overseer?: Overseer;
    agents: Agent[];
    rigs: Rig[];
    summary: Summary;
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
    type: string;
    thread_id?: string;
    reply_to?: string;
    pinned?: boolean;
    cc?: string[];
  }

  // State
  let status: TownStatus | null = $state(null);
  let loading = $state(true);
  let error: string | null = $state(null);
  let activeTab: 'dashboard' | 'issues' | 'mail' = $state('dashboard');
  let theme: 'dark' | 'light' = $state('dark');
  let selectedAgent: Agent | null = $state(null);
  let sidebarOpen = $state(true);

  // Mail state
  let mailMessages: MailMessage[] = $state([]);
  let mailLoading = $state(false);
  let mailError: string | null = $state(null);
  let selectedMessage: MailMessage | null = $state(null);
  let mailCount = $state({ total: 0, unread: 0 });

  // Dialog state
  let showCreateMenu = $state(false);
  let showSlingDialog = $state(false);
  let showConvoyCreator = $state(false);
  let showIssueCreator = $state(false);
  let slingIssueId = $state('');
  let showMailComposer = $state(false);

  // Theme handling
  function initTheme() {
    const saved = localStorage.getItem('gt-theme');
    if (saved === 'light' || saved === 'dark') {
      theme = saved;
    } else if (window.matchMedia('(prefers-color-scheme: light)').matches) {
      theme = 'light';
    }
    applyTheme();
  }

  function toggleTheme() {
    theme = theme === 'dark' ? 'light' : 'dark';
    localStorage.setItem('gt-theme', theme);
    applyTheme();
  }

  function applyTheme() {
    document.documentElement.setAttribute('data-theme', theme);
  }

  async function fetchStatus() {
    try {
      const response = await fetch('/api/v1/status');
      const data = await response.json();
      if (data.success) {
        status = data.data;
        error = null;
      } else {
        error = data.error?.message || 'Unknown error';
      }
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to fetch status';
    } finally {
      loading = false;
    }
  }

  async function fetchMail() {
    mailLoading = true;
    mailError = null;
    try {
      const [messagesRes, countRes] = await Promise.all([
        fetch('/api/v1/mail'),
        fetch('/api/v1/mail/count')
      ]);

      const messagesData = await messagesRes.json();
      const countData = await countRes.json();

      if (messagesData.success) {
        mailMessages = messagesData.data || [];
      } else {
        mailError = messagesData.error?.message || 'Failed to load messages';
      }

      if (countData.success) {
        mailCount = countData.data;
      }
    } catch (e) {
      mailError = e instanceof Error ? e.message : 'Failed to fetch mail';
    } finally {
      mailLoading = false;
    }
  }

  async function markAsRead(id: string) {
    try {
      await fetch(`/api/v1/mail/${id}/read`, { method: 'POST' });
      // Refresh mail list
      await fetchMail();
    } catch (e) {
      console.error('Failed to mark as read:', e);
    }
  }

  function selectMessage(msg: MailMessage) {
    selectedMessage = msg;
    if (!msg.read) {
      markAsRead(msg.id);
    }
  }

  function formatDate(timestamp: string): string {
    const date = new Date(timestamp);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMs / 3600000);
    const diffDays = Math.floor(diffMs / 86400000);

    if (diffMins < 1) return 'just now';
    if (diffMins < 60) return `${diffMins}m ago`;
    if (diffHours < 24) return `${diffHours}h ago`;
    if (diffDays < 7) return `${diffDays}d ago`;
    return date.toLocaleDateString();
  }

  function getPriorityClass(priority: string): string {
    const classes: Record<string, string> = {
      urgent: 'priority-urgent',
      high: 'priority-high',
      normal: 'priority-normal',
      low: 'priority-low',
    };
    return classes[priority] || 'priority-normal';
  }

  onMount(() => {
    initTheme();
    fetchStatus();
    fetchMail();

    // Refresh status every 5 seconds
    const statusInterval = setInterval(fetchStatus, 5000);
    // Refresh mail every 30 seconds
    const mailInterval = setInterval(fetchMail, 30000);

    return () => {
      clearInterval(statusInterval);
      clearInterval(mailInterval);
    };
  });

  function getRoleIcon(role: string): string {
    const icons: Record<string, string> = {
      mayor: '👨‍⚖️',
      deacon: '⛪',
      witness: '👀',
      refinery: '⚙️',
      polecat: '🐱',
      crew: '👷',
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
    };
    return colors[role] || 'var(--color-gray)';
  }

  function handleDispatchFromIssues(issueId: string) {
    slingIssueId = issueId;
    showSlingDialog = true;
  }

  function handleAgentSelect(agent: Agent) {
    selectedAgent = agent;
    // Clear tab selection when agent is selected
    activeTab = 'dashboard';
  }

  function toggleSidebar() {
    sidebarOpen = !sidebarOpen;
  }
</script>

<div class="app">
  <header class="header">
    <div class="header-left">
      <button class="sidebar-toggle" onclick={toggleSidebar} title="Toggle sidebar">
        ☰
      </button>
      <div class="logo">
        <span class="logo-icon">⛽</span>
        <h1>Gas Town</h1>
      </div>
      <nav class="nav-tabs">
        <button
          class="nav-tab"
          class:active={activeTab === 'dashboard' && !selectedAgent}
          onclick={() => { activeTab = 'dashboard'; selectedAgent = null; }}
        >
          Dashboard
        </button>
        <button
          class="nav-tab"
          class:active={activeTab === 'issues'}
          onclick={() => { activeTab = 'issues'; selectedAgent = null; }}
        >
          Issues
        </button>
        <button
          class="nav-tab"
          class:active={activeTab === 'mail'}
          onclick={() => { activeTab = 'mail'; selectedAgent = null; fetchMail(); }}
        >
          Mail
          {#if mailCount.unread > 0}
            <span class="mail-badge">{mailCount.unread}</span>
          {/if}
        </button>
      </nav>
    </div>
    <div class="header-right">
      <div class="create-dropdown">
        <button class="btn btn-primary create-btn" onclick={() => showCreateMenu = !showCreateMenu}>
          + Create
        </button>
        {#if showCreateMenu}
          <div class="create-menu">
            <button onclick={() => { showCreateMenu = false; showIssueCreator = true; }}>
              📝 New Issue
            </button>
            <button onclick={() => { showCreateMenu = false; slingIssueId = ''; showSlingDialog = true; }}>
              🚀 Dispatch Work
            </button>
            <button onclick={() => { showCreateMenu = false; showConvoyCreator = true; }}>
              📦 New Convoy
            </button>
          </div>
        {/if}
      </div>
      {#if status}
        <span class="town-name">{status.name}</span>
      {/if}
      <button class="theme-toggle" onclick={toggleTheme} title="Toggle theme">
        {#if theme === 'dark'}
          <span class="theme-icon">☀️</span>
        {:else}
          <span class="theme-icon">🌙</span>
        {/if}
      </button>
    </div>
  </header>

  <div class="app-layout">
    <!-- Agent Sidebar -->
    {#if sidebarOpen}
      <AgentSidebar
        globalAgents={status?.agents || []}
        rigs={status?.rigs || []}
        {selectedAgent}
        onSelect={handleAgentSelect}
      />
    {/if}

    <main class="main">
      <!-- Selected Agent View -->
      {#if selectedAgent}
        <div class="agent-workspace-placeholder">
          <div class="agent-header">
            <h2>{getRoleIcon(selectedAgent.role)} {selectedAgent.name}</h2>
            <span class="agent-role">{selectedAgent.role}</span>
            <span class="status-badge" class:running={selectedAgent.running}>
              {selectedAgent.running ? 'Running' : 'Stopped'}
            </span>
          </div>
          {#if selectedAgent.has_work && selectedAgent.work_title}
            <div class="current-work-banner">
              <span class="work-label">Working on:</span>
              <span class="work-id">{selectedAgent.hook_bead}</span>
              <span class="work-title-text">{selectedAgent.work_title}</span>
            </div>
          {/if}
          <div class="agent-content">
            <!-- Terminal Output -->
            <TerminalOutput agentAddress={selectedAgent.address} />

            <!-- Message Input -->
            <MessageInput
              agentAddress={selectedAgent.address}
              disabled={!selectedAgent.running}
              placeholder="Send a message to {selectedAgent.name}..."
            />

            <!-- Agent Stats -->
            <div class="agent-stats">
              <div class="stat">
                <span class="stat-label">Address</span>
                <span class="stat-value mono">{selectedAgent.address}</span>
              </div>
              <div class="stat">
                <span class="stat-label">Session</span>
                <span class="stat-value mono">{selectedAgent.session}</span>
              </div>
              <div class="stat">
                <span class="stat-label">Unread Mail</span>
                <span class="stat-value">{selectedAgent.unread_mail}</span>
              </div>
            </div>
          </div>
        </div>
      {:else if activeTab === 'dashboard'}
        {#if loading}
        <div class="loading">
          <div class="spinner"></div>
          <p>Loading town status...</p>
        </div>
      {:else if error}
        <div class="error-container">
          <div class="error-icon">⚠️</div>
          <p class="error-text">{error}</p>
          <button class="btn btn-primary" onclick={fetchStatus}>Retry</button>
        </div>
      {:else if status}
        <!-- Summary Cards -->
        <section class="summary-grid">
          <div class="card summary-card">
            <div class="summary-value">{status.summary.rig_count}</div>
            <div class="summary-label">Rigs</div>
          </div>
          <div class="card summary-card">
            <div class="summary-value">{status.summary.polecat_count}</div>
            <div class="summary-label">Polecats</div>
          </div>
          <div class="card summary-card">
            <div class="summary-value">{status.summary.crew_count}</div>
            <div class="summary-label">Crews</div>
          </div>
          <div class="card summary-card">
            <div class="summary-value">{status.summary.active_hooks}</div>
            <div class="summary-label">Active Hooks</div>
          </div>
        </section>

        <!-- Global Agents -->
        {#if status.agents?.length > 0}
          <section class="card">
            <h2 class="card-header">Global Agents</h2>
            <div class="agent-list">
              {#each status.agents as agent}
                <div class="agent-item">
                  <span class="agent-icon" style="color: {getRoleColor(agent.role)}">
                    {getRoleIcon(agent.role)}
                  </span>
                  <span class="agent-name">{agent.name}</span>
                  <span class="status-dot" class:running={agent.running} class:stopped={!agent.running}></span>
                  {#if agent.unread_mail > 0}
                    <span class="mail-badge small">{agent.unread_mail}</span>
                  {/if}
                </div>
              {/each}
            </div>
          </section>
        {/if}

        <!-- Rigs -->
        <section class="rigs-section">
          <h2 class="section-header">Rigs</h2>
          <div class="rigs-grid">
            {#each status.rigs ?? [] as rig}
              <div class="card rig-card">
                <div class="rig-header">
                  <h3 class="rig-name">{rig.name}</h3>
                  <div class="rig-services">
                    {#if rig.has_witness}
                      <span class="service-badge" title="Witness">👀</span>
                    {/if}
                    {#if rig.has_refinery}
                      <span class="service-badge" title="Refinery">⚙️</span>
                    {/if}
                  </div>
                </div>
                <p class="rig-path mono text-muted">{rig.path}</p>
                <div class="rig-stats">
                  <span>{rig.polecat_count} polecats</span>
                  <span>{rig.crew_count} crews</span>
                </div>
                {#if rig.agents && rig.agents.length > 0}
                  <div class="rig-agents">
                    {#each rig.agents as agent}
                      <div class="agent-row">
                        <span class="agent-icon small" style="color: {getRoleColor(agent.role)}">
                          {getRoleIcon(agent.role)}
                        </span>
                        <span class="agent-name">{agent.name}</span>
                        <span class="status-dot" class:running={agent.running} class:working={agent.has_work}></span>
                        {#if agent.has_work && agent.work_title}
                          <span class="work-title mono">{agent.work_title}</span>
                        {/if}
                      </div>
                    {/each}
                  </div>
                {/if}
              </div>
            {/each}
          </div>
        </section>

        <!-- Merge Queue Section -->
        <section class="card">
          <h2 class="card-header">Merge Queue</h2>
          <MergeQueue />
        </section>

        <!-- Activity Feed Section -->
        <section class="card">
          <h2 class="card-header">Recent Activity</h2>
          <ActivityFeed />
        </section>
      {/if}
    {:else if activeTab === 'issues'}
      <div class="issues-view">
        <IssuesList onDispatch={handleDispatchFromIssues} />
      </div>
    {:else if activeTab === 'mail'}
      <div class="mail-container">
        <div class="mail-list" class:has-selected={selectedMessage}>
          <div class="mail-list-header">
            <h2>Inbox</h2>
            <div class="mail-list-actions">
              <button class="btn btn-primary btn-sm" onclick={() => showMailComposer = true}>
                + Compose
              </button>
              <button class="btn btn-icon" onclick={fetchMail} title="Refresh">
                🔄
              </button>
            </div>
          </div>
          {#if mailLoading && mailMessages.length === 0}
            <div class="loading-inline">
              <div class="spinner small"></div>
              <span>Loading messages...</span>
            </div>
          {:else if mailError}
            <div class="error-inline">
              <span>⚠️ {mailError}</span>
              <button class="btn btn-sm" onclick={fetchMail}>Retry</button>
            </div>
          {:else if mailMessages.length === 0}
            <div class="empty-state">
              <span class="empty-icon">📭</span>
              <p>No messages</p>
            </div>
          {:else}
            <div class="mail-items">
              {#each mailMessages as msg}
                <button
                  class="mail-item"
                  class:unread={!msg.read}
                  class:selected={selectedMessage?.id === msg.id}
                  onclick={() => selectMessage(msg)}
                >
                  <div class="mail-item-header">
                    <span class="mail-from">{msg.from}</span>
                    <span class="mail-time">{formatDate(msg.timestamp)}</span>
                  </div>
                  <div class="mail-subject">
                    {#if msg.pinned}
                      <span class="pin-icon">📌</span>
                    {/if}
                    <span class={getPriorityClass(msg.priority)}>{msg.subject}</span>
                  </div>
                  <div class="mail-preview">{msg.body.slice(0, 80)}{msg.body.length > 80 ? '...' : ''}</div>
                </button>
              {/each}
            </div>
          {/if}
        </div>

        <div class="mail-detail" class:visible={selectedMessage}>
          {#if selectedMessage}
            <div class="mail-detail-header">
              <button class="btn-back" onclick={() => selectedMessage = null}>
                ← Back
              </button>
              <div class="mail-detail-meta">
                <span class="badge {getPriorityClass(selectedMessage.priority)}">{selectedMessage.priority}</span>
                <span class="badge">{selectedMessage.type}</span>
              </div>
            </div>
            <div class="mail-detail-content">
              <h2 class="mail-detail-subject">{selectedMessage.subject}</h2>
              <div class="mail-detail-info">
                <div><strong>From:</strong> {selectedMessage.from}</div>
                <div><strong>To:</strong> {selectedMessage.to}</div>
                <div><strong>Date:</strong> {new Date(selectedMessage.timestamp).toLocaleString()}</div>
                {#if selectedMessage.cc && selectedMessage.cc.length > 0}
                  <div><strong>CC:</strong> {selectedMessage.cc.join(', ')}</div>
                {/if}
              </div>
              <div class="mail-detail-body">
                <pre>{selectedMessage.body}</pre>
              </div>
            </div>
          {:else}
            <div class="empty-state">
              <span class="empty-icon">📧</span>
              <p>Select a message to read</p>
            </div>
          {/if}
        </div>
      </div>
    {/if}
    </main>
  </div>

  <footer class="footer">
    <p class="text-muted">Gas Town Dashboard v0.1.1</p>
  </footer>

  <!-- Dialogs -->
  <SlingDialog bind:isOpen={showSlingDialog} initialIssueId={slingIssueId} />

  {#if showConvoyCreator}
    <div class="modal-overlay" onclick={(e) => { if (e.target === e.currentTarget) showConvoyCreator = false; }}>
      <div class="modal-content">
        <ConvoyCreator onCreated={() => showConvoyCreator = false} />
      </div>
    </div>
  {/if}

  {#if showIssueCreator}
    <div class="modal-overlay" onclick={(e) => { if (e.target === e.currentTarget) showIssueCreator = false; }}>
      <div class="modal-content">
        <IssueCreator onCreated={() => showIssueCreator = false} />
      </div>
    </div>
  {/if}

  {#if showMailComposer}
    <div class="modal-overlay" onclick={(e) => { if (e.target === e.currentTarget) showMailComposer = false; }}>
      <div class="modal-content">
        <div class="composer-header">
          <h3>New Message</h3>
          <button class="close-btn" onclick={() => showMailComposer = false}>✕</button>
        </div>
        <MailComposer onSent={() => { showMailComposer = false; fetchMail(); }} />
      </div>
    </div>
  {/if}
</div>

<style>
  .app {
    display: flex;
    flex-direction: column;
    min-height: 100vh;
  }

  .app-layout {
    display: flex;
    flex: 1;
    overflow: hidden;
  }

  .sidebar-toggle {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 2rem;
    height: 2rem;
    background: none;
    border: 1px solid var(--color-border);
    border-radius: var(--radius-md);
    cursor: pointer;
    font-size: 1rem;
    color: var(--color-text);
    transition: all var(--transition-fast);
  }

  .sidebar-toggle:hover {
    background-color: var(--color-surface-raised);
  }

  /* Agent Workspace Placeholder */
  .agent-workspace-placeholder {
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
  }

  .agent-header {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    flex-wrap: wrap;
  }

  .agent-header h2 {
    font-size: 1.5rem;
    font-weight: 600;
    margin: 0;
  }

  .agent-role {
    padding: var(--space-1) var(--space-2);
    background-color: var(--color-surface-raised);
    border-radius: var(--radius-sm);
    font-size: 0.75rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: var(--color-text-muted);
  }

  .status-badge {
    padding: var(--space-1) var(--space-2);
    border-radius: var(--radius-sm);
    font-size: 0.75rem;
    font-weight: 500;
    background-color: var(--color-text-muted);
    color: white;
  }

  .status-badge.running {
    background-color: var(--color-success);
  }

  .current-work-banner {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    padding: var(--space-3);
    background-color: var(--color-surface-raised);
    border-radius: var(--radius-md);
    border-left: 3px solid var(--color-primary);
  }

  .work-label {
    font-size: 0.75rem;
    color: var(--color-text-muted);
    text-transform: uppercase;
  }

  .work-id {
    font-family: var(--font-mono);
    font-size: 0.875rem;
    color: var(--color-primary);
  }

  .work-title-text {
    font-size: 0.875rem;
    color: var(--color-text);
  }

  .agent-content {
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
  }

  .placeholder-text {
    color: var(--color-text-muted);
    font-style: italic;
  }

  .agent-stats {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: var(--space-3);
  }

  .stat {
    padding: var(--space-3);
    background-color: var(--color-surface);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-md);
  }

  .stat-label {
    display: block;
    font-size: 0.75rem;
    color: var(--color-text-muted);
    text-transform: uppercase;
    letter-spacing: 0.05em;
    margin-bottom: var(--space-1);
  }

  .stat-value {
    font-size: 0.875rem;
    font-weight: 500;
  }

  .header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: var(--space-3) var(--space-4);
    background-color: var(--color-surface);
    border-bottom: 1px solid var(--color-border);
    flex-wrap: wrap;
    gap: var(--space-3);
  }

  .header-left {
    display: flex;
    align-items: center;
    gap: var(--space-4);
    flex-wrap: wrap;
  }

  .header-right {
    display: flex;
    align-items: center;
    gap: var(--space-3);
  }

  .logo {
    display: flex;
    align-items: center;
    gap: var(--space-2);
  }

  .logo-icon {
    font-size: 1.25rem;
  }

  .logo h1 {
    font-size: 1.125rem;
    font-weight: 600;
  }

  .nav-tabs {
    display: flex;
    gap: var(--space-1);
  }

  .nav-tab {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    padding: var(--space-2) var(--space-3);
    background: none;
    border: none;
    border-radius: var(--radius-md);
    color: var(--color-text-muted);
    font-size: 0.875rem;
    cursor: pointer;
    transition: all var(--transition-fast);
  }

  .nav-tab:hover {
    background-color: var(--color-surface-raised);
    color: var(--color-text);
  }

  .nav-tab.active {
    background-color: var(--color-primary);
    color: white;
  }

  .town-name {
    font-weight: 500;
    color: var(--color-primary);
    display: none;
  }

  @media (min-width: 640px) {
    .town-name {
      display: block;
    }
  }

  .theme-toggle {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 2rem;
    height: 2rem;
    background: none;
    border: 1px solid var(--color-border);
    border-radius: var(--radius-md);
    cursor: pointer;
    transition: all var(--transition-fast);
  }

  .theme-toggle:hover {
    background-color: var(--color-surface-raised);
  }

  .theme-icon {
    font-size: 1rem;
  }

  .mail-badge {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-width: 1.25rem;
    height: 1.25rem;
    padding: 0 var(--space-1);
    font-size: 0.75rem;
    font-weight: 600;
    background-color: var(--color-error);
    color: white;
    border-radius: 9999px;
  }

  .mail-badge.small {
    min-width: 1rem;
    height: 1rem;
    font-size: 0.625rem;
  }

  .main {
    flex: 1;
    padding: var(--space-4);
    max-width: 1400px;
    margin: 0 auto;
    width: 100%;
  }

  @media (min-width: 768px) {
    .main {
      padding: var(--space-6);
    }
  }

  .loading, .error-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: var(--space-4);
    padding: var(--space-8);
    text-align: center;
  }

  .error-icon {
    font-size: 2.5rem;
  }

  .error-text {
    color: var(--color-error);
  }

  .btn {
    padding: var(--space-2) var(--space-4);
    border: none;
    border-radius: var(--radius-md);
    font-size: 0.875rem;
    cursor: pointer;
    transition: all var(--transition-fast);
  }

  .btn-primary {
    background-color: var(--color-primary);
    color: white;
  }

  .btn-primary:hover {
    opacity: 0.9;
  }

  .btn-icon {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 2rem;
    height: 2rem;
    padding: 0;
    background: none;
    border: 1px solid var(--color-border);
    border-radius: var(--radius-md);
    cursor: pointer;
  }

  .btn-sm {
    padding: var(--space-1) var(--space-2);
    font-size: 0.75rem;
  }

  .summary-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: var(--space-3);
    margin-bottom: var(--space-4);
  }

  @media (min-width: 640px) {
    .summary-grid {
      grid-template-columns: repeat(4, 1fr);
      gap: var(--space-4);
      margin-bottom: var(--space-6);
    }
  }

  .summary-card {
    text-align: center;
    padding: var(--space-4);
  }

  @media (min-width: 640px) {
    .summary-card {
      padding: var(--space-6);
    }
  }

  .summary-value {
    font-size: 1.5rem;
    font-weight: 700;
    color: var(--color-primary);
  }

  @media (min-width: 640px) {
    .summary-value {
      font-size: 2rem;
    }
  }

  .summary-label {
    font-size: 0.75rem;
    color: var(--color-text-muted);
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  @media (min-width: 640px) {
    .summary-label {
      font-size: 0.875rem;
    }
  }

  .agent-list {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-3);
  }

  .agent-item {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    padding: var(--space-2) var(--space-3);
    background-color: var(--color-surface-raised);
    border-radius: var(--radius-md);
    flex: 1 1 auto;
    min-width: 150px;
  }

  .agent-icon {
    font-size: 1.25rem;
  }

  .agent-icon.small {
    font-size: 1rem;
  }

  .agent-name {
    font-weight: 500;
    font-size: 0.875rem;
  }

  .section-header {
    font-size: 0.875rem;
    font-weight: 500;
    color: var(--color-text-muted);
    text-transform: uppercase;
    letter-spacing: 0.05em;
    margin: var(--space-4) 0;
  }

  @media (min-width: 640px) {
    .section-header {
      font-size: 1rem;
      margin: var(--space-6) 0 var(--space-4);
    }
  }

  .rigs-grid {
    display: grid;
    grid-template-columns: 1fr;
    gap: var(--space-3);
  }

  @media (min-width: 640px) {
    .rigs-grid {
      grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
      gap: var(--space-4);
    }
  }

  .rig-card {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }

  .rig-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .rig-name {
    font-size: 1rem;
    font-weight: 600;
  }

  .rig-services {
    display: flex;
    gap: var(--space-2);
  }

  .service-badge {
    font-size: 1rem;
  }

  .rig-path {
    font-size: 0.7rem;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .rig-stats {
    display: flex;
    gap: var(--space-3);
    font-size: 0.8rem;
    color: var(--color-text-muted);
  }

  .rig-agents {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    margin-top: var(--space-2);
    padding-top: var(--space-2);
    border-top: 1px solid var(--color-border);
  }

  .agent-row {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    font-size: 0.8rem;
    flex-wrap: wrap;
  }

  .work-title {
    margin-left: auto;
    font-size: 0.7rem;
    color: var(--color-text-muted);
    max-width: 150px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  /* Mail styles */
  .mail-container {
    display: flex;
    gap: var(--space-4);
    height: calc(100vh - 180px);
    min-height: 400px;
  }

  .mail-list {
    flex: 1;
    display: flex;
    flex-direction: column;
    background-color: var(--color-surface);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-md);
    overflow: hidden;
    max-width: 100%;
  }

  @media (min-width: 768px) {
    .mail-list {
      max-width: 400px;
    }

    .mail-list.has-selected {
      display: flex;
    }
  }

  @media (max-width: 767px) {
    .mail-list.has-selected {
      display: none;
    }
  }

  .mail-list-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: var(--space-3);
    border-bottom: 1px solid var(--color-border);
  }

  .mail-list-header h2 {
    font-size: 1rem;
    font-weight: 600;
  }

  .mail-list-actions {
    display: flex;
    gap: var(--space-2);
    align-items: center;
  }

  .mail-items {
    flex: 1;
    overflow-y: auto;
  }

  .mail-item {
    display: block;
    width: 100%;
    padding: var(--space-3);
    background: none;
    border: none;
    border-bottom: 1px solid var(--color-border);
    text-align: left;
    cursor: pointer;
    transition: background-color var(--transition-fast);
  }

  .mail-item:hover {
    background-color: var(--color-surface-raised);
  }

  .mail-item.selected {
    background-color: var(--color-surface-raised);
    border-left: 3px solid var(--color-primary);
  }

  .mail-item.unread {
    background-color: rgba(59, 130, 246, 0.05);
  }

  .mail-item.unread .mail-subject {
    font-weight: 600;
  }

  .mail-item-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: var(--space-1);
  }

  .mail-from {
    font-size: 0.8rem;
    font-weight: 500;
    color: var(--color-text);
  }

  .mail-time {
    font-size: 0.7rem;
    color: var(--color-text-muted);
  }

  .mail-subject {
    display: flex;
    align-items: center;
    gap: var(--space-1);
    font-size: 0.875rem;
    color: var(--color-text);
    margin-bottom: var(--space-1);
  }

  .pin-icon {
    font-size: 0.75rem;
  }

  .mail-preview {
    font-size: 0.75rem;
    color: var(--color-text-muted);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .mail-detail {
    display: none;
    flex: 2;
    background-color: var(--color-surface);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-md);
    overflow: hidden;
  }

  @media (min-width: 768px) {
    .mail-detail {
      display: flex;
      flex-direction: column;
    }
  }

  @media (max-width: 767px) {
    .mail-detail.visible {
      display: flex;
      flex-direction: column;
      position: fixed;
      top: 0;
      left: 0;
      right: 0;
      bottom: 0;
      z-index: 100;
      border-radius: 0;
    }
  }

  .mail-detail-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: var(--space-3);
    border-bottom: 1px solid var(--color-border);
  }

  .btn-back {
    display: none;
    padding: var(--space-2);
    background: none;
    border: none;
    color: var(--color-primary);
    cursor: pointer;
  }

  @media (max-width: 767px) {
    .btn-back {
      display: block;
    }
  }

  .mail-detail-meta {
    display: flex;
    gap: var(--space-2);
  }

  .mail-detail-content {
    flex: 1;
    overflow-y: auto;
    padding: var(--space-4);
  }

  .mail-detail-subject {
    font-size: 1.25rem;
    font-weight: 600;
    margin-bottom: var(--space-4);
  }

  .mail-detail-info {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    padding-bottom: var(--space-4);
    margin-bottom: var(--space-4);
    border-bottom: 1px solid var(--color-border);
    font-size: 0.875rem;
  }

  .mail-detail-body {
    font-size: 0.9rem;
    line-height: 1.6;
  }

  .mail-detail-body pre {
    white-space: pre-wrap;
    word-wrap: break-word;
    font-family: inherit;
    margin: 0;
  }

  /* Priority colors */
  .priority-urgent {
    color: var(--color-error);
  }

  .priority-high {
    color: var(--color-warning);
  }

  .priority-normal {
    color: var(--color-text);
  }

  .priority-low {
    color: var(--color-text-muted);
  }

  /* Empty and loading states */
  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: var(--space-8);
    color: var(--color-text-muted);
    height: 100%;
  }

  .empty-icon {
    font-size: 3rem;
    margin-bottom: var(--space-3);
  }

  .loading-inline, .error-inline {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: var(--space-3);
    padding: var(--space-4);
    color: var(--color-text-muted);
  }

  .error-inline {
    color: var(--color-error);
  }

  .spinner.small {
    width: 1rem;
    height: 1rem;
  }

  .footer {
    padding: var(--space-3) var(--space-4);
    text-align: center;
    border-top: 1px solid var(--color-border);
  }

  /* Create dropdown */
  .create-dropdown {
    position: relative;
  }

  .create-btn {
    display: flex;
    align-items: center;
    gap: var(--space-1);
    padding: var(--space-2) var(--space-3);
    font-size: 0.875rem;
  }

  .create-menu {
    position: absolute;
    top: 100%;
    right: 0;
    margin-top: var(--space-2);
    background: var(--color-surface);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-md);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
    min-width: 180px;
    z-index: 100;
    overflow: hidden;
  }

  .create-menu button {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    width: 100%;
    padding: var(--space-3) var(--space-4);
    text-align: left;
    background: none;
    border: none;
    cursor: pointer;
    font-size: 0.875rem;
    color: var(--color-text);
    transition: background-color var(--transition-fast);
  }

  .create-menu button:hover {
    background: var(--color-surface-raised);
  }

  /* Modal overlay */
  .modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
    padding: var(--space-4);
  }

  .modal-content {
    background: var(--color-surface);
    border-radius: var(--radius-lg);
    max-width: 500px;
    width: 100%;
    max-height: 90vh;
    overflow: auto;
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
  }

  /* Issues view */
  .issues-view {
    max-width: 900px;
  }

  /* Mail composer modal header */
  .composer-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: var(--space-4);
    border-bottom: 1px solid var(--color-border);
  }

  .composer-header h3 {
    margin: 0;
    font-size: 1.125rem;
    font-weight: 600;
  }

  .close-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 2rem;
    height: 2rem;
    background: none;
    border: none;
    color: var(--color-text-muted);
    font-size: 1.25rem;
    cursor: pointer;
    border-radius: var(--radius-sm);
    transition: background-color var(--transition-fast), color var(--transition-fast);
  }

  .close-btn:hover {
    background-color: var(--color-surface-raised);
    color: var(--color-text);
  }
</style>
