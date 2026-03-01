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

<div class="my-1 rounded-lg border border-surface0 overflow-hidden flex-1 mr-4">
	<table class="w-full text-sm">
		{#if headerRows.length > 0}
			<thead class="bg-surface0">
				{#each headerRows as row}
					<tr>
						{#each row.markdownOpts?.table?.cells ?? [] as cell, colIdx}
							<th
								class="px-4 py-2 text-left font-bold"
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
			<tbody class="divide-y divide-surface0">
				{#each dataRows as row, rowIdx}
					<tr class={rowIdx % 2 === 1 ? 'bg-mantle' : 'bg-base'}>
						{#each row.markdownOpts?.table?.cells ?? [] as cell, colIdx}
							<td
								class="px-4 py-2"
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
</div>
