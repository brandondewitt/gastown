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

  let {
    globalAgents = [],
    rigs = [],
    selectedAgent = null,
    onSelect
  }: {
    globalAgents: Agent[];
    rigs: Rig[];
    selectedAgent: Agent | null;
    onSelect: (agent: Agent) => void;
  } = $props();

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
</script>

<aside class="agent-sidebar">
  <div class="sidebar-header">AGENTS</div>

  <!-- Global Agents -->
  {#if globalAgents.length > 0}
    <div class="agent-group">
      <div class="group-header">GLOBAL</div>
      {#each globalAgents as agent}
        <button
          class="agent-item"
          class:selected={selectedAgent?.address === agent.address}
          class:running={agent.running}
          onclick={() => onSelect(agent)}
        >
          <span class="status-dot" class:active={agent.running}></span>
          <span class="agent-icon">{getRoleIcon(agent.role)}</span>
          <span class="agent-name">{agent.name}</span>
          {#if agent.unread_mail > 0}
            <span class="mail-badge">{agent.unread_mail}</span>
          {/if}
        </button>
      {/each}
    </div>
  {/if}

  <!-- Rig Agents -->
  {#each rigs as rig}
    {#if rig.agents && rig.agents.length > 0}
      <div class="agent-group">
        <div class="group-header">{rig.name.toUpperCase()}</div>
        {#each rig.agents as agent}
          <button
            class="agent-item"
            class:selected={selectedAgent?.address === agent.address}
            class:running={agent.running}
            onclick={() => onSelect(agent)}
          >
            <span class="status-dot" class:active={agent.running}></span>
            <span class="agent-icon">{getRoleIcon(agent.role)}</span>
            <span class="agent-name">{agent.name}</span>
            {#if agent.has_work}
              <span class="work-indicator" title={agent.work_title}>📋</span>
            {/if}
            {#if agent.unread_mail > 0}
              <span class="mail-badge">{agent.unread_mail}</span>
            {/if}
          </button>
        {/each}
      </div>
    {/if}
  {/each}
</aside>

<style>
  .agent-sidebar {
    width: 240px;
    min-width: 240px;
    background: var(--color-surface);
    border-right: 1px solid var(--color-border);
    height: 100%;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
  }

  .sidebar-header {
    padding: var(--space-3) var(--space-4);
    font-size: 0.75rem;
    font-weight: 600;
    color: var(--color-text-muted);
    text-transform: uppercase;
    letter-spacing: 0.05em;
    border-bottom: 1px solid var(--color-border);
  }

  .agent-group {
    padding: var(--space-2) 0;
  }

  .group-header {
    padding: var(--space-2) var(--space-4);
    font-size: 0.7rem;
    font-weight: 600;
    color: var(--color-text-muted);
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .agent-item {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    width: 100%;
    padding: var(--space-2) var(--space-4);
    background: none;
    border: none;
    text-align: left;
    cursor: pointer;
    transition: background-color var(--transition-fast);
    color: var(--color-text);
  }

  .agent-item:hover {
    background-color: var(--color-surface-raised);
  }

  .agent-item.selected {
    background-color: var(--color-primary);
    color: white;
  }

  .agent-item.selected .status-dot.active {
    background-color: #86efac;
  }

  .status-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background-color: var(--color-text-muted);
    flex-shrink: 0;
  }

  .status-dot.active {
    background-color: var(--color-success);
  }

  .agent-icon {
    font-size: 0.875rem;
    flex-shrink: 0;
  }

  .agent-name {
    flex: 1;
    font-size: 0.875rem;
    font-weight: 500;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .work-indicator {
    font-size: 0.75rem;
    flex-shrink: 0;
  }

  .mail-badge {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-width: 1rem;
    height: 1rem;
    padding: 0 var(--space-1);
    font-size: 0.625rem;
    font-weight: 600;
    background-color: var(--color-error);
    color: white;
    border-radius: 9999px;
    flex-shrink: 0;
  }

  .agent-item.selected .mail-badge {
    background-color: rgba(255, 255, 255, 0.3);
  }
</style>
