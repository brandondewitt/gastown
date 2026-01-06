<script lang="ts">
  import { onMount } from 'svelte';

  interface Rig {
    name: string;
    path: string;
    polecats: string[];
    polecat_count: number;
    crews: string[];
    crew_count: number;
    has_witness: boolean;
    has_refinery: boolean;
  }

  interface Props {
    isOpen?: boolean;
  }

  // Props
  let { isOpen = $bindable(false) }: Props = $props();

  // Form state
  let issueId = $state('');
  let selectedRig = $state('');
  let rigs: Rig[] = $state([]);
  let loading = $state(false);
  let error: string | null = $state(null);
  let success = $state(false);
  let rigsLoading = $state(false);

  // Fetch rigs on mount
  onMount(async () => {
    await fetchRigs();
  });

  async function fetchRigs() {
    try {
      rigsLoading = true;
      const response = await fetch('/api/v1/rigs');
      const data = await response.json();
      if (data.success && Array.isArray(data.data)) {
        rigs = data.data;
        if (rigs.length > 0) {
          selectedRig = rigs[0].name;
        }
      } else {
        error = data.error?.message || 'Failed to load rigs';
      }
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load rigs';
    } finally {
      rigsLoading = false;
    }
  }

  async function handleSubmit(e: Event) {
    e.preventDefault();

    // Validate inputs
    if (!issueId.trim()) {
      error = 'Issue ID is required';
      return;
    }
    if (!selectedRig) {
      error = 'Rig selection is required';
      return;
    }

    try {
      loading = true;
      error = null;
      success = false;

      const response = await fetch('/api/v1/sling', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          issue_id: issueId.trim(),
          target: selectedRig,
        }),
      });

      const data = await response.json();

      if (data.success) {
        success = true;
        // Reset form after short delay
        setTimeout(() => {
          closeDialog();
        }, 500);
      } else {
        error = data.error?.message || 'Failed to sling work';
      }
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to sling work';
    } finally {
      loading = false;
    }
  }

  function closeDialog() {
    isOpen = false;
    issueId = '';
    selectedRig = rigs.length > 0 ? rigs[0].name : '';
    error = null;
    success = false;
  }

  function handleBackdropClick(e: MouseEvent) {
    if (e.target === e.currentTarget) {
      closeDialog();
    }
  }
</script>

{#if isOpen}
  <div class="dialog-backdrop" role="presentation" onclick={handleBackdropClick}>
    <div class="dialog">
      <div class="dialog-header">
        <h2 class="dialog-title">Dispatch Work</h2>
        <button class="close-btn" onclick={closeDialog} aria-label="Close dialog">
          ✕
        </button>
      </div>

      <form onsubmit={handleSubmit} class="dialog-form">
        {#if success}
          <div class="success-message">
            <span class="success-icon">✓</span>
            <p>Work dispatched successfully to {selectedRig}</p>
          </div>
        {/if}

        {#if error}
          <div class="error-message">
            <span class="error-icon">!</span>
            <p>{error}</p>
          </div>
        {/if}

        {#if !success}
          <div class="form-group">
            <label for="issue-id" class="form-label">Issue ID</label>
            <input
              id="issue-id"
              type="text"
              class="form-input"
              placeholder="e.g., gt-123, hq-456"
              bind:value={issueId}
              disabled={loading}
              required
            />
            <p class="form-hint">The ID of the issue to dispatch</p>
          </div>

          <div class="form-group">
            <label for="rig-select" class="form-label">Rig</label>
            {#if rigsLoading}
              <div class="spinner-inline"></div>
              <p class="form-hint">Loading rigs...</p>
            {:else}
              <select
                id="rig-select"
                class="form-select"
                bind:value={selectedRig}
                disabled={loading || rigs.length === 0}
              >
                {#each rigs as rig}
                  <option value={rig.name}>{rig.name}</option>
                {/each}
              </select>
              {#if rigs.length === 0}
                <p class="form-hint error">No rigs available</p>
              {:else}
                <p class="form-hint">Dispatch to which rig</p>
              {/if}
            {/if}
          </div>

          <div class="form-actions">
            <button
              type="button"
              class="btn btn-secondary"
              onclick={closeDialog}
              disabled={loading}
            >
              Cancel
            </button>
            <button
              type="submit"
              class="btn btn-primary"
              disabled={loading || rigs.length === 0}
            >
              {#if loading}
                <span class="spinner-inline-small"></span>
                Dispatching...
              {:else}
                Dispatch
              {/if}
            </button>
          </div>
        {/if}
      </form>
    </div>
  </div>
{/if}

<style>
  .dialog-backdrop {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
    animation: fadeIn var(--transition-fast);
  }

  @keyframes fadeIn {
    from {
      opacity: 0;
    }
    to {
      opacity: 1;
    }
  }

  .dialog {
    background-color: var(--color-surface);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-md);
    box-shadow: 0 10px 40px rgba(0, 0, 0, 0.3);
    width: 90%;
    max-width: 500px;
    animation: slideUp var(--transition-normal);
  }

  @keyframes slideUp {
    from {
      transform: translateY(20px);
      opacity: 0;
    }
    to {
      transform: translateY(0);
      opacity: 1;
    }
  }

  .dialog-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: var(--space-6);
    border-bottom: 1px solid var(--color-border);
  }

  .dialog-title {
    font-size: 1.25rem;
    font-weight: 600;
    margin: 0;
  }

  .close-btn {
    background: none;
    border: none;
    color: var(--color-text-muted);
    font-size: 1.5rem;
    cursor: pointer;
    padding: 0;
    width: 2rem;
    height: 2rem;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: var(--radius-sm);
    transition: background-color var(--transition-fast);
  }

  .close-btn:hover {
    background-color: var(--color-surface-raised);
    color: var(--color-text);
  }

  .dialog-form {
    padding: var(--space-6);
  }

  .form-group {
    margin-bottom: var(--space-4);
  }

  .form-label {
    display: block;
    font-weight: 500;
    margin-bottom: var(--space-2);
    color: var(--color-text);
  }

  .form-input,
  .form-select {
    width: 100%;
    padding: var(--space-2) var(--space-3);
    background-color: var(--color-surface-raised);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-sm);
    color: var(--color-text);
    font-family: var(--font-sans);
    font-size: 1rem;
    transition: border-color var(--transition-fast), background-color var(--transition-fast);
  }

  .form-input:focus,
  .form-select:focus {
    outline: none;
    border-color: var(--color-primary);
    background-color: var(--color-surface);
  }

  .form-input:disabled,
  .form-select:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .form-hint {
    font-size: 0.875rem;
    color: var(--color-text-muted);
    margin-top: var(--space-1);
  }

  .form-hint.error {
    color: var(--color-error);
  }

  .form-actions {
    display: flex;
    gap: var(--space-3);
    justify-content: flex-end;
    margin-top: var(--space-6);
  }

  .btn {
    padding: var(--space-2) var(--space-4);
    border: none;
    border-radius: var(--radius-sm);
    font-weight: 500;
    cursor: pointer;
    transition: background-color var(--transition-fast), opacity var(--transition-fast);
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    font-size: 0.875rem;
  }

  .btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .btn-primary {
    background-color: var(--color-primary);
    color: white;
  }

  .btn-primary:hover:not(:disabled) {
    background-color: #2563eb;
  }

  .btn-secondary {
    background-color: var(--color-surface-raised);
    color: var(--color-text);
  }

  .btn-secondary:hover:not(:disabled) {
    background-color: var(--color-border);
  }

  .success-message {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    padding: var(--space-3) var(--space-4);
    background-color: rgba(34, 197, 94, 0.1);
    border: 1px solid rgba(34, 197, 94, 0.3);
    border-radius: var(--radius-sm);
    color: var(--color-success);
    margin-bottom: var(--space-4);
  }

  .success-icon {
    font-size: 1.25rem;
    font-weight: bold;
  }

  .error-message {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    padding: var(--space-3) var(--space-4);
    background-color: rgba(239, 68, 68, 0.1);
    border: 1px solid rgba(239, 68, 68, 0.3);
    border-radius: var(--radius-sm);
    color: var(--color-error);
    margin-bottom: var(--space-4);
  }

  .error-icon {
    font-size: 1.25rem;
    font-weight: bold;
  }

  .spinner-inline {
    display: inline-block;
    width: 1rem;
    height: 1rem;
    border: 2px solid var(--color-border);
    border-top-color: var(--color-primary);
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
  }

  .spinner-inline-small {
    display: inline-block;
    width: 0.75rem;
    height: 0.75rem;
    border: 2px solid rgba(255, 255, 255, 0.3);
    border-top-color: white;
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }
</style>
