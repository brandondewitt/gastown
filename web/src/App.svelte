<script lang="ts">
  import { onMount } from 'svelte';
  import './app.css';

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
  let selectedAgent: Agent | null = null;
  let mayorAgent: Agent | null = null;
  let viewMode: 'chat' | 'dashboard' = 'chat';
  let sidebarOpen = true;
  let expandedRigs: Set<string> = new Set();

  async function fetchStatus() {
    try {
      const response = await fetch('/api/v1/status');
      const data = await response.json();
      if (data.success) {
        status = data.data;
        error = null;

        // Create a virtual Mayor agent from overseer data
        if (status.overseer) {
          mayorAgent = {
            name: status.overseer.name || 'Mayor',
            address: 'mayor',
            session: 'global',
            role: 'mayor',
            running: true,
            has_work: false,
            unread_mail: status.overseer.unread_mail || 0,
          };

          // Auto-select Mayor on first load
          if (!selectedAgent) {
            selectedAgent = mayorAgent;
          }
        }
      } else {
        error = data.error?.message || 'Unknown error';
      }
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to fetch status';
    } finally {
      loading = false;
    }
  }

  onMount(() => {
    fetchStatus();
    // Refresh every 5 seconds
    const interval = setInterval(fetchStatus, 5000);
    return () => clearInterval(interval);
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

  function selectAgent(agent: Agent) {
    selectedAgent = agent;
    viewMode = 'chat';
  }

  function toggleRig(rigName: string) {
    if (expandedRigs.has(rigName)) {
      expandedRigs.delete(rigName);
    } else {
      expandedRigs.add(rigName);
    }
    expandedRigs = expandedRigs;
  }
</script>

<div class="app">
  <header class="header">
    <button class="sidebar-toggle" on:click={() => sidebarOpen = !sidebarOpen} aria-label="Toggle sidebar">
      <span class="toggle-icon">☰</span>
    </button>
    <div class="logo">
      <span class="logo-icon">⛽</span>
      <h1>Gas Town</h1>
    </div>
    {#if status}
      <div class="header-info">
        <span class="town-name">{status.name}</span>
        {#if status.overseer}
          <span class="overseer">
            <span class="overseer-label">{status.overseer.name}</span>
            {#if status.overseer.unread_mail > 0}
              <span class="mail-badge">{status.overseer.unread_mail}</span>
            {/if}
          </span>
        {/if}
      </div>
    {/if}
  </header>

  <div class="container">
    <!-- Sidebar -->
    <aside class="sidebar" class:open={sidebarOpen}>
      {#if loading}
        <div class="sidebar-loading">
          <div class="spinner"></div>
        </div>
      {:else if error}
        <div class="sidebar-error">
          <p>Error loading agents</p>
        </div>
      {:else if status}
        <!-- Mayor Section -->
        {#if mayorAgent}
          <div class="sidebar-section mayor-section">
            <button
              class="agent-button mayor-button"
              class:selected={selectedAgent?.address === mayorAgent.address}
              on:click={() => selectAgent(mayorAgent)}
            >
              <div class="agent-button-content">
                <span class="agent-icon mayor-icon">{getRoleIcon(mayorAgent.role)}</span>
                <div class="agent-info">
                  <div class="agent-name">Mayor</div>
                  <div class="agent-status-line">
                    <span class="status-dot" class:running={mayorAgent.running}></span>
                    <span class="status-text">{mayorAgent.running ? 'Online' : 'Offline'}</span>
                  </div>
                </div>
              </div>
              {#if mayorAgent.unread_mail > 0}
                <span class="mail-badge">{mayorAgent.unread_mail}</span>
              {/if}
            </button>
          </div>
        {/if}

        <!-- Rigs & Agents -->
        <div class="sidebar-section agents-section">
          <h2 class="section-title">Agents</h2>
          {#each status.rigs ?? [] as rig}
            <div class="rig-group">
              <button
                class="rig-header-button"
                on:click={() => toggleRig(rig.name)}
              >
                <span class="rig-toggle">{expandedRigs.has(rig.name) ? '▼' : '▶'}</span>
                <span class="rig-name">{rig.name}</span>
                <span class="rig-badge">({rig.polecat_count})</span>
              </button>
              {#if expandedRigs.has(rig.name) && rig.agents}
                <div class="agents-list">
                  {#each rig.agents as agent}
                    <button
                      class="agent-button"
                      class:selected={selectedAgent?.address === agent.address}
                      on:click={() => selectAgent(agent)}
                    >
                      <div class="agent-button-content">
                        <span class="agent-icon" style="color: {getRoleColor(agent.role)}">
                          {getRoleIcon(agent.role)}
                        </span>
                        <div class="agent-info">
                          <div class="agent-name">{agent.name}</div>
                          <div class="agent-status-line">
                            <span class="status-dot" class:running={agent.running} class:working={agent.has_work}></span>
                            <span class="status-text">
                              {#if agent.has_work}
                                Working
                              {:else if agent.running}
                                Online
                              {:else}
                                Offline
                              {/if}
                            </span>
                          </div>
                        </div>
                      </div>
                      {#if agent.unread_mail > 0}
                        <span class="mail-badge">{agent.unread_mail}</span>
                      {/if}
                    </button>
                  {/each}
                </div>
              {/if}
            </div>
          {/each}
        </div>
      {/if}
    </aside>

    <!-- Main Content -->
    <main class="main-content">
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
        <!-- View Selector -->
        <div class="view-selector">
          <button
            class="view-button"
            class:active={viewMode === 'chat'}
            on:click={() => viewMode = 'chat'}
          >
            💬 Chat
          </button>
          <button
            class="view-button"
            class:active={viewMode === 'dashboard'}
            on:click={() => viewMode = 'dashboard'}
          >
            📊 Dashboard
          </button>
        </div>

        {#if viewMode === 'chat'}
          <!-- Chat View -->
          <div class="chat-container">
            {#if selectedAgent}
              <div class="chat-header">
                <div class="chat-header-content">
                  <span class="chat-icon" style="color: {getRoleColor(selectedAgent.role)}">
                    {getRoleIcon(selectedAgent.role)}
                  </span>
                  <div class="chat-header-info">
                    <h2 class="chat-title">
                      {selectedAgent.role === 'mayor' ? 'Mayor' : selectedAgent.name}
                    </h2>
                    <p class="chat-subtitle">
                      {selectedAgent.running ? '🟢 Online' : '⚫ Offline'}
                      {#if selectedAgent.has_work}
                        · Working on task
                      {/if}
                    </p>
                  </div>
                </div>
                {#if selectedAgent.unread_mail > 0}
                  <div class="mail-indicator">
                    <span class="mail-count">{selectedAgent.unread_mail}</span>
                    <span class="mail-label">unread</span>
                  </div>
                {/if}
              </div>

              <div class="chat-content">
                <div class="message-placeholder">
                  <p>Chat interface coming soon</p>
                  <p class="text-muted">Messages with {selectedAgent.name} will appear here</p>
                </div>
              </div>

              <div class="chat-input-area">
                <input
                  type="text"
                  class="chat-input"
                  placeholder="Type a message..."
                  disabled
                />
                <button class="send-button" disabled>Send</button>
              </div>
            {/if}
          </div>
        {:else}
          <!-- Dashboard View -->
          <div class="dashboard-container">
            <!-- Summary Cards -->
            <section class="summary-grid">
              <div class="summary-card">
                <div class="summary-value">{status.summary.rig_count}</div>
                <div class="summary-label">Rigs</div>
              </div>
              <div class="summary-card">
                <div class="summary-value">{status.summary.polecat_count}</div>
                <div class="summary-label">Polecats</div>
              </div>
              <div class="summary-card">
                <div class="summary-value">{status.summary.crew_count}</div>
                <div class="summary-label">Crews</div>
              </div>
              <div class="summary-card">
                <div class="summary-value">{status.summary.active_hooks}</div>
                <div class="summary-label">Active Hooks</div>
              </div>
            </section>

            <!-- Rigs Details -->
            <section class="rigs-section">
              <h2 class="section-header">Rig Details</h2>
              <div class="rigs-grid">
                {#each status.rigs ?? [] as rig}
                  <div class="rig-card">
                    <div class="rig-header-detail">
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
                    <p class="rig-path">{rig.path}</p>
                    <div class="rig-stats">
                      <span><strong>{rig.polecat_count}</strong> polecats</span>
                      <span><strong>{rig.crew_count}</strong> crews</span>
                    </div>
                  </div>
                {/each}
              </div>
            </section>
          </div>
        {/if}
      {/if}
    </main>
  </div>
</div>

<style>
  .app {
    display: flex;
    flex-direction: column;
    min-height: 100vh;
    background-color: var(--color-background);
    color: var(--color-text);
  }

  .header {
    display: flex;
    align-items: center;
    gap: var(--space-4);
    padding: var(--space-3) var(--space-6);
    background-color: var(--color-surface);
    border-bottom: 2px solid var(--color-accent-amber);
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
  }

  .sidebar-toggle {
    display: none;
    background: none;
    border: none;
    font-size: 1.5rem;
    cursor: pointer;
    padding: 0;
    color: var(--color-accent-amber);

    @media (max-width: 768px) {
      display: flex;
      align-items: center;
      justify-content: center;
    }
  }

  .toggle-icon {
    line-height: 1;
  }

  .logo {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    flex: 0 0 auto;
  }

  .logo-icon {
    font-size: 1.75rem;
    line-height: 1;
  }

  .logo h1 {
    font-size: 1.25rem;
    font-weight: 700;
    letter-spacing: -0.01em;
    color: var(--color-accent-amber);
    margin: 0;
  }

  .header-info {
    display: flex;
    align-items: center;
    gap: var(--space-4);
    margin-left: auto;
  }

  .town-name {
    font-weight: 600;
    color: var(--color-text);
  }

  .overseer {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    color: var(--color-text-muted);
    font-size: 0.875rem;
  }

  .overseer-label {
    font-weight: 500;
  }

  .mail-badge {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-width: 1.5rem;
    height: 1.5rem;
    padding: 0 var(--space-1);
    font-size: 0.75rem;
    font-weight: 700;
    background-color: var(--color-accent-amber);
    color: var(--color-background);
    border-radius: 50%;
  }

  .container {
    display: flex;
    flex: 1;
    overflow: hidden;
  }

  /* Sidebar */
  .sidebar {
    width: 280px;
    background-color: var(--color-surface);
    border-right: 2px solid var(--color-border);
    display: flex;
    flex-direction: column;
    overflow-y: auto;
    transition: transform 0.2s ease;

    @media (max-width: 768px) {
      position: absolute;
      left: 0;
      top: 0;
      height: 100%;
      transform: translateX(-100%);
      z-index: 100;
      box-shadow: 2px 0 12px rgba(0, 0, 0, 0.2);

      &.open {
        transform: translateX(0);
      }
    }
  }

  .sidebar-loading,
  .sidebar-error {
    display: flex;
    align-items: center;
    justify-content: center;
    padding: var(--space-6);
    text-align: center;
  }

  .sidebar-section {
    padding: var(--space-3);
    border-bottom: 1px solid var(--color-border);
  }

  .sidebar-section:last-child {
    border-bottom: none;
  }

  .mayor-section {
    background: linear-gradient(135deg, rgba(217, 119, 6, 0.05) 0%, rgba(217, 119, 6, 0) 100%);
    border-bottom: 2px solid var(--color-accent-amber);
  }

  .section-title {
    font-size: 0.75rem;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.1em;
    color: var(--color-text-muted);
    margin: 0 0 var(--space-2) 0;
    padding: 0 var(--space-2);
  }

  .rig-group {
    margin-bottom: var(--space-2);
  }

  .rig-header-button {
    width: 100%;
    display: flex;
    align-items: center;
    gap: var(--space-2);
    padding: var(--space-2);
    background: none;
    border: none;
    cursor: pointer;
    color: var(--color-text-muted);
    font-size: 0.875rem;
    font-weight: 600;
    text-align: left;
    transition: background-color 0.15s ease;
    border-radius: var(--radius-sm);

    &:hover {
      background-color: rgba(217, 119, 6, 0.1);
      color: var(--color-accent-amber);
    }
  }

  .rig-toggle {
    font-size: 0.625rem;
    width: 1rem;
    text-align: center;
  }

  .rig-name {
    flex: 1;
  }

  .rig-badge {
    font-size: 0.75rem;
    color: var(--color-text-muted);
    font-weight: normal;
  }

  .agents-list {
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
    padding-left: var(--space-3);
  }

  .agent-button {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    padding: var(--space-2) var(--space-3);
    background: none;
    border: 2px solid transparent;
    cursor: pointer;
    text-align: left;
    border-radius: var(--radius-md);
    transition: all 0.15s ease;
    color: var(--color-text);

    &:hover {
      background-color: var(--color-surface-raised);
      border-color: var(--color-accent-amber);
    }

    &.selected {
      background-color: rgba(217, 119, 6, 0.15);
      border-color: var(--color-accent-amber);
      font-weight: 600;
    }
  }

  .agent-button-content {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    flex: 1;
    min-width: 0;
  }

  .agent-icon {
    font-size: 1.25rem;
    flex: 0 0 auto;
  }

  .agent-info {
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
    min-width: 0;
  }

  .agent-name {
    font-weight: 600;
    font-size: 0.875rem;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .agent-status-line {
    display: flex;
    align-items: center;
    gap: var(--space-1);
    font-size: 0.75rem;
    color: var(--color-text-muted);
  }

  .status-dot {
    width: 0.5rem;
    height: 0.5rem;
    border-radius: 50%;
    background-color: var(--color-text-muted);
    flex: 0 0 auto;

    &.running {
      background-color: #10b981;
      box-shadow: 0 0 4px rgba(16, 185, 129, 0.5);
    }

    &.working {
      background-color: var(--color-accent-amber);
      box-shadow: 0 0 4px rgba(217, 119, 6, 0.5);
      animation: pulse 2s infinite;
    }
  }

  @keyframes pulse {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.6; }
  }

  .status-text {
    flex: 1;
  }

  .mayor-button {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--space-3);
    padding: var(--space-4);
    border: 2px solid var(--color-accent-amber);

    &:hover,
    &.selected {
      background-color: rgba(217, 119, 6, 0.15);
    }
  }

  .mayor-icon {
    font-size: 2rem;
  }

  /* Main Content */
  .main-content {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    background-color: var(--color-background);
  }

  .view-selector {
    display: flex;
    gap: var(--space-2);
    padding: var(--space-4) var(--space-6);
    background-color: var(--color-surface);
    border-bottom: 1px solid var(--color-border);
  }

  .view-button {
    padding: var(--space-2) var(--space-4);
    background: none;
    border: 2px solid var(--color-border);
    border-radius: var(--radius-md);
    cursor: pointer;
    font-weight: 600;
    font-size: 0.875rem;
    color: var(--color-text-muted);
    transition: all 0.15s ease;

    &:hover {
      border-color: var(--color-accent-amber);
      color: var(--color-accent-amber);
    }

    &.active {
      background-color: var(--color-accent-amber);
      border-color: var(--color-accent-amber);
      color: var(--color-background);
    }
  }

  .loading,
  .error-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: var(--space-4);
    padding: var(--space-8);
    flex: 1;
  }

  .error-container button {
    padding: var(--space-2) var(--space-6);
    background-color: var(--color-accent-amber);
    color: var(--color-background);
    border: none;
    border-radius: var(--radius-md);
    cursor: pointer;
    font-weight: 600;
  }

  /* Chat View */
  .chat-container {
    display: flex;
    flex-direction: column;
    flex: 1;
    overflow: hidden;
  }

  .chat-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: var(--space-4) var(--space-6);
    background-color: var(--color-surface);
    border-bottom: 1px solid var(--color-border);
  }

  .chat-header-content {
    display: flex;
    align-items: center;
    gap: var(--space-3);
  }

  .chat-icon {
    font-size: 2rem;
  }

  .chat-header-info {
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
  }

  .chat-title {
    font-size: 1.125rem;
    font-weight: 700;
    margin: 0;
  }

  .chat-subtitle {
    font-size: 0.875rem;
    color: var(--color-text-muted);
    margin: 0;
  }

  .mail-indicator {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: var(--space-1);
  }

  .mail-count {
    font-size: 1.5rem;
    font-weight: 700;
    color: var(--color-accent-amber);
  }

  .mail-label {
    font-size: 0.75rem;
    color: var(--color-text-muted);
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .chat-content {
    flex: 1;
    overflow-y: auto;
    padding: var(--space-6);
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .message-placeholder {
    text-align: center;
    color: var(--color-text-muted);
  }

  .message-placeholder p {
    margin: var(--space-2) 0;
  }

  .text-muted {
    color: var(--color-text-muted);
    font-size: 0.875rem;
  }

  .chat-input-area {
    display: flex;
    gap: var(--space-2);
    padding: var(--space-4) var(--space-6);
    background-color: var(--color-surface);
    border-top: 1px solid var(--color-border);
  }

  .chat-input {
    flex: 1;
    padding: var(--space-2) var(--space-3);
    background-color: var(--color-surface-raised);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-md);
    color: var(--color-text);
    font-family: inherit;

    &::placeholder {
      color: var(--color-text-muted);
    }

    &:disabled {
      opacity: 0.5;
      cursor: not-allowed;
    }
  }

  .send-button {
    padding: var(--space-2) var(--space-4);
    background-color: var(--color-accent-amber);
    color: var(--color-background);
    border: none;
    border-radius: var(--radius-md);
    cursor: pointer;
    font-weight: 600;

    &:disabled {
      opacity: 0.5;
      cursor: not-allowed;
    }
  }

  /* Dashboard View */
  .dashboard-container {
    flex: 1;
    overflow-y: auto;
    padding: var(--space-6);
  }

  .summary-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
    gap: var(--space-4);
    margin-bottom: var(--space-8);
  }

  .summary-card {
    padding: var(--space-6);
    background-color: var(--color-surface);
    border: 2px solid var(--color-border);
    border-radius: var(--radius-lg);
    text-align: center;
    transition: all 0.15s ease;

    &:hover {
      border-color: var(--color-accent-amber);
      background-color: rgba(217, 119, 6, 0.05);
    }
  }

  .summary-value {
    font-size: 2.5rem;
    font-weight: 700;
    color: var(--color-accent-amber);
  }

  .summary-label {
    font-size: 0.75rem;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.1em;
    color: var(--color-text-muted);
    margin-top: var(--space-2);
  }

  .section-header {
    font-size: 0.875rem;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.1em;
    color: var(--color-text-muted);
    margin-bottom: var(--space-4);
  }

  .rigs-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
    gap: var(--space-4);
  }

  .rig-card {
    padding: var(--space-4);
    background-color: var(--color-surface);
    border: 2px solid var(--color-border);
    border-radius: var(--radius-lg);
    transition: all 0.15s ease;

    &:hover {
      border-color: var(--color-accent-amber);
      box-shadow: 0 4px 12px rgba(217, 119, 6, 0.1);
    }
  }

  .rig-header-detail {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: var(--space-3);
  }

  .rig-name {
    font-size: 1rem;
    font-weight: 700;
    margin: 0;
    color: var(--color-text);
  }

  .rig-services {
    display: flex;
    gap: var(--space-2);
  }

  .service-badge {
    font-size: 1.25rem;
  }

  .rig-path {
    font-size: 0.75rem;
    color: var(--color-text-muted);
    margin: 0 0 var(--space-3) 0;
    word-break: break-all;
    font-family: 'Courier New', monospace;
  }

  .rig-stats {
    display: flex;
    gap: var(--space-4);
    font-size: 0.875rem;
    color: var(--color-text-muted);
  }

  .spinner {
    width: 40px;
    height: 40px;
    border: 3px solid var(--color-border);
    border-top-color: var(--color-accent-amber);
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  @media (max-width: 768px) {
    .header {
      gap: var(--space-2);
    }

    .logo h1 {
      display: none;
    }

    .header-info {
      display: none;
    }

    .container {
      position: relative;
    }

    .sidebar {
      width: 100%;
    }
  }
</style>
