<script>
	import { SendKey } from '$lib/wailsjs/go/main/App';
	import { onDestroy, onMount } from 'svelte';
	import StatusLine from './StatusLine.svelte';
	import { Grid } from 'svelte-virtual';

	/**@type{App.VimMode}*/
	let mode = 'n';

	/**@type{App.BufLine[]}*/
	let lines = [];

	/**@type{App.NvimRange}*/
	let selection_range = {
		start_row: -1,
		end_row: -1,
		start_col: 0,
		end_col: 0
	};

	let cursor = { row: 5, col: 5, key: ' ' };

	function send_key(e) {
		SendKey(e.key, e.ctrlKey, e.shiftKey, e.altKey);
	}

	/**@param {App.CursorMoveEvent} e*/
	function on_cursor_moved(e) {
		cursor = { ...cursor, row: e.row, col: e.col, key: e.key };
		lines = lines;
		scroll_into_view(e.top_line);
	}

	onMount(async () => {
		window.addEventListener('keydown', send_key);
		const runtime = await import('$lib/wailsjs/runtime/runtime');
		runtime.EventsOn('buf-lines-changed', (updated_lines) => {
			lines = updated_lines;
			console.log('updated line:', updated_lines);
		});

		runtime.EventsOn('cursor-changed', on_cursor_moved);
		runtime.EventsOn('mode-changed', (new_mode) => {
			mode = new_mode;
		});
		runtime.EventsOn('visual-selection-changed', (new_selection_range) => {
			selection_range = new_selection_range;
			console.log('new_selection_range', new_selection_range);
		});
	});

	onDestroy(() => {
		window.removeEventListener('keydown', send_key);
	});

	/**@param {number} line*/
	function scroll_into_view(line) {
		const lineElement = document.querySelector(`.buf-line-${line - 1}`);
		if (lineElement) {
			lineElement.scrollIntoView();
		}
	}

	/**
	 * @param {number} row
	 * @param {number} col
	 * @returns {boolean}
	 */
	function is_in_selection_range(row, col) {
		let start_row = selection_range.start_row;
		let end_row = selection_range.end_row;
		let start_col = selection_range.start_col;
		let end_col = selection_range.end_col;

		if (end_row < start_row) {
			start_row = selection_range.end_row;
			end_row = selection_range.start_row;
			start_col = selection_range.end_col;
			end_col = selection_range.start_col;
		}

		if (mode === 'V') {
			return row >= start_row && row <= end_row;
		}

		if (start_row === end_row) {
			if (row === start_row) {
				return col >= start_col && col <= end_col;
			}
		} else {
			if (row === start_row) {
				return col >= start_col;
			}
			if (row === end_row) {
				return col <= end_col;
			}
		}
		return row >= start_row && row <= end_row;
	}
</script>

<div class="flex h-screen flex-col">
	<div class="h-full w-screen grow snap-y overflow-y-scroll whitespace-pre font-mono text-xl">
		<Grid
			itemCount={lines.length * 2}
			itemWidth={50}
			itemHeight={28}
			columnCount={2}
			height={1400}
		>
			<div slot="item" let:columnIndex let:rowIndex let:style {style}>
				{#if lines?.length >= rowIndex}
					{#if columnIndex === 0}
						<div class="buf-line-{rowIndex} text-overlay0">
							<span class="text-right">{lines[rowIndex].sign}</span>
							<span class="ml-1 text-right">{lines[rowIndex].row + 1}</span>
						</div>
					{:else}
						<div
							class:bg-surface0={mode !== 'v' && mode !== 'V' && cursor.row === lines[rowIndex].row}
							class="victor-mono relative ml-4 flex snap-start whitespace-pre"
						>
							{#each lines[rowIndex]?.tokens || [] as token}
								<div
									class:strikethrough={token.strikethrough}
									class:underline={token.underline}
									class:italic={token.italic}
									style="color:{token.foreground}; background:{token.background}"
									class="flex"
								>
									{#each token.text || '' as cell, i}
										{#if cursor.row === lines[rowIndex].row && cursor.col === token.start_col + i}
											{#if cell === '\t'}
												<span
													class:bg-surface2={is_in_selection_range(
														lines[rowIndex].row,
														i + token.start_col
													)}
													class="h-full !bg-rosewater !text-mantle"
												>
													{' '}
												</span>
												<span
													class:bg-surface2={is_in_selection_range(
														lines[rowIndex].row,
														i + token.start_col
													)}
													class="whitespace-pre">{'   '}</span
												>
											{:else}
												<span
													class:bg-surface2={is_in_selection_range(
														lines[rowIndex].row,
														i + token.start_col
													)}
													class="h-full !bg-rosewater !text-mantle"
												>
													{cell}
												</span>
											{/if}
										{:else}
											<span
												class:bg-surface2={is_in_selection_range(
													lines[rowIndex].row,
													i + token.start_col
												)}>{cell}</span
											>
										{/if}
									{/each}
								</div>
							{/each}
						</div>
					{/if}
				{/if}
			</div>
		</Grid>
	</div>
	<StatusLine {mode}></StatusLine>
</div>

<style>
	.victor-mono {
		font-family: VictorMono Nerd Font Mono;
	}
</style>
