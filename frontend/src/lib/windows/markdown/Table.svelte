<script lang="ts">
	import MarkdownToken from './MarkdownToken.svelte';
	let {
		rows,
		decode,
		cursorRow
	}: {
		rows: App.NvimRow[];
		decode?: (input: string) => string;
		cursorRow?: number;
	} = $props();

	let headerRows = $derived(rows.filter((r) => r.markdownOpts?.table?.rowType === 'header'));
	let dataRows = $derived(rows.filter((r) => r.markdownOpts?.table?.rowType === 'data'));
	let alignments = $derived(rows[0]?.markdownOpts?.table?.alignments ?? []);

	let cursorInside = $derived(
		cursorRow != null &&
			rows.length > 0 &&
			cursorRow >= rows[0].index &&
			cursorRow <= rows[rows.length - 1].index
	);
</script>

<div class="my-1 rounded-lg border border-surface0 overflow-hidden flex-1 mx-4 mr-4">
	{#if cursorInside}
		<table class="w-full text-sm">
			<tbody class="divide-y divide-surface0">
				{#each rows as row (row.index)}
					{#if row.markdownOpts?.table?.cells?.length}
						<tr>
							<td class="px-4 py-2 whitespace-pre" colspan={alignments.length || 1}>
								{#each row.tokens as token, i (i)}
									<MarkdownToken {token} {decode} />
								{/each}
							</td>
						</tr>
					{/if}
				{/each}
			</tbody>
		</table>
	{:else}
		<table class="w-full text-sm">
			{#if headerRows.length > 0}
				<thead class="bg-surface0">
					{#each headerRows as row (row.index)}
						<tr>
							{#each row.markdownOpts?.table?.cells ?? [] as cell, colIdx (colIdx)}
								<th
									class="px-4 py-2 text-left font-bold"
									style:text-align={alignments[colIdx] ?? 'left'}
								>
									{#each cell as token, i (i)}
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
					{#each dataRows as row, rowIdx (row.index)}
						<tr class={rowIdx % 2 === 1 ? 'bg-mantle' : 'bg-base'}>
							{#each row.markdownOpts?.table?.cells ?? [] as cell, colIdx (colIdx)}
								<td
									class="px-4 py-2"
									style:text-align={alignments[colIdx] ?? 'left'}
								>
									{#each cell as token, i (i)}
										<MarkdownToken {token} {decode} />
									{/each}
								</td>
							{/each}
						</tr>
					{/each}
				</tbody>
			{/if}
		</table>
	{/if}
</div>
