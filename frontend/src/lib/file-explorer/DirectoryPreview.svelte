<script lang="ts">
	interface Props {
		title: string;
		directory: App.FileExplorerDirectory;
	}

	let { title, directory }: Props = $props();
</script>

<div class="bg-very-dark flex h-full flex-col overflow-hidden">
	<div class="border-mantle text-surface0 flex shrink-0 items-center gap-3 border-b-2 px-6 py-2 text-xl">
		<span class="text-blue">󰉋</span>
		<span class="overflow-hidden text-ellipsis whitespace-nowrap">{title}</span>
	</div>
	<div class="min-h-0 flex-1 overflow-y-auto py-1">
		{#each directory.entries as entry (entry.id)}
			<div class="flex h-9 cursor-default items-center gap-1 whitespace-nowrap px-4 py-1">
				<span
					class:text-peach={entry.isDir}
					class={`grid h-[2ch] w-[2ch] shrink-0 place-items-center rounded-full ${entry.iconClass ?? ''}`}
					>{entry.icon}</span
				>
				<span class="overflow-hidden text-ellipsis">{entry.text}</span>
			</div>
		{/each}
	</div>
</div>
