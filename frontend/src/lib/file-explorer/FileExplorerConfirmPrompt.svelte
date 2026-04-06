<script lang="ts">
	import { Call } from '@wailsio/runtime';
	import { getFileExplorerConfirmPrompt } from '$lib/state.svelte';

	let prompt = $derived(getFileExplorerConfirmPrompt());

	function choose(index: number) {
		void Call.ByName('main.App.OnFileExplorerConfirmChoice', index).catch((err) => {
			console.error('failed to send file explorer confirm choice', err);
		});
	}
</script>

{#if prompt}
	<div class="confirm-overlay pointer-events-auto absolute inset-0 z-[220] flex items-center justify-center">
		<div class="confirm-dialog bg-base-100 text-text border border-surface0 rounded-md shadow-xl w-[42rem] max-w-[92vw]">
			<div class="p-5 whitespace-pre-wrap leading-relaxed">{prompt.message}</div>
			<div class="border-t border-surface0 px-5 py-4 flex items-center justify-end gap-3">
				{#each prompt.choices as choice, idx (idx)}
					<button
						type="button"
						class="btn btn-sm bg-surface0 hover:bg-surface1 text-text border border-surface1"
						onclick={() => choose(idx + 1)}
					>
						{choice}
					</button>
				{/each}
			</div>
		</div>
	</div>
{/if}
