<script lang="ts">
	interface Props {
		data?: App.FileExplorerDirectory;
		role: 'parent' | 'current';
	}

	let { data, role }: Props = $props();

	let isCurrentPane = $derived(role === 'current');
	let currentColor = $derived(data?.dirty ? 'yellow' : 'mauve');

	function splitEntryTextAtCursor(entry: App.FileExplorerEntry, cursorCol: number) {
		const cursorDisplayCol = Math.min(Math.max(0, cursorCol), entry.text.length);
		const hasCharUnderCursor = cursorDisplayCol < entry.text.length;
		return {
			before: entry.text.slice(0, cursorDisplayCol),
			cursor: hasCharUnderCursor
				? entry.text.slice(cursorDisplayCol, cursorDisplayCol + 1)
				: '\u00a0',
			after: hasCharUnderCursor ? entry.text.slice(cursorDisplayCol + 1) : '',
			atEnd: !hasCharUnderCursor
		};
	}

	function scrollSelectedIntoView(node: HTMLElement, isSelected: boolean) {
		if (isSelected) {
			node.scrollIntoView({ block: 'center' });
		}
		return {
			update(isSelected: boolean) {
				if (isSelected) {
					node.scrollIntoView({ block: 'center' });
				}
			}
		};
	}
</script>

{#if data}
	<div
		class="border-r-mantle flex h-full flex-col overflow-hidden border-r-2"
		class:bg-very-dark={!isCurrentPane}
		class:bg-pretty-dark={isCurrentPane}
		class:focused={isCurrentPane}
		class:parent-dir={role === 'parent'}
		class:current-dir={isCurrentPane}
	>
		<div class="border-mantle relative flex shrink-0 flex-col border-b-2">
			{#if isCurrentPane}
				{#if !data.dirty}
					<span
						class="absolute inset-x-0 top-0 h-0.5 rounded-full bg-[linear-gradient(90deg,rgba(var(--ctp-mauve),1)_0%,rgba(var(--ctp-blue),0.35)_65%,rgba(var(--ctp-blue),0)_100%)] shadow-[0_0_8px_rgba(var(--ctp-mauve),0.8),0_0_16px_rgba(var(--ctp-mauve),0.35)]"
						aria-hidden="true"
					></span>
				{:else}
					<span
						class="absolute inset-x-0 top-0 h-0.5 rounded-full bg-[linear-gradient(90deg,rgba(var(--ctp-yellow),1)_0%,rgba(var(--ctp-green),0.35)_65%,rgba(var(--ctp-green),0)_100%)] shadow-[0_0_8px_rgba(var(--ctp-yellow),0.8),0_0_16px_rgba(var(--ctp-yellow),0.35)]"
						aria-hidden="true"
					></span>
				{/if}
			{/if}

			<div
				class:text-mauve={isCurrentPane && !data.dirty}
				class:text-yellow={isCurrentPane && data.dirty}
				class:text-surface0={!isCurrentPane}
				class="mx-2 flex items-center gap-4 px-4 py-2 text-xl"
			>
				{#if isCurrentPane}
					<span
						class="inline-block [filter:drop-shadow(0_0_6px_rgba(var(--ctp-{currentColor}),0.8))] [text-shadow:0_0_2px_rgba(var(--ctp-{currentColor}),0.95),0_0_5px_rgba(var(--ctp-{currentColor}),0.6)]"
						>
					</span>
				{/if}
				<span>{data.title}</span>
			</div>
		</div>
		<div class="min-h-0 flex-1 overflow-y-auto">
			{#each data.entries as entry (entry.id)}
				{@const selected = data.selectedEntryId === entry.id}
				{@const cursorActive = isCurrentPane && selected}
				{@const split = cursorActive ? splitEntryTextAtCursor(entry, data.cursorCol ?? 0) : null}
				<div
					use:scrollSelectedIntoView={selected}
					class="flex h-9 cursor-default items-center gap-1 whitespace-nowrap px-4 py-1"
					class:bg-surface0={selected}
				>
					<span
						class:text-peach={entry.isDir}
						class={`grid h-[2ch] w-[2ch] shrink-0 place-items-center rounded-full ${entry.iconClass ?? ''}`}
						>{entry.icon}</span
					>
					<span class="flex h-full overflow-hidden text-ellipsis">
						{#if cursorActive && split}
							<span class="grid h-full place-items-center">{split.before}</span>
							<span class="grid h-full min-w-[1ch] place-items-center" class:cursor={true}
								>{split.cursor}</span
							><span class="grid h-full place-items-center">{split.after}</span>
						{:else}
							<span class="grid h-full place-items-center">{entry.text}</span>
						{/if}
					</span>
				</div>
			{/each}
		</div>
	</div>
{/if}
