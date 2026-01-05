<script lang="ts">
  interface Rig {
    name: string;
    has_witness: boolean;
    has_refinery: boolean;
    polecat_count: number;
  }

  interface ActionState {
    loading: boolean;
    error: string | null;
  }

  export let rig: Rig;
  export let onActionComplete: (() => void) | null = null;

  let states: Record<string, ActionState> = {
    witness_start: { loading: false, error: null },
    witness_stop: { loading: false, error: null },
    refinery_start: { loading: false, error: null },
    refinery_stop: { loading: false, error: null },
    polecat_add: { loading: false, error: null },
    polecat_remove: { loading: false, error: null },
  };

  async function performAction(service: string, action: string, actionKey: string) {
    states[actionKey].loading = true;
    states[actionKey].error = null;

    try {
      const response = await fetch(`/api/v1/rigs/${rig.name}/services/${service}/${action}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error?.message || `Failed to ${action} ${service}`);
      }

      states[actionKey].error = null;
      if (onActionComplete) {
        onActionComplete();
      }
    } catch (e) {
      states[actionKey].error = e instanceof Error ? e.message : 'Unknown error';
    } finally {
      states[actionKey].loading = false;
    }
  }

  function startWitness() {
    performAction('witness', 'start', 'witness_start');
  }

  function stopWitness() {
    performAction('witness', 'stop', 'witness_stop');
  }

  function startRefinery() {
    performAction('refinery', 'start', 'refinery_start');
  }

  function stopRefinery() {
    performAction('refinery', 'stop', 'refinery_stop');
  }

  function addPolecat() {
    performAction('polecat', 'add', 'polecat_add');
  }

  function removePolecat() {
    performAction('polecat', 'remove', 'polecat_remove');
  }
</script>

<div class="action-panel">
  <h3 class="panel-title">Services</h3>

  <!-- Witness Actions -->
  <div class="action-group">
    <div class="action-header">
      <span class="action-icon">👀</span>
      <span class="action-label">Witness</span>
      <span class="status-badge" class:active={rig.has_witness}>
        {rig.has_witness ? 'Running' : 'Stopped'}
      </span>
    </div>
    <div class="button-group">
      <button
        on:click={startWitness}
        disabled={states.witness_start.loading || rig.has_witness}
        class="btn btn-success"
      >
        {states.witness_start.loading ? 'Starting...' : 'Start'}
      </button>
      <button
        on:click={stopWitness}
        disabled={states.witness_stop.loading || !rig.has_witness}
        class="btn btn-error"
      >
        {states.witness_stop.loading ? 'Stopping...' : 'Stop'}
      </button>
    </div>
    {#if states.witness_start.error}
      <div class="error-message">{states.witness_start.error}</div>
    {/if}
    {#if states.witness_stop.error}
      <div class="error-message">{states.witness_stop.error}</div>
    {/if}
  </div>

  <!-- Refinery Actions -->
  <div class="action-group">
    <div class="action-header">
      <span class="action-icon">⚙️</span>
      <span class="action-label">Refinery</span>
      <span class="status-badge" class:active={rig.has_refinery}>
        {rig.has_refinery ? 'Running' : 'Stopped'}
      </span>
    </div>
    <div class="button-group">
      <button
        on:click={startRefinery}
        disabled={states.refinery_start.loading || rig.has_refinery}
        class="btn btn-success"
      >
        {states.refinery_start.loading ? 'Starting...' : 'Start'}
      </button>
      <button
        on:click={stopRefinery}
        disabled={states.refinery_stop.loading || !rig.has_refinery}
        class="btn btn-error"
      >
        {states.refinery_stop.loading ? 'Stopping...' : 'Stop'}
      </button>
    </div>
    {#if states.refinery_start.error}
      <div class="error-message">{states.refinery_start.error}</div>
    {/if}
    {#if states.refinery_stop.error}
      <div class="error-message">{states.refinery_stop.error}</div>
    {/if}
  </div>

  <!-- Polecat Actions -->
  <div class="action-group">
    <div class="action-header">
      <span class="action-icon">🐱</span>
      <span class="action-label">Polecats</span>
      <span class="count-badge">{rig.polecat_count}</span>
    </div>
    <div class="button-group">
      <button
        on:click={addPolecat}
        disabled={states.polecat_add.loading}
        class="btn btn-primary"
      >
        {states.polecat_add.loading ? 'Adding...' : 'Add'}
      </button>
      <button
        on:click={removePolecat}
        disabled={states.polecat_remove.loading || rig.polecat_count === 0}
        class="btn btn-warning"
      >
        {states.polecat_remove.loading ? 'Removing...' : 'Remove'}
      </button>
    </div>
    {#if states.polecat_add.error}
      <div class="error-message">{states.polecat_add.error}</div>
    {/if}
    {#if states.polecat_remove.error}
      <div class="error-message">{states.polecat_remove.error}</div>
    {/if}
  </div>
</div>

<style>
  .action-panel {
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
  }

  .panel-title {
    font-size: 0.875rem;
    font-weight: 500;
    color: var(--color-text-muted);
    text-transform: uppercase;
    letter-spacing: 0.05em;
    margin: 0;
  }

  .action-group {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    padding: var(--space-3);
    background-color: var(--color-surface-raised);
    border-radius: var(--radius-md);
  }

  .action-header {
    display: flex;
    align-items: center;
    gap: var(--space-2);
  }

  .action-icon {
    font-size: 1.25rem;
  }

  .action-label {
    font-weight: 500;
    flex: 1;
  }

  .status-badge {
    display: inline-flex;
    align-items: center;
    padding: var(--space-1) var(--space-2);
    font-size: 0.75rem;
    font-weight: 500;
    border-radius: var(--radius-sm);
    background-color: var(--color-surface);
    color: var(--color-text-muted);
  }

  .status-badge.active {
    background-color: rgba(34, 197, 94, 0.15);
    color: var(--color-success);
  }

  .count-badge {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-width: 1.5rem;
    height: 1.5rem;
    padding: 0 var(--space-1);
    font-size: 0.75rem;
    font-weight: 600;
    background-color: var(--color-primary);
    color: white;
    border-radius: 9999px;
  }

  .button-group {
    display: flex;
    gap: var(--space-2);
  }

  .btn {
    flex: 1;
    padding: var(--space-2) var(--space-3);
    font-size: 0.875rem;
    font-weight: 500;
    border: none;
    border-radius: var(--radius-md);
    cursor: pointer;
    transition: all var(--transition-fast);
  }

  .btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .btn-primary {
    background-color: var(--color-primary);
    color: white;
  }

  .btn-primary:not(:disabled):hover {
    background-color: #2563eb;
  }

  .btn-success {
    background-color: var(--color-success);
    color: white;
  }

  .btn-success:not(:disabled):hover {
    background-color: #16a34a;
  }

  .btn-warning {
    background-color: var(--color-warning);
    color: #000;
  }

  .btn-warning:not(:disabled):hover {
    background-color: #ca8a04;
  }

  .btn-error {
    background-color: var(--color-error);
    color: white;
  }

  .btn-error:not(:disabled):hover {
    background-color: #dc2626;
  }

  .error-message {
    padding: var(--space-2) var(--space-3);
    font-size: 0.75rem;
    background-color: rgba(239, 68, 68, 0.15);
    color: var(--color-error);
    border-radius: var(--radius-md);
  }
</style>
