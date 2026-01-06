<script lang="ts">
  import { onMount, onDestroy } from 'svelte';

  let { agentAddress, autoScroll = true } = $props<{
    agentAddress: string;
    autoScroll?: boolean;
  }>();

  let output = $state('');
  let loading = $state(true);
  let error = $state<string | null>(null);
  let intervalId: number | undefined;
  let terminalEl: HTMLDivElement;

  async function fetchOutput() {
    try {
      const res = await fetch(`/api/v1/agents/${agentAddress}/output`);
      const data = await res.json();

      if (data.error) {
        error = data.error;
        output = '';
      } else {
        error = null;
        output = data.output || '';

        // Auto-scroll to bottom when new content arrives
        if (autoScroll && terminalEl) {
          requestAnimationFrame(() => {
            terminalEl.scrollTop = terminalEl.scrollHeight;
          });
        }
      }
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to fetch output';
    } finally {
      loading = false;
    }
  }

  onMount(() => {
    fetchOutput();
    // Poll for updates every second
    intervalId = setInterval(fetchOutput, 1000);
  });

  onDestroy(() => {
    if (intervalId) {
      clearInterval(intervalId);
    }
  });

  // Re-fetch when agent address changes
  $effect(() => {
    if (agentAddress) {
      loading = true;
      error = null;
      output = '';
      fetchOutput();
    }
  });
</script>

<div class="terminal-output" bind:this={terminalEl}>
  {#if loading && !output}
    <div class="terminal-loading">
      <span class="loading-text">Connecting to terminal...</span>
    </div>
  {:else if error}
    <div class="terminal-error">
      <span class="error-icon">⚠️</span>
      <span class="error-text">{error}</span>
    </div>
  {:else if !output}
    <div class="terminal-empty">
      <span class="empty-text">No terminal output available</span>
    </div>
  {:else}
    <pre class="terminal-content">{output}</pre>
  {/if}
</div>

<style>
  .terminal-output {
    background: #0d1117;
    color: #c9d1d9;
    font-family: 'SF Mono', 'Monaco', 'Inconsolata', 'Fira Mono', 'Droid Sans Mono', 'Source Code Pro', monospace;
    font-size: 0.8125rem;
    line-height: 1.5;
    padding: var(--space-3);
    border-radius: var(--radius-md);
    height: 400px;
    overflow-y: auto;
    overflow-x: hidden;
    border: 1px solid var(--color-border);
  }

  .terminal-content {
    margin: 0;
    white-space: pre-wrap;
    word-wrap: break-word;
    color: #c9d1d9;
  }

  .terminal-loading,
  .terminal-error,
  .terminal-empty {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 100%;
    gap: var(--space-2);
  }

  .loading-text {
    color: #8b949e;
    animation: pulse 1.5s ease-in-out infinite;
  }

  @keyframes pulse {
    0%, 100% { opacity: 0.5; }
    50% { opacity: 1; }
  }

  .terminal-error {
    flex-direction: column;
    color: #f85149;
  }

  .error-icon {
    font-size: 1.5rem;
  }

  .error-text {
    font-size: 0.875rem;
  }

  .terminal-empty {
    color: #8b949e;
    font-style: italic;
  }

  /* Scrollbar styling for dark terminal */
  .terminal-output::-webkit-scrollbar {
    width: 8px;
  }

  .terminal-output::-webkit-scrollbar-track {
    background: #21262d;
    border-radius: 4px;
  }

  .terminal-output::-webkit-scrollbar-thumb {
    background: #484f58;
    border-radius: 4px;
  }

  .terminal-output::-webkit-scrollbar-thumb:hover {
    background: #6e7681;
  }
</style>
