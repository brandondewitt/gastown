<script lang="ts">
  import { onMount } from 'svelte';
  import './app.css';
  import { getWebSocketClient, MessageTypes, type WSMessage, type AgentUpdate, destroyWebSocketClient } from './lib/websocket';

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

  let status: TownStatus | null = null;
  let loading = true;
  let error: string | null = null;
  let wsConnected = false;
  let lastUpdate: Date | null = null;

  async function fetchStatus() {
    try {
      const response = await fetch('/api/v1/status');
      const data = await response.json();
      if (data.success) {
        status = data.data;
        error = null;
        lastUpdate = new Date();
      } else {
        error = data.error?.message || 'Unknown error';
      }
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to fetch status';
    } finally {
      loading = false;
    }
  }

  function handleAgentUpdate(message: WSMessage) {
    if (!status) return;

    const update = message.payload as AgentUpdate;
    lastUpdate = new Date();

    // Update agent in global agents list
    const globalIdx = status.agents?.findIndex(a => a.address === update.address);
    if (globalIdx !== undefined && globalIdx >= 0) {
      status.agents[globalIdx] = {
        ...status.agents[globalIdx],
        running: update.running,
        has_work: update.has_work,
        hook_bead: update.hook_bead,
        work_title: update.work_title,
        state: update.state,
      };
      status = status; // Trigger reactivity
      return;
    }

    // Update agent in rig agents
    for (const rig of status.rigs ?? []) {
      const rigIdx = rig.agents?.findIndex(a => a.address === update.address);
      if (rigIdx !== undefined && rigIdx >= 0 && rig.agents) {
        rig.agents[rigIdx] = {
          ...rig.agents[rigIdx],
          running: update.running,
          has_work: update.has_work,
          hook_bead: update.hook_bead,
          work_title: update.work_title,
          state: update.state,
        };
        status = status; // Trigger reactivity
        return;
      }
    }

    // Agent not found - might be new, refresh full status
    console.log('[WS] Agent not found, refreshing:', update.address);
    fetchStatus();
  }

  onMount(() => {
    // Initial fetch
    fetchStatus();

    // Setup WebSocket for real-time updates
    const ws = getWebSocketClient();

    ws.on('_connected', () => {
      wsConnected = true;
      console.log('[App] WebSocket connected');
    });

    ws.on('_disconnected', () => {
      wsConnected = false;
      console.log('[App] WebSocket disconnected');
    });

    ws.on(MessageTypes.AGENT_UPDATE, handleAgentUpdate);

    // Connect
    ws.connect();

    // Fallback polling every 30s in case WS is disconnected
    const interval = setInterval(() => {
      if (!wsConnected) {
        fetchStatus();
      }
    }, 30000);

    return () => {
      clearInterval(interval);
      destroyWebSocketClient();
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
</script>

<div class="app">
  <header class="header">
    <div class="logo">
      <span class="logo-icon">⛽</span>
      <h1>Gas Town</h1>
    </div>
    {#if status}
      <div class="header-info">
        <span class="town-name">{status.name}</span>
        {#if status.overseer}
          <span class="overseer">
            {status.overseer.name}
            {#if status.overseer.unread_mail > 0}
              <span class="mail-badge">{status.overseer.unread_mail}</span>
            {/if}
          </span>
        {/if}
        <span class="ws-status" class:connected={wsConnected} title={wsConnected ? 'Live updates active' : 'Reconnecting...'}>
          <span class="ws-dot"></span>
          {wsConnected ? 'Live' : 'Offline'}
        </span>
      </div>
    {/if}
  </header>

  <main class="main">
    {#if loading}
      <div class="loading">
        <div class="spinner"></div>
        <p>Loading town status...</p>
      </div>
    {:else if error}
      <div class="error-container">
        <p class="text-error">Error: {error}</p>
        <button on:click={fetchStatus}>Retry</button>
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
    {/if}
  </main>

  <footer class="footer">
    <p class="text-muted">Gas Town Dashboard v0.1.0</p>
  </footer>
</div>

<style>
  .app {
    display: flex;
    flex-direction: column;
    min-height: 100vh;
  }

  .header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: var(--space-4) var(--space-6);
    background-color: var(--color-surface);
    border-bottom: 1px solid var(--color-border);
  }

  .logo {
    display: flex;
    align-items: center;
    gap: var(--space-3);
  }

  .logo-icon {
    font-size: 1.5rem;
  }

  .logo h1 {
    font-size: 1.25rem;
    font-weight: 600;
  }

  .header-info {
    display: flex;
    align-items: center;
    gap: var(--space-4);
  }

  .town-name {
    font-weight: 500;
    color: var(--color-primary);
  }

  .overseer {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    color: var(--color-text-muted);
  }

  .ws-status {
    display: flex;
    align-items: center;
    gap: var(--space-1);
    font-size: 0.75rem;
    color: var(--color-text-muted);
    padding: var(--space-1) var(--space-2);
    border-radius: var(--radius-sm);
    background-color: var(--color-surface-raised);
  }

  .ws-status.connected {
    color: var(--color-success, #22c55e);
  }

  .ws-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background-color: var(--color-text-muted);
  }

  .ws-status.connected .ws-dot {
    background-color: var(--color-success, #22c55e);
    animation: pulse 2s ease-in-out infinite;
  }

  @keyframes pulse {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.5; }
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
    background-color: var(--color-primary);
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
    padding: var(--space-6);
    max-width: 1400px;
    margin: 0 auto;
    width: 100%;
  }

  .loading {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: var(--space-4);
    padding: var(--space-8);
  }

  .error-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: var(--space-4);
    padding: var(--space-8);
  }

  .error-container button {
    padding: var(--space-2) var(--space-4);
    background-color: var(--color-primary);
    color: white;
    border: none;
    border-radius: var(--radius-md);
    cursor: pointer;
  }

  .summary-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
    gap: var(--space-4);
    margin-bottom: var(--space-6);
  }

  .summary-card {
    text-align: center;
    padding: var(--space-6);
  }

  .summary-value {
    font-size: 2rem;
    font-weight: 700;
    color: var(--color-primary);
  }

  .summary-label {
    font-size: 0.875rem;
    color: var(--color-text-muted);
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .agent-list {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-4);
  }

  .agent-item {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    padding: var(--space-2) var(--space-3);
    background-color: var(--color-surface-raised);
    border-radius: var(--radius-md);
  }

  .agent-icon {
    font-size: 1.25rem;
  }

  .agent-icon.small {
    font-size: 1rem;
  }

  .agent-name {
    font-weight: 500;
  }

  .section-header {
    font-size: 1rem;
    font-weight: 500;
    color: var(--color-text-muted);
    text-transform: uppercase;
    letter-spacing: 0.05em;
    margin: var(--space-6) 0 var(--space-4);
  }

  .rigs-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
    gap: var(--space-4);
  }

  .rig-card {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
  }

  .rig-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .rig-name {
    font-size: 1.125rem;
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
    font-size: 0.75rem;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .rig-stats {
    display: flex;
    gap: var(--space-4);
    font-size: 0.875rem;
    color: var(--color-text-muted);
  }

  .rig-agents {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    margin-top: var(--space-2);
    padding-top: var(--space-3);
    border-top: 1px solid var(--color-border);
  }

  .agent-row {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    font-size: 0.875rem;
  }

  .work-title {
    margin-left: auto;
    font-size: 0.75rem;
    color: var(--color-text-muted);
    max-width: 150px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .footer {
    padding: var(--space-4) var(--space-6);
    text-align: center;
    border-top: 1px solid var(--color-border);
  }
</style>
