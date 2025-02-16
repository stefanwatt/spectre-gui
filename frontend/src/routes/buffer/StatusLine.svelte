<script>
	/**@type{App.VimMode}*/
	export let mode = 'n';
	export let cursor ={ row: 0, col: 0 }
	const mode_names = new Map();
	mode_names.set('n', 'normal');
	mode_names.set('i', 'insert');
	mode_names.set('v', 'visual');
	mode_names.set('V', 'v-line');
	mode_names.set('c', 'command');
	$: mode_name = mode_names.get(mode);
</script>

<div class="flex h-10 w-full items-center bg-mantle p-1 text-text">
	<span
		class:bg-blue={mode === 'n'}
		class:bg-green={mode === 'i'}
		class:bg-mauve={mode === 'v' || mode === 'V'}
		class:bg-peach={mode === 'c'}
		class="rounded-md px-2 text-xl font-bold uppercase text-mantle"
	>
		{mode_name}
	</span>
	<span class="flex-grow">
	</span>
	<span 
		class="w-24 flex items-center justify-center rounded-md bg-blue px-2 text-xl font-bold uppercase text-mantle"
	>
		{cursor.row+1}:{cursor.col+1}
	</span>
</div>
