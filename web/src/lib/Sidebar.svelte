<script lang="ts">
  import { createEventDispatcher } from 'svelte';

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

  export let status: TownStatus;
  export let selectedAgent: Agent | null;
  export let currentView: 'chat' | 'dashboard';

  const dispatch = createEventDispatcher();

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
    dispatch('selectAgent', agent);
  }

  function changeView(view: 'chat' | 'dashboard') {
    dispatch('viewChange', view);
  }

  // Find Mayor agent
  const mayor = status.agents.find(a => a.role === 'mayor');

  // Group rig agents by rig
  const rigAgents = new Map<string, Agent[]>();
  status.rigs.forEach(rig => {
    if (rig.agents?.length) {
      rigAgents.set(rig.name, rig.agents);
    }
  });
</script>

<aside class="sidebar">
  <!-- Header -->
  <div class="sidebar-header">
    <div class="logo">
      <span class="logo-icon">⛽</span>
      <div class="logo-text">
        <div class="logo-title">GAS TOWN</div>
        <div class="logo-subtitle">Control Center</div>
      </div>
    </div>
  </div>

  <!-- Navigation -->
  <nav class="sidebar-nav">
    <button
      class="nav-item"
      class:active={currentView === 'chat'}
      on:click={() => changeView('chat')}
    >
      <span class="nav-icon">💬</span>
      <span class="nav-label">Chat</span>
    </button>
    <button
      class="nav-item"
      class:active={currentView === 'dashboard'}
      on:click={() => changeView('dashboard')}
    >
      <span class="nav-icon">📊</span>
      <span class="nav-label">Dashboard</span>
    </button>
  </nav>

  <!-- Divider -->
  <div class="sidebar-divider"></div>

  <!-- Mayor Section (Prominent) -->
  {#if mayor}
    <div class="agents-section">
      <div class="section-title mayor-section-title">COMMAND</div>
      <button
        class="agent-button mayor"
        class:selected={selectedAgent?.role === 'mayor'}
        on:click={() => selectAgent(mayor)}
      >
        <div class="agent-button-icon" style="color: {getRoleColor(mayor.role)}">
          {getRoleIcon(mayor.role)}
        </div>
        <div class="agent-button-content">
          <div class="agent-button-name">{mayor.name}</div>
          <div class="agent-button-status">
            <span class="status-dot" class:running={mayor.running}></span>
            {mayor.running ? 'Active' : 'Offline'}
          </div>
        </div>
        {#if mayor.unread_mail > 0}
          <div class="mail-badge">{mayor.unread_mail}</div>
        {/if}
      </button>
    </div>

    <!-- Divider -->
    <div class="sidebar-divider"></div>
  {/if}

  <!-- Rigs and their agents -->
  <div class="agents-section">
    <div class="section-title">RIGS & AGENTS</div>

    <!-- Global agents (non-rig) -->
    {#each status.agents as agent}
      {#if agent.role !== 'mayor' && !Array.from(rigAgents.values()).flat().some(a => a.name === agent.name)}
        <button
          class="agent-button"
          class:selected={selectedAgent?.name === agent.name}
          on:click={() => selectAgent(agent)}
        >
          <div class="agent-button-icon" style="color: {getRoleColor(agent.role)}">
            {getRoleIcon(agent.role)}
          </div>
          <div class="agent-button-content">
            <div class="agent-button-name">{agent.name}</div>
            <div class="agent-button-role">{agent.role}</div>
          </div>
          {#if agent.unread_mail > 0}
            <div class="mail-badge small">{agent.unread_mail}</div>
          {/if}
        </button>
      {/if}
    {/each}

    <!-- Rigs with their agents -->
    {#each status.rigs as rig}
      {#if rigAgents.has(rig.name)}
        {@const agents = rigAgents.get(rig.name)}
        {#if agents && agents.length > 0}
          <div class="rig-group">
            <div class="rig-name">{rig.name}</div>
            {#each agents as agent}
            <button
              class="agent-button nested"
              class:selected={selectedAgent?.name === agent.name}
              on:click={() => selectAgent(agent)}
            >
              <div class="agent-button-icon small" style="color: {getRoleColor(agent.role)}">
                {getRoleIcon(agent.role)}
              </div>
              <div class="agent-button-content">
                <div class="agent-button-name">{agent.name}</div>
                <div class="agent-button-status">
                  <span class="status-dot" class:running={agent.running} class:working={agent.has_work}></span>
                  {agent.running ? (agent.has_work ? 'Working' : 'Ready') : 'Offline'}
                </div>
              </div>
              {#if agent.unread_mail > 0}
                <div class="mail-badge small">{agent.unread_mail}</div>
              {/if}
            </button>
            {/each}
          </div>
        {/if}
      {/if}
    {/each}
  </div>

  <!-- Footer -->
  <div class="sidebar-footer">
    <div class="footer-stat">
      <span class="label">Rigs</span>
      <span class="value">{status.summary.rig_count}</span>
    </div>
    <div class="footer-stat">
      <span class="label">Active</span>
      <span class="value">{status.summary.active_hooks}</span>
    </div>
  </div>
</aside>

<style>
  .sidebar {
    width: 320px;
    background-color: var(--color-surface);
    border-right: 2px solid var(--color-border);
    display: flex;
    flex-direction: column;
    overflow-y: auto;
    font-family: var(--font-mono);
  }

  .sidebar-header {
    padding: var(--space-4);
    border-bottom: 2px solid var(--color-border);
  }

  .logo {
    display: flex;
    align-items: center;
    gap: var(--space-3);
  }

  .logo-icon {
    font-size: 1.75rem;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 40px;
    height: 40px;
    background: linear-gradient(135deg, #d4af37, #f4d03f);
    border-radius: 4px;
    box-shadow: 0 0 10px rgba(212, 175, 55, 0.3);
  }

  .logo-text {
    flex: 1;
  }

  .logo-title {
    font-size: 0.875rem;
    font-weight: 700;
    letter-spacing: 0.1em;
    color: var(--color-text);
    text-shadow: 0 0 10px rgba(168, 85, 247, 0.3);
  }

  .logo-subtitle {
    font-size: 0.625rem;
    color: var(--color-text-muted);
    letter-spacing: 0.05em;
    margin-top: 2px;
  }

  .sidebar-nav {
    display: flex;
    gap: var(--space-2);
    padding: var(--space-3);
    border-bottom: 1px solid var(--color-border);
  }

  .nav-item {
    flex: 1;
    padding: var(--space-2) var(--space-3);
    background-color: transparent;
    border: 1px solid var(--color-border);
    border-radius: 4px;
    color: var(--color-text-muted);
    font-size: 0.75rem;
    font-family: var(--font-mono);
    cursor: pointer;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 4px;
    transition: all var(--transition-normal);
  }

  .nav-item:hover {
    border-color: var(--color-mayor);
    color: var(--color-mayor);
  }

  .nav-item.active {
    background-color: rgba(168, 85, 247, 0.1);
    border-color: var(--color-mayor);
    color: var(--color-mayor);
    box-shadow: 0 0 8px rgba(168, 85, 247, 0.2);
  }

  .nav-icon {
    font-size: 1.25rem;
  }

  .nav-label {
    font-weight: 600;
    letter-spacing: 0.05em;
  }

  .sidebar-divider {
    height: 1px;
    background: linear-gradient(90deg, transparent, var(--color-border), transparent);
    margin: var(--space-3) 0;
  }

  .agents-section {
    flex: 1;
    overflow-y: auto;
    padding: 0 var(--space-2);
    min-width: 0;
  }

  .section-title {
    font-size: 0.625rem;
    font-weight: 700;
    letter-spacing: 0.15em;
    color: var(--color-text-muted);
    padding: var(--space-3) var(--space-2);
    text-transform: uppercase;
    margin-top: var(--space-2);
  }

  .mayor-section-title {
    color: var(--color-mayor);
    text-shadow: 0 0 8px rgba(168, 85, 247, 0.3);
  }

  .agent-button {
    width: 100%;
    display: flex;
    align-items: center;
    gap: var(--space-2);
    padding: var(--space-2) var(--space-2);
    margin-bottom: var(--space-1);
    background-color: transparent;
    border: 1px solid transparent;
    border-radius: 4px;
    color: var(--color-text);
    cursor: pointer;
    font-family: var(--font-mono);
    font-size: 0.8rem;
    transition: all var(--transition-normal);
    position: relative;
  }

  .agent-button:hover {
    background-color: rgba(255, 255, 255, 0.05);
    border-color: var(--color-border);
  }

  .agent-button.mayor {
    padding: var(--space-3);
    margin-bottom: var(--space-2);
    border: 2px solid var(--color-mayor);
    background: linear-gradient(135deg, rgba(168, 85, 247, 0.1), rgba(168, 85, 247, 0.05));
    box-shadow: 0 0 12px rgba(168, 85, 247, 0.2);
    font-weight: 600;
  }

  .agent-button.mayor:hover {
    background: linear-gradient(135deg, rgba(168, 85, 247, 0.15), rgba(168, 85, 247, 0.1));
    box-shadow: 0 0 16px rgba(168, 85, 247, 0.3);
  }

  .agent-button.selected {
    background-color: rgba(255, 255, 255, 0.1);
    border-color: var(--color-primary);
    box-shadow: 0 0 8px rgba(59, 130, 246, 0.3);
  }

  .agent-button.nested {
    padding: var(--space-2) var(--space-2) var(--space-2) var(--space-3);
    margin-left: var(--space-2);
  }

  .agent-button-icon {
    font-size: 1.25rem;
    min-width: 32px;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .agent-button-icon.small {
    font-size: 1rem;
    min-width: 24px;
  }

  .agent-button-content {
    flex: 1;
    min-width: 0;
    text-align: left;
  }

  .agent-button-name {
    font-weight: 600;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .agent-button-role {
    font-size: 0.7rem;
    color: var(--color-text-muted);
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .agent-button-status {
    font-size: 0.7rem;
    color: var(--color-text-muted);
    display: flex;
    align-items: center;
    gap: var(--space-1);
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

  .status-dot.working {
    background-color: var(--color-primary);
    animation: pulse 2s infinite;
  }

  .mail-badge {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-width: 1.25rem;
    height: 1.25rem;
    padding: 0 var(--space-1);
    font-size: 0.65rem;
    font-weight: 700;
    background-color: var(--color-error);
    color: white;
    border-radius: 50%;
    flex-shrink: 0;
  }

  .mail-badge.small {
    min-width: 1rem;
    height: 1rem;
    font-size: 0.6rem;
  }

  .rig-group {
    margin-top: var(--space-3);
  }

  .rig-name {
    font-size: 0.7rem;
    font-weight: 700;
    color: var(--color-refinery);
    padding: var(--space-2) var(--space-2);
    text-transform: uppercase;
    letter-spacing: 0.08em;
    margin-top: var(--space-2);
  }

  .sidebar-footer {
    padding: var(--space-3);
    border-top: 2px solid var(--color-border);
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: var(--space-3);
  }

  .footer-stat {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: var(--space-1);
  }

  .footer-stat .label {
    font-size: 0.65rem;
    color: var(--color-text-muted);
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .footer-stat .value {
    font-size: 1.25rem;
    font-weight: 700;
    color: var(--color-primary);
  }

  @keyframes pulse {
    0%, 100% {
      opacity: 1;
    }
    50% {
      opacity: 0.5;
    }
  }

  /* Scrollbar styling */
  .agents-section::-webkit-scrollbar {
    width: 6px;
  }

  .agents-section::-webkit-scrollbar-track {
    background: transparent;
  }

  .agents-section::-webkit-scrollbar-thumb {
    background: var(--color-border);
    border-radius: 3px;
  }

  .agents-section::-webkit-scrollbar-thumb:hover {
    background: var(--color-gray);
  }
</style>
