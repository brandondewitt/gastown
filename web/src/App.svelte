<script lang="ts">
  import { onMount } from 'svelte';
  import './app.css';
  import Sidebar from './lib/Sidebar.svelte';
  import ChatInterface from './lib/ChatInterface.svelte';
  import Dashboard from './lib/Dashboard.svelte';

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
  let currentView: 'chat' | 'dashboard' = 'chat';
  let selectedAgent: Agent | null = null;

  async function fetchStatus() {
    try {
      const response = await fetch('/api/v1/status');
      const data = await response.json();
      if (data.success) {
        status = data.data;
        error = null;

        // Auto-select Mayor if not already selected
        if (!selectedAgent && status && status.agents?.length > 0) {
          const mayor = status.agents.find(a => a.role === 'mayor');
          if (mayor) {
            selectedAgent = mayor;
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
    const interval = setInterval(fetchStatus, 5000);
    return () => clearInterval(interval);
  });

  function handleSelectAgent(agent: Agent) {
    selectedAgent = agent;
    currentView = 'chat';
  }

  function handleViewChange(view: 'chat' | 'dashboard') {
    currentView = view;
  }
</script>

<div class="app">
  {#if status}
    <Sidebar
      {status}
      {selectedAgent}
      {currentView}
      on:selectAgent={(e) => handleSelectAgent(e.detail)}
      on:viewChange={(e) => handleViewChange(e.detail)}
    />
  {/if}

  <div class="main-container">
    {#if loading}
      <div class="loading-screen">
        <div class="spinner"></div>
        <p>⛽ Initializing Gas Town...</p>
      </div>
    {:else if error}
      <div class="error-screen">
        <p class="text-error">Error: {error}</p>
        <button on:click={fetchStatus}>Reconnect</button>
      </div>
    {:else if status}
      {#if currentView === 'chat'}
        <ChatInterface agent={selectedAgent} />
      {:else}
        <Dashboard {status} />
      {/if}
    {/if}
  </div>
</div>

<style>
  .app {
    display: flex;
    height: 100vh;
    background-color: var(--color-bg);
    overflow: hidden;
  }

  .main-container {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .loading-screen {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: var(--space-4);
    padding: var(--space-8);
    height: 100%;
  }

  .error-screen {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: var(--space-4);
    padding: var(--space-8);
    height: 100%;
  }

  .error-screen button {
    padding: var(--space-2) var(--space-4);
    background-color: var(--color-primary);
    color: white;
    border: none;
    border-radius: var(--radius-md);
    cursor: pointer;
    font-family: var(--font-mono);
  }

  .error-screen button:hover {
    opacity: 0.9;
  }
</style>
