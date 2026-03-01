<script lang="ts">
	import MarkdownToken from './MarkdownToken.svelte';
	let {
		rows,
		decode
	}: {
		rows: App.NvimRow[];
		decode?: (input: string) => string;
	} = $props();

	let headerRows = $derived(rows.filter((r) => r.markdownOpts?.table?.rowType === 'header'));
	let dataRows = $derived(rows.filter((r) => r.markdownOpts?.table?.rowType === 'data'));
	let alignments = $derived(rows[0]?.markdownOpts?.table?.alignments ?? []);
</script>

<table class="border-collapse my-1 w-full">
	{#if headerRows.length > 0}
		<thead>
			{#each headerRows as row}
				<tr class="border-b-2 border-overlay2">
					{#each row.markdownOpts?.table?.cells ?? [] as cell, colIdx}
						<th
							class="px-3 py-1.5 text-left font-bold"
							style:text-align={alignments[colIdx] ?? 'left'}
						>
							{#each cell as token}
								<MarkdownToken {token} {decode} />
							{/each}
						</th>
					{/each}
				</tr>
			{/each}
		</thead>
	{/if}
	{#if dataRows.length > 0}
		<tbody>
			{#each dataRows as row, rowIdx}
				<tr class="border-b border-surface1 {rowIdx % 2 === 1 ? 'bg-mantle' : ''}">
					{#each row.markdownOpts?.table?.cells ?? [] as cell, colIdx}
						<td
							class="px-3 py-1.5"
							style:text-align={alignments[colIdx] ?? 'left'}
						>
							{#each cell as token}
								<MarkdownToken {token} {decode} />
							{/each}
						</td>
					{/each}
				</tr>
			{/each}
		</tbody>
	{/if}
</table>
