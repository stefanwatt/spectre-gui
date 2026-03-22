<script lang="ts">
	let {
		mode,
		cursor,
		filepath
	}: { mode?: App.VimMode; cursor: { row: number; col: number }; filepath?: string } = $props();
	const modeNames = new Map();
	modeNames.set('normal', 'normal');
	modeNames.set('insert', 'insert');
	modeNames.set('visual', 'visual');
	modeNames.set('V', 'v-line');
	modeNames.set('cmdline_normal', 'command');
	modeNames.set('cmdline_insert', 'command');
	let modeName = $derived(modeNames.get(mode||'normal'));
</script>

<div class="flex h-full w-full items-center bg-crust p-1 text-text">
	<span
		class:bg-blue={mode === 'normal'}
		class:bg-green={mode === 'insert'}
		class:bg-mauve={mode === 'visual'}
		class:bg-peach={mode === 'cmdline_normal'}
		class="rounded-md px-2 text-xl font-bold uppercase text-mantle"
	>
		{modeName}
	</span>
	<span class="mx-2">
		{filepath}
	</span>
	<span class="flex-grow"> </span>
	<span
		class="flex w-24 items-center justify-center rounded-md bg-blue px-2 text-xl font-bold uppercase text-mantle"
	>
		{cursor.row}:{cursor.col + 1}
	</span>
</div>
